package parse

import (
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	modsregistrar "github.com/desmos-labs/juno/modules/registrar"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/go-co-op/gocron"
	"github.com/rs/zerolog/log"
	tmtypes "github.com/tendermint/tendermint/types"

	"github.com/desmos-labs/juno/client"
	"github.com/desmos-labs/juno/modules"
	"github.com/desmos-labs/juno/types"
	"github.com/desmos-labs/juno/worker"

	"github.com/desmos-labs/juno/db"

	"github.com/spf13/cobra"
)

var (
	waitGroup sync.WaitGroup
)

// ParseCmd returns the command that should be run when we want to start parsing a chain state.
func ParseCmd(cfg *Config) *cobra.Command {
	return &cobra.Command{
		Use:     "parse",
		Short:   "Start parsing the blockchain data",
		PreRunE: types.ConcatCobraCmdFuncs(ReadConfig(cfg), setupLogging),
		RunE: func(cmd *cobra.Command, args []string) error {
			parserData, err := SetupParsing(cfg)
			if err != nil {
				return err
			}

			return StartParsing(parserData)
		},
	}
}

// SetupParsing setups all the things that should be later passed to StartParsing in order
// to parse the chain data properly.
func SetupParsing(parseConfig *Config) (*ParserData, error) {
	// Get the global config
	cfg := types.Cfg

	// Build the codec
	encodingConfig := parseConfig.GetEncodingConfigBuilder()()

	// Setup the SDK configuration
	sdkConfig := sdk.GetConfig()
	parseConfig.GetSetupConfig()(cfg, sdkConfig)
	sdkConfig.Seal()

	// Get the database
	database, err := parseConfig.GetDBBuilder()(cfg, &encodingConfig)
	if err != nil {
		return nil, err
	}

	// Init the client
	cp, err := client.NewClientProxy(cfg, &encodingConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to start client: %s", err)
	}

	// Get the modules
	mods := parseConfig.GetRegistrar().BuildModules(cfg, &encodingConfig, sdkConfig, database, cp)
	registeredModules := modsregistrar.GetModules(mods, cfg.GetCosmosConfig().Modules)

	// Run all the additional operations
	for _, module := range registeredModules {
		if module, ok := module.(modules.AdditionalOperationsModule); ok {
			err := module.RunAdditionalOperations()
			if err != nil {
				return nil, err
			}
		}
	}

	return NewParserData(&encodingConfig, cp, database, registeredModules), nil
}

// StartParsing represents the function that should be called when the parse command is executed
func StartParsing(data *ParserData) error {
	// Get the config
	cfg := types.Cfg.GetParsingConfig()

	// Start periodic operations
	scheduler := gocron.NewScheduler(time.UTC)
	for _, module := range data.Modules {
		if module, ok := module.(modules.PeriodicOperationsModule); ok {
			err := module.RegisterPeriodicOperations(scheduler)
			if err != nil {
				return err
			}
		}
	}
	scheduler.StartAsync()

	// Create a queue that will collect, aggregate, and export blocks and metadata
	exportQueue := types.NewQueue(25)

	// Create workers
	config := worker.NewConfig(exportQueue, data.EncodingConfig, data.Proxy, data.Database, data.Modules)
	workers := make([]worker.Worker, cfg.Workers, cfg.Workers)
	for i := range workers {
		workers[i] = worker.NewWorker(config)
	}

	waitGroup.Add(1)

	// Run all the async operations
	for _, module := range data.Modules {
		if module, ok := module.(modules.AsyncOperationsModule); ok {
			go module.RunAsyncOperations()
		}
	}

	// Start each blocking worker in a go-routine where the worker consumes jobs
	// off of the export queue.
	for i, w := range workers {
		log.Debug().Int("number", i+1).Msg("starting worker...")

		go w.Start()
	}

	// Listen for and trap any OS signal to gracefully shutdown and exit
	trapSignal(data.Proxy, data.Database)

	if cfg.ParseGenesis {
		// Add the genesis to the queue if requested
		exportQueue <- 1
	}

	if cfg.ParseOldBlocks {
		go enqueueMissingBlocks(exportQueue, data)
	}

	if cfg.ParseNewBlocks {
		go startNewBlockListener(exportQueue, data)
	}

	// Block main process (signal capture will call WaitGroup's Done)
	waitGroup.Wait()
	return nil
}

// enqueueMissingBlocks enqueues jobs (block heights) for missed blocks starting
// at the startHeight up until the latest known height.
func enqueueMissingBlocks(exportQueue types.HeightQueue, data *ParserData) {
	// Get the config
	cfg := types.Cfg.GetParsingConfig()

	// Get the latest height
	latestBlockHeight, err := data.Proxy.LatestHeight()
	if err != nil {
		log.Fatal().Err(fmt.Errorf("failed to get last block from RPC client: %s", err))
	}

	if cfg.FastSync {
		log.Info().Int64("latest_block_height", latestBlockHeight).
			Msg("fast sync is enabled, ignoring all previous blocks")
		for _, module := range data.Modules {
			if mod, ok := module.(modules.FastSyncModule); ok {
				err := mod.DownloadState(latestBlockHeight)
				if err != nil {
					log.Error().Err(err).
						Int64("last_block_height", latestBlockHeight).
						Str("module", module.Name()).
						Msg("error while performing fast sync")
				}
			}
		}
	} else {
		log.Info().Int64("latest_block_height", latestBlockHeight).
			Msg("syncing missing blocks...")
		for i := cfg.StartHeight; i <= latestBlockHeight; i++ {
			log.Debug().Int64("height", i).Msg("enqueueing missing block")
			exportQueue <- i
		}
	}
}

// startNewBlockListener subscribes to new block events via the Tendermint RPC
// and enqueues each new block height onto the provided queue. It blocks as new
// blocks are incoming.
func startNewBlockListener(exportQueue types.HeightQueue, data *ParserData) {
	eventCh, cancel, err := data.Proxy.SubscribeNewBlocks("juno-client")
	defer cancel()

	if err != nil {
		log.Fatal().Err(fmt.Errorf("failed to subscribe to new blocks: %s", err))
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
func trapSignal(cp *client.Proxy, db db.Database) {
	var sigCh = make(chan os.Signal)

	signal.Notify(sigCh, syscall.SIGTERM)
	signal.Notify(sigCh, syscall.SIGINT)

	go func() {
		sig := <-sigCh
		log.Info().Str("signal", sig.String()).Msg("caught signal; shutting down...")
		defer cp.Stop()
		defer db.Close()
		defer waitGroup.Done()
	}()
}
