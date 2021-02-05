package cmd

import (
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	modsregistrar "github.com/desmos-labs/juno/modules/registrar"

	"github.com/cosmos/cosmos-sdk/simapp/params"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/go-co-op/gocron"
	"github.com/rs/zerolog/log"
	tmtypes "github.com/tendermint/tendermint/types"

	"github.com/desmos-labs/juno/client"
	"github.com/desmos-labs/juno/modules"
	"github.com/desmos-labs/juno/types"
	"github.com/desmos-labs/juno/worker"

	"github.com/desmos-labs/juno/config"
	"github.com/desmos-labs/juno/db"

	"github.com/spf13/cobra"
)

var (
	waitGroup sync.WaitGroup
)

// ParseCmd returns the command that should be run when we want to start parsing a chain state.
func ParseCmd(
	name string,
	registrar modsregistrar.Registrar, encodingConfigBuilder types.EncodingConfigBuilder,
	setupCfg types.SdkConfigSetup, buildDb db.Builder,
) *cobra.Command {
	return &cobra.Command{
		Use:     "parse",
		Short:   "Start parsing the blockchain data",
		PreRunE: concatCobraCmdFuncs(readConfig(name), setupLogging),
		RunE: func(cmd *cobra.Command, args []string) error {
			cdc, cp, database, registeredModules, err := SetupParsing(
				registrar, encodingConfigBuilder, setupCfg, buildDb,
			)
			if err != nil {
				return err
			}

			return StartParsing(cdc, cp, database, registeredModules)
		},
	}
}

// SetupParsing setups all the things that should be later passed to StartParsing in order
// to parse the chain data properly.
func SetupParsing(
	registrar modsregistrar.Registrar, buildEncodingConfig types.EncodingConfigBuilder,
	setupCfg types.SdkConfigSetup, buildDb db.Builder,
) (*params.EncodingConfig, *client.Proxy, db.Database, []modules.Module, error) {
	// Get the global config
	cfg := config.Cfg

	// Build the codec
	encodingConfig := buildEncodingConfig()

	// Setup the SDK configuration
	sdkConfig := sdk.GetConfig()
	setupCfg(cfg, sdkConfig)
	sdkConfig.Seal()

	// Get the database
	database, err := buildDb(cfg, &encodingConfig)
	if err != nil {
		return nil, nil, nil, nil, err
	}

	// Init the client
	cp, err := client.NewClientProxy(cfg, &encodingConfig)
	if err != nil {
		return nil, nil, nil, nil, fmt.Errorf("failed to start client: %s", err)
	}

	// Get the modules
	mods := registrar.BuildModules(cfg, &encodingConfig, sdkConfig, database, cp)
	registeredModules := modsregistrar.GetModules(mods, cfg.Cosmos.Modules)

	// Run all the additional operations
	for _, module := range registeredModules {
		if module, ok := module.(modules.AdditionalOperationsModule); ok {
			err := module.RunAdditionalOperations()
			if err != nil {
				return nil, nil, nil, nil, err
			}
		}
	}

	return &encodingConfig, cp, database, registeredModules, nil
}

// parseCmdHandler represents the function that should be called when the parse command is executed
func StartParsing(
	encodingConfig *params.EncodingConfig, cp *client.Proxy, db db.Database, registeredModules []modules.Module,
) error {
	// Get the config
	cfg := config.Cfg.Parsing

	// Start periodic operations
	scheduler := gocron.NewScheduler(time.UTC)
	for _, module := range registeredModules {
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
	workers := make([]worker.Worker, cfg.Workers, cfg.Workers)
	for i := range workers {
		workers[i] = worker.NewWorker(encodingConfig, exportQueue, cp, db, registeredModules)
	}

	waitGroup.Add(1)

	// Run all the async operations
	for _, module := range registeredModules {
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
	trapSignal(cp)

	if cfg.ParseOldBlocks {
		go enqueueMissingBlocks(registeredModules, exportQueue, cp)
	}

	if cfg.ParseNewBlocks {
		go startNewBlockListener(exportQueue, cp)
	}

	// Block main process (signal capture will call WaitGroup's Done)
	waitGroup.Wait()
	return nil
}

// enqueueMissingBlocks enqueues jobs (block heights) for missed blocks starting
// at the startHeight up until the latest known height.
func enqueueMissingBlocks(registeredModules []modules.Module, exportQueue types.HeightQueue, cp *client.Proxy) {
	// Get the config
	cfg := config.Cfg.Parsing

	// Get the latest height
	latestBlockHeight, err := cp.LatestHeight()
	if err != nil {
		log.Fatal().Err(fmt.Errorf("failed to get last block from RPC client: %s", err))
	}

	startHeight := cfg.StartHeight
	if cfg.FastSync {
		log.Debug().Msg("fast sync is enabled, ignoring all previous blocks")
		for _, module := range registeredModules {
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
		log.Debug().Int64("latest_block_height", latestBlockHeight).Msg("syncing missing blocks...")
		for i := startHeight; i <= latestBlockHeight; i++ {
			log.Debug().Int64("height", i).Msg("enqueueing missing block")
			exportQueue <- i
		}
	}
}

// startNewBlockListener subscribes to new block events via the Tendermint RPC
// and enqueues each new block height onto the provided queue. It blocks as new
// blocks are incoming.
func startNewBlockListener(exportQueue types.HeightQueue, cp *client.Proxy) {
	eventCh, cancel, err := cp.SubscribeNewBlocks("juno-client")
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
func trapSignal(cp *client.Proxy) {
	var sigCh = make(chan os.Signal)

	signal.Notify(sigCh, syscall.SIGTERM)
	signal.Notify(sigCh, syscall.SIGINT)

	go func() {
		sig := <-sigCh
		log.Info().Str("signal", sig.String()).Msg("caught signal; shutting down...")
		defer cp.Stop() // nolint: errcheck
		defer waitGroup.Done()
	}()
}
