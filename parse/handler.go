package parse

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/desmos-labs/juno/config"
	"github.com/desmos-labs/juno/db"
	"github.com/desmos-labs/juno/parse/client"
	"github.com/desmos-labs/juno/parse/worker"
	"github.com/desmos-labs/juno/types"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
	tmtypes "github.com/tendermint/tendermint/types"
)

// parseCmdHandler represents the function that should be called when the parse command is executed
func ParseCmdHandler(codec *codec.Codec, dbBuilder db.Builder, configPath string, ops []AdditionalOperation) error {

	// Setup the logger
	err := SetupLogging()
	if err != nil {
		return err
	}

	// Setup the config
	cfg, err := SetupConfig(configPath)
	if err != nil {
		return err
	}

	// Init the client
	cp, err := client.New(*cfg, codec)
	if err != nil {
		return errors.Wrap(err, "failed to start RPC client")
	}
	defer cp.Stop()

	// Create a queue that will collect, aggregate, and export blocks and metadata
	exportQueue := types.NewQueue(25)

	database, err := dbBuilder(*cfg, codec)
	if err != nil {
		return errors.Wrap(err, "failed to open database connection")
	}

	// Create workers
	workerCount := viper.GetInt64(config.FlagWorkerCount)
	workers := make([]worker.Worker, workerCount, workerCount)
	for i := range workers {
		workers[i] = worker.NewWorker(codec, cp, exportQueue, *database)
	}

	waitGroup.Add(1)

	// Start each blocking worker in a go-routine where the worker consumes jobs
	// off of the export queue.
	for i, w := range workers {
		log.Debug().Int("number", i+1).Msg("starting worker...")

		go w.Start()
	}

	// Listen for and trap any OS signal to gracefully shutdown and exit
	trapSignal()

	if viper.GetBool(config.FlagParseOldBlocks) {
		go enqueueMissingBlocks(exportQueue, cp)
	}

	if viper.GetBool(config.FlagListenNewBlocks) {
		go startNewBlockListener(exportQueue, cp)
	}

	// Perform additional operations
	for _, op := range ops {
		if err := op(*cfg, codec, cp, *database); err != nil {
			return err
		}
	}

	// Block main process (signal capture will call WaitGroup's Done)
	waitGroup.Wait()
	return nil
}

// enqueueMissingBlocks enqueues jobs (block heights) for missed blocks starting
// at the startHeight up until the latest known height.
func enqueueMissingBlocks(exportQueue types.Queue, cp client.ClientProxy) {
	latestBlockHeight, err := cp.LatestHeight()
	if err != nil {
		log.Fatal().Err(errors.Wrap(err, "failed to get lastest block from RPC client"))
	}

	log.Debug().Int64("latestBlockHeight", latestBlockHeight).Msg("syncing missing blocks...")

	startHeight := viper.GetInt64(config.FlagStartHeight)
	for i := startHeight; i <= latestBlockHeight; i++ {
		log.Debug().Int64("height", i).Msg("enqueueing missing block")
		exportQueue <- i
	}
}

// startNewBlockListener subscribes to new block events via the Tendermint RPC
// and enqueues each new block height onto the provided queue. It blocks as new
// blocks are incoming.
func startNewBlockListener(exportQueue types.Queue, cp client.ClientProxy) {
	eventCh, cancel, err := cp.SubscribeNewBlocks("juno-client")
	defer cancel()

	if err != nil {
		log.Fatal().Err(errors.Wrap(err, "failed to subscribe to new blocks"))
	}

	log.Info().Msg("listening for new block events...")

	for e := range eventCh {
		newBlock := e.Data.(tmtypes.EventDataNewBlock).Block
		height := newBlock.Header.Height

		log.Debug().Int64("height", height).Msg("enqueueing new block")
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
		log.Info().Str("signal", sig.String()).Msg("caught signal; shutting down...")
		defer waitGroup.Done()
	}()
}
