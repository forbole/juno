package cmd

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/alexanderbez/juno/client"
	"github.com/alexanderbez/juno/config"
	"github.com/alexanderbez/juno/db"
	"github.com/alexanderbez/juno/processor"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	tmtypes "github.com/tendermint/tendermint/types"
)

var (
	startHeight int64
	workerCount int16

	wg sync.WaitGroup
)

var rootCmd = &cobra.Command{
	Use:   "juno [config-file]",
	Args:  cobra.ExactArgs(1),
	Short: "Juno is a Cosmos Hub data aggregator and exporter",
	// Long: `A Fast and Flexible Static Site Generator built with
	//               love by spf13 and friends in Go.
	//               Complete documentation is available at http://hugo.spf13.com`,
	RunE: junoCmdHandler,
}

func init() {
	rootCmd.PersistentFlags().Int64Var(&startHeight, "start-height", 2, "sync missing or failed blocks starting from a given height")
	rootCmd.PersistentFlags().Int16Var(&workerCount, "workers", 1, "number of workers to run concurrently")
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func junoCmdHandler(cmd *cobra.Command, args []string) error {
	cfgFile := args[0]
	cfg := config.ParseConfig(cfgFile)

	cp, err := client.New(cfg.RPCNode, cfg.ClientNode)
	if err != nil {
		return errors.Wrap(err, "failed to start RPC client")
	}

	defer cp.Stop() // nolint: errcheck

	db, err := db.OpenDB(cfg)
	if err != nil {
		return errors.Wrap(err, "failed to open database connection")
	}

	defer db.Close()

	if err := db.Ping(); err != nil {
		return errors.Wrap(err, "failed to ping database")
	}

	// create a queue that will collect, aggregate, and export blocks and metadata
	exportQueue := processor.NewQueue(100)

	workers := make([]processor.Worker, workerCount, workerCount)
	for i := range workers {
		workers[i] = processor.NewWorker(db, cp, exportQueue)
	}

	wg.Add(1)

	// Start each blocking worker in a go-routine where the worker consumes jobs
	// off of the export queue.
	for i, w := range workers {
		log.Printf("starting worker %d...\n", i+1)

		go w.Start()
	}

	// listen for and trap any OS signal to gracefully shutdown and exit
	trapSignal()

	go startNewBlockListener(exportQueue, cp)
	go enqueueMissingBlocks(exportQueue, cp)

	// block main process (signal capture will call WaitGroup's Done)
	wg.Wait()
	return nil
}

// enqueueMissingBlocks enqueues jobs (block heights) for missed blocks starting
// at the startHeight up until the latest known height.
func enqueueMissingBlocks(exportQueue processor.Queue, cp client.ClientProxy) {
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
func startNewBlockListener(exportQueue processor.Queue, cp client.ClientProxy) {
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
