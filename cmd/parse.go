package cmd

import (
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/cosmos/cosmos-sdk/simapp/params"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/go-co-op/gocron"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
	tmtypes "github.com/tendermint/tendermint/types"

	"github.com/desmos-labs/juno/client"
	"github.com/desmos-labs/juno/modules"
	"github.com/desmos-labs/juno/types"
	"github.com/desmos-labs/juno/worker"

	"github.com/desmos-labs/juno/config"
	"github.com/desmos-labs/juno/db"

	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
)

const (
	LogLevelJSON = "json"
	LogLevelText = "text"
)

var (
	waitGroup sync.WaitGroup
)

// ParseCmd returns the command that should be run when we want to start parsing a chain state.
func ParseCmd(
	registrar modules.Registrar,
	encodingConfigBuilder config.EncodingConfigBuilder, setupCfg config.SdkConfigSetup, buildDb db.Builder,
) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "parse [config-file]",
		Short: "Start parsing a blockchain using the provided config file",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cdc, cp, database, registeredModules, err := SetupParsing(
				args, registrar, encodingConfigBuilder, setupCfg, buildDb,
			)
			if err != nil {
				return err
			}

			return StartParsing(cdc, cp, database, registeredModules)
		},
	}

	return SetupFlags(cmd)
}

// SetupParsing setups all the things that should be later passed to StartParsing in order
// to parse the chain data properly.
func SetupParsing(
	args []string, registrar modules.Registrar,
	buildEncodingConfig config.EncodingConfigBuilder, setupCfg config.SdkConfigSetup, buildDb db.Builder,
) (*params.EncodingConfig, *client.Proxy, db.Database, []modules.Module, error) {
	// Setup the logger
	err := setupLogging()
	if err != nil {
		return nil, nil, nil, nil, err
	}

	// Setup the config
	cfg, err := config.Read(args[0])
	if err != nil {
		return nil, nil, nil, nil, err
	}

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
	cp, err := client.New(cfg, &encodingConfig)
	if err != nil {
		return nil, nil, nil, nil, fmt.Errorf("failed to start client: %s", err)
	}

	// nolint:errcheck
	defer cp.Stop()

	// Get the modules
	mods := registrar.BuildModules(cfg, &encodingConfig, sdkConfig, database, cp)
	registeredModules := modules.GetModules(mods, cfg.CosmosConfig.Modules)

	return &encodingConfig, cp, database, registeredModules, nil
}

// SetupFlags setups all the flags for the parse command
func SetupFlags(cmd *cobra.Command) *cobra.Command {
	cmd.Flags().Int64(FlagStartHeight, 1, "sync missing or failed blocks starting from a given height")
	cmd.Flags().Int64(FlagWorkerCount, 1, "number of workers to run concurrently")
	cmd.Flags().Bool(FlagParseOldBlocks, true, "parse old blocks")
	cmd.Flags().Bool(FlagListenNewBlocks, true, "listen to new blocks")
	cmd.Flags().String(FlagLogLevel, zerolog.InfoLevel.String(), "logging level")
	cmd.Flags().String(FlagLogFormat, LogLevelJSON, "logging format; must be either json or text")
	return cmd
}

// setupLogging setups the logging for the entire project
func setupLogging() error {
	// Init logging level
	logLvl, err := zerolog.ParseLevel(viper.GetString(FlagLogLevel))
	if err != nil {
		return err
	}
	zerolog.SetGlobalLevel(logLvl)

	// Init logging format
	logFormat := viper.GetString(FlagLogFormat)
	switch logFormat {
	case LogLevelJSON:
		// JSON is the default logging format
		break

	case LogLevelText:
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
		break

	default:
		return fmt.Errorf("invalid logging format: %s", logFormat)
	}
	return err
}

// parseCmdHandler represents the function that should be called when the parse command is executed
func StartParsing(encodingConfig *params.EncodingConfig, cp *client.Proxy, db db.Database, modules []modules.Module) error {
	// Run all the additional operations
	for _, module := range modules {
		go module.RunAdditionalOperations()
	}

	// Start periodic operations
	scheduler := gocron.NewScheduler(time.UTC)
	for _, module := range modules {
		err := module.RegisterPeriodicOperations(scheduler)
		if err != nil {
			return err
		}
	}
	scheduler.StartAsync()

	// Create a queue that will collect, aggregate, and export blocks and metadata
	exportQueue := types.NewQueue(25)

	// Create workers
	workerCount := viper.GetInt64(FlagWorkerCount)
	workers := make([]worker.Worker, workerCount, workerCount)
	for i := range workers {
		workers[i] = worker.NewWorker(encodingConfig, exportQueue, cp, db, modules)
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

	if viper.GetBool(FlagParseOldBlocks) {
		go enqueueMissingBlocks(exportQueue, cp)
	}

	if viper.GetBool(FlagListenNewBlocks) {
		go startNewBlockListener(exportQueue, cp)
	}

	// Block main process (signal capture will call WaitGroup's Done)
	waitGroup.Wait()
	return nil
}

// enqueueMissingBlocks enqueues jobs (block heights) for missed blocks starting
// at the startHeight up until the latest known height.
func enqueueMissingBlocks(exportQueue types.Queue, cp *client.Proxy) {
	latestBlockHeight, err := cp.LatestHeight()
	if err != nil {
		log.Fatal().Err(fmt.Errorf("failed to get last block from RPC client: %s", err))
	}

	log.Debug().Int64("latestBlockHeight", latestBlockHeight).Msg("syncing missing blocks...")

	startHeight := viper.GetInt64(FlagStartHeight)
	for i := startHeight; i <= latestBlockHeight; i++ {
		log.Debug().Int64("height", i).Msg("enqueueing missing block")
		exportQueue <- i
	}
}

// startNewBlockListener subscribes to new block events via the Tendermint RPC
// and enqueues each new block height onto the provided queue. It blocks as new
// blocks are incoming.
func startNewBlockListener(exportQueue types.Queue, cp *client.Proxy) {
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
