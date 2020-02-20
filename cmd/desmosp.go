package cmd

import (
	"fmt"
	"github.com/angelorc/desmos-parser/client"
	"github.com/angelorc/desmos-parser/config"
	"github.com/angelorc/desmos-parser/db"
	"github.com/angelorc/desmos-parser/processor"
	"github.com/pkg/errors"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"

	sdk "github.com/cosmos/cosmos-sdk/types"
	tmtypes "github.com/tendermint/tendermint/types"
)

const (
	logLevelJSON = "json"
	logLevelText = "text"
)

var (
	startHeight int64
	workerCount int16
	logLevel    string
	logFormat   string

	wg sync.WaitGroup
)

var rootCmd = &cobra.Command{
	Use:   "desmos-parser [config-file]",
	Args:  cobra.ExactArgs(1),
	Short: "Desmos Parser",
	RunE:  desmospCmdHandler,
}

func init() {
	rootCmd.PersistentFlags().Int64Var(&startHeight, "start-height", 1, "sync missing or failed blocks starting from a given height")
	rootCmd.PersistentFlags().Int16Var(&workerCount, "workers", 1, "number of workers to run concurrently")
	rootCmd.PersistentFlags().StringVar(&logLevel, "log-level", zerolog.InfoLevel.String(), "logging level")
	rootCmd.PersistentFlags().StringVar(&logFormat, "log-format", logLevelJSON, "logging format; must be either json or text")

	rootCmd.AddCommand(getVersionCmd())
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

const (
	// Bech32PrefixAccAddr defines the Bech32 prefix of an account's address
	Bech32PrefixAccAddr = "desmos"
	// Bech32PrefixAccPub defines the Bech32 prefix of an account's public key
	Bech32PrefixAccPub = Bech32PrefixAccAddr + "pub"
	// Bech32PrefixValAddr defines the Bech32 prefix of a validator's operator address
	Bech32PrefixValAddr = Bech32PrefixAccAddr + "valoper"
	// Bech32PrefixValPub defines the Bech32 prefix of a validator's operator public key
	Bech32PrefixValPub = Bech32PrefixAccAddr + "valoperpub"
	// Bech32PrefixConsAddr defines the Bech32 prefix of a consensus node address
	Bech32PrefixConsAddr = Bech32PrefixAccAddr + "valcons"
	// Bech32PrefixConsPub defines the Bech32 prefix of a consensus node public key
	Bech32PrefixConsPub = Bech32PrefixAccAddr + "valconspub"
)

func desmospCmdHandler(cmd *cobra.Command, args []string) error {
	logLvl, err := zerolog.ParseLevel(logLevel)
	if err != nil {
		return err
	}

	zerolog.SetGlobalLevel(logLvl)

	switch logFormat {
	case logLevelJSON:
		// JSON is the default logging format

	case logLevelText:
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	default:
		return fmt.Errorf("invalid logging format: %s", logFormat)
	}

	sdkConfig := sdk.GetConfig()
	sdkConfig.SetBech32PrefixForAccount(Bech32PrefixAccAddr, Bech32PrefixAccPub)
	sdkConfig.SetBech32PrefixForValidator(Bech32PrefixValAddr, Bech32PrefixValPub)
	sdkConfig.SetBech32PrefixForConsensusNode(Bech32PrefixConsAddr, Bech32PrefixConsPub)
	sdkConfig.Seal()

	cfgFile := args[0]
	cfg := config.ParseConfig(cfgFile)

	cp, err := client.New(cfg.RPCNode, cfg.ClientNode)
	if err != nil {
		return errors.Wrap(err, "failed to start RPC client")
	}

	defer cp.Stop() // nolint: errcheck

	// Init MongoDB
	db, err := db.OpenDB()
	if err != nil {
		return errors.Wrap(err, "failed to open mongodb connection")
	}
	// End MongoDB

	// create a queue that will collect, aggregate, and export blocks and metadata
	exportQueue := processor.NewQueue(25)

	workers := make([]processor.Worker, workerCount, workerCount)
	for i := range workers {
		workers[i] = processor.NewWorker(cp, exportQueue, db)
	}

	wg.Add(1)

	// Start each blocking worker in a go-routine where the worker consumes jobs
	// off of the export queue.
	for i, w := range workers {
		log.Info().Int("number", i+1).Msg("starting worker...")

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
		log.Fatal().Err(errors.Wrap(err, "failed to get lastest block from RPC client"))
	}

	log.Info().Int64("latestBlockHeight", latestBlockHeight).Msg("syncing missing blocks...")

	for i := startHeight; i <= latestBlockHeight; i++ {
		log.Info().Int64("height", i).Msg("enqueueing missing block")
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
