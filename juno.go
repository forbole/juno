package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/alexanderbez/juno/client"
	"github.com/alexanderbez/juno/config"
	"github.com/alexanderbez/juno/db"
	"github.com/pkg/errors"
	tmtypes "github.com/tendermint/tendermint/types"
)

var (
	configPath  string
	startHeight int64

	wg sync.WaitGroup
)

func main() {
	flag.StringVar(&configPath, "config", "", "Configuration file")
	flag.Int64Var(&startHeight, "start-height", 2, "Sync missing or failed blocks starting from a given height")
	flag.Parse()

	cfg := config.ParseConfig(configPath)
	cp, err := client.New(cfg.RPCNode, cfg.ClientNode)
	if err != nil {
		log.Fatal(errors.Wrap(err, "failed to start RPC client"))
	}

	defer cp.Stop() // nolint: errcheck

	db, err := db.OpenDB(cfg)
	if err != nil {
		log.Fatal(errors.Wrap(err, "failed to open database connection"))
	}

	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatal(errors.Wrap(err, "failed to ping database"))
	}

	// create a queue that will collect, aggregate, and export blocks and metadata
	exportQueue := newQueue(100)
	worker := newWorker(db, cp, exportQueue)

	// Start a blocking worker in a go-routine where the worker consumes jobs off
	// of the export queue.
	log.Println("starting worker pool...")
	wg.Add(1)
	go worker.start()

	// listen for and trap any OS signal to gracefully shutdown and exit
	trapSignal()

	go startNewBlockListener(exportQueue, cp)
	go enqueueMissingBlocks(exportQueue, cp)

	// block main process (signal capture will call WaitGroup's Done)
	wg.Wait()
}

// enqueueMissingBlocks enqueues jobs (block heights) for missed blocks starting
// at the startHeight up until the latest known height.
func enqueueMissingBlocks(exportQueue queue, cp client.ClientProxy) {
	latestBlockHeight, err := cp.LatestHeight()
	if err != nil {
		log.Fatal(errors.Wrap(err, "failed to get lastest block from RPC client"))
	}

	log.Println("syncing missing blocks...")

	for i := startHeight; i <= latestBlockHeight; i++ {
		if i == 1 {
			// skip the first block since it has no pre-commits (violates table constraints)
			continue
		}

		if i%10 == 0 {
			log.Printf("enqueueing missing block %d\n", i)
		}

		exportQueue <- i
	}
}

// startNewBlockListener subscribes to new block events via the Tendermint RPC
// and enqueues each new block height onto the provided queue. It blocks as new
// blocks are incoming.
func startNewBlockListener(exportQueue queue, cp client.ClientProxy) {
	eventCh, cancel, err := cp.SubscribeNewBlocks("juno-client")
	defer cancel()

	if err != nil {
		log.Fatal(errors.Wrap(err, "failed to subscribe to new blocks"))
	}

	log.Println("listening for new block events...")

	for e := range eventCh {
		newBlock := e.Data.(tmtypes.EventDataNewBlock).Block
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
