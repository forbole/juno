package main

import (
	"context"
	"flag"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/pkg/errors"
	"github.com/tendermint/tendermint/types"
)

var (
	configPath  string
	syncMissing bool

	wg sync.WaitGroup
)

func main() {
	flag.StringVar(&configPath, "config", "", "Configuration file")
	flag.BoolVar(&syncMissing, "sync-missing", false, "Sync missing or failed blocks from a previous export")
	flag.Parse()

	cfg := parseConfig(configPath)
	rpc, err := newRPCClient(cfg.Node)
	if err != nil {
		log.Fatal(errors.Wrap(err, "failed to start RPC client"))
	}

	defer rpc.node.Stop()

	db, err := openDB(cfg)
	if err != nil {
		log.Fatal(errors.Wrap(err, "failed to open database connection"))
	}

	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatal(errors.Wrap(err, "failed to ping database"))
	}

	// create a queue that will collect, aggregate, and export blocks and metadata
	exportQueue := newQueue(100)
	worker := newWorker(db, rpc, exportQueue)

	// Start a blocking worker in a go-routine where the worker consumes jobs off
	// of the export queue.
	log.Println("starting worker pool...")
	wg.Add(1)
	go worker.start()

	// listen for and trap any OS signal to gracefully shutdown and exit
	trapSignal()

	if syncMissing {
		// Sync any missing blocks from a previous export in the background. Start
		// in a go-routine as to not block further exporting.
		go enqueueSyncMissing(exportQueue, db)
	}

	enqueueSyncNew(exportQueue, db, rpc)
	go startBlockListener(exportQueue, rpc)

	// block main process (signal capture will call WaitGroup's Done)
	wg.Wait()
}

func parseConfig(configPath string) config {
	if configPath == "" {
		log.Fatal("invalid configuration file")
	}

	configData, err := ioutil.ReadFile(configPath)
	if err != nil {
		log.Fatal(errors.Wrap(err, "failed to read config"))
	}

	var cfg config
	if _, err := toml.Decode(string(configData), &cfg); err != nil {
		log.Fatal(errors.Wrap(err, "failed to decode config"))
	}

	return cfg
}

// enqueueSyncNew enqueues jobs (block heights) for new blocks starting at the
// latest stored block height.
func enqueueSyncNew(exportQueue queue, db *database, rpc rpcClient) {
	lastBlockHeight, err := db.lastBlockHeight()
	if err != nil {
		log.Fatal(errors.Wrap(err, "failed to get last block from database"))
	}

	latestBlockHeight, err := rpc.latestHeight()
	if err != nil {
		log.Fatal(errors.Wrap(err, "failed to get lastest block from RPC client"))
	}

	log.Println("syncing new blocks...")

	for i := lastBlockHeight + 1; i <= latestBlockHeight; i++ {
		if i == 1 {
			// skip the first block since it has no pre-commits (violates table constraints)
			continue
		}

		if i%10 == 0 {
			log.Printf("enqueueing new block %d\n", i)
		}

		exportQueue <- i
	}
}

// enqueueSyncMissing enqueues jobs (block heights) missing from a previous
// export in cases where the RPC client could fail.
//
// NOTE: We separate this from enqueueSyncNew to allow skipping scanning the
// entire database for any missing blocks. Since jobs are re-enqueued, this
// should be used seldomly.
func enqueueSyncMissing(exportQueue queue, db *database) {
	lastBlockHeight, err := db.lastBlockHeight()
	if err != nil {
		log.Fatal(errors.Wrap(err, "failed to get last block from database"))
	}

	log.Println("syncing missing blocks...")

	for i := int64(2); i < lastBlockHeight; i++ {
		ok, err := db.hasBlock(i)
		if !ok && err == nil {
			log.Printf("enqueueing missing block %d\n", i)
			exportQueue <- i
		}
	}
}

// startBlockListener subscribes to new block events via the Tendermint RPC and
// enqueues each new block height onto the provided queue. It blocks as new
// blocks are incoming.
func startBlockListener(exportQueue queue, rpc rpcClient) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	eventCh, err := rpc.node.Subscribe(ctx, "juno-client", "tm.event = 'NewBlock'")
	if err != nil {
		log.Fatal(errors.Wrap(err, "failed to subscribe to new blocks"))
	}

	log.Println("listening for new block events...")

	for e := range eventCh {
		newBlock := e.Data.(types.EventDataNewBlock).Block
		height := newBlock.Header.Height

		if height%10 == 0 {
			log.Printf("enqueueing new block %d\n", height)
		}

		exportQueue <- height
	}
}

// trapSignal will listen for any OS signal and invoke Done on the main
// WaitGroup allowing the main process to gracefully exit.
func trapSignal() {
	var sigCh = make(chan os.Signal)

	signal.Notify(sigCh, syscall.SIGTERM)
	signal.Notify(sigCh, syscall.SIGINT)

	go func() {
		sig := <-sigCh
		log.Printf("caught signal: %+v; shutting down...\n", sig)
		defer wg.Done()
	}()
}
