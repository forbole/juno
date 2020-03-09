package parse

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/desmos-labs/juno/config"
	"github.com/desmos-labs/juno/db"
	"github.com/desmos-labs/juno/parse/client"
	"github.com/desmos-labs/juno/parse/worker"
	"github.com/desmos-labs/juno/types"
	"github.com/pkg/errors"
	"github.com/spf13/viper"

	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"

	tmtypes "github.com/tendermint/tendermint/types"
)

const (
	logLevelJSON = "json"
	logLevelText = "text"
)

var (
	wg sync.WaitGroup
)

// GetParseCmd returns the command that should be run when we want to start parsing a chain state
func GetParseCmd(cdc *codec.Codec, builder db.Builder) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "parse [config-file]",
		Short: "Start parsing a blockchain using the provided config file",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return ParseCmdHandler(cdc, builder, args[0])
		},
	}

	return SetupFlags(cmd)
}

// SetupFlags allows to setup the given cmd by setting the required parse flags
func SetupFlags(cmd *cobra.Command) *cobra.Command {
	cmd.Flags().Int64(config.FlagStartHeight, 1, "sync missing or failed blocks starting from a given height")
	cmd.Flags().Int64(config.FlagWorkerCount, 1, "number of workers to run concurrently")
	cmd.Flags().String(config.FlagLogLevel, zerolog.InfoLevel.String(), "logging level")
	cmd.Flags().String(config.FlagLogFormat, logLevelJSON, "logging format; must be either json or text")
	return cmd
}

// parseCmdHandler represents the function that should be called when the parse command is executed
func ParseCmdHandler(codec *codec.Codec, dbBuilder db.Builder, configPath string) error {

	// Init logging level
	logLvl, err := zerolog.ParseLevel(viper.GetString(config.FlagLogLevel))
	if err != nil {
		return err
	}
	zerolog.SetGlobalLevel(logLvl)

	// Init logging format
	logFormat := viper.GetString(config.FlagLogFormat)
	switch logFormat {
	case logLevelJSON:
		// JSON is the default logging format
		break

	case logLevelText:
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
		break

	default:
		return fmt.Errorf("invalid logging format: %s", logFormat)
	}

	// Init config
	log.Info().Msg("Reading config file")
	cfg, err := config.ParseConfig(configPath)
	if err != nil {
		return err
	}

	// Init the client
	cp, err := client.New(*cfg, codec)
	if err != nil {
		return errors.Wrap(err, "failed to start RPC client")
	}
	defer cp.Stop() // nolint: errcheck

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

	wg.Add(1)

	// Start each blocking worker in a go-routine where the worker consumes jobs
	// off of the export queue.
	for i, w := range workers {
		log.Info().Int("number", i+1).Msg("starting worker...")

		go w.Start()
	}

	// Listen for and trap any OS signal to gracefully shutdown and exit
	trapSignal()

	go startNewBlockListener(exportQueue, cp)
	go enqueueMissingBlocks(exportQueue, cp)

	// Block main process (signal capture will call WaitGroup's Done)
	wg.Wait()
	return nil
}

// enqueueMissingBlocks enqueues jobs (block heights) for missed blocks starting
// at the startHeight up until the latest known height.
func enqueueMissingBlocks(exportQueue types.Queue, cp client.ClientProxy) {
	latestBlockHeight, err := cp.LatestHeight()
	if err != nil {
		log.Fatal().Err(errors.Wrap(err, "failed to get lastest block from RPC client"))
	}

	log.Info().Int64("latestBlockHeight", latestBlockHeight).Msg("syncing missing blocks...")

	startHeight := viper.GetInt64(config.FlagStartHeight)
	for i := startHeight; i <= latestBlockHeight; i++ {
		log.Info().Int64("height", i).Msg("enqueueing missing block")
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

		log.Info().Int64("height", height).Msg("enqueueing new block")
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
		defer wg.Done()
	}()
}
