package start

import (
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	parsecmdtypes "github.com/saifullah619/juno/v3/cmd/parse/types"
	"github.com/saifullah619/juno/v3/types/utils"

	"github.com/saifullah619/juno/v3/logging"

	"github.com/saifullah619/juno/v3/types/config"

	"github.com/go-co-op/gocron"

	"github.com/saifullah619/juno/v3/modules"
	"github.com/saifullah619/juno/v3/parser"
	"github.com/saifullah619/juno/v3/types"

	"github.com/spf13/cobra"
)

var (
	waitGroup sync.WaitGroup
)

// NewStartCmd returns the command that should be run when we want to start parsing a chain state.
func NewStartCmd(cmdCfg *parsecmdtypes.Config) *cobra.Command {
	return &cobra.Command{
		Use:     "start",
		Short:   "Start parsing the blockchain data",
		PreRunE: parsecmdtypes.ReadConfigPreRunE(cmdCfg),
		RunE: func(cmd *cobra.Command, args []string) error {
			context, err := parsecmdtypes.GetParserContext(config.Cfg, cmdCfg)
			if err != nil {
				return err
			}

			// Run all the additional operations
			for _, module := range context.Modules {
				if module, ok := module.(modules.AdditionalOperationsModule); ok {
					err = module.RunAdditionalOperations()
					if err != nil {
						return err
					}
				}
			}

			return startParsing(context)
		},
	}
}

// startParsing represents the function that should be called when the parse command is executed
func startParsing(ctx *parser.Context) error {
	// Get the config
	cfg := config.Cfg.Parser
	logging.StartHeight.Add(float64(cfg.StartHeight))

	// Start periodic operations
	scheduler := gocron.NewScheduler(time.UTC)
	for _, module := range ctx.Modules {
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
	workers := make([]parser.Worker, cfg.Workers)
	for i := range workers {
		workers[i] = parser.NewWorker(ctx, exportQueue, i)
	}

	waitGroup.Add(1)

	// Run all the async operations
	for _, module := range ctx.Modules {
		if module, ok := module.(modules.AsyncOperationsModule); ok {
			go module.RunAsyncOperations()
		}
	}

	// Start each blocking worker in a go-routine where the worker consumes jobs
	// off of the export queue.
	for i, w := range workers {
		ctx.Logger.Debug("starting worker...", "number", i+1)
		go w.Start()
	}

	// Listen for and trap any OS signal to gracefully shutdown and exit
	trapSignal(ctx)

	if cfg.ParseGenesis {
		// Add the genesis to the queue if requested
		exportQueue <- 0
	}

	if cfg.ParseOldBlocks {
		go enqueueMissingBlocks(exportQueue, ctx)
	}

	if cfg.ParseNewBlocks {
		go enqueueNewBlocks(exportQueue, ctx)
	}

	// Block main process (signal capture will call WaitGroup's Done)
	waitGroup.Wait()
	return nil
}

// enqueueMissingBlocks enqueues jobs (block heights) for missed blocks starting
// at the startHeight up until the latest known height.
func enqueueMissingBlocks(exportQueue types.HeightQueue, ctx *parser.Context) {
	// Get the config
	cfg := config.Cfg.Parser

	// Get the latest height
	latestBlockHeight := mustGetLatestHeight(ctx)

	lastDbBlockHeight, err := ctx.Database.GetLastBlockHeight()
	if err != nil {
		ctx.Logger.Error("failed to get last block height from database", "error", err)
	}

	// Get the start height, default to the config's height
	startHeight := cfg.StartHeight

	// Set startHeight to the latest height in database
	// if is not set inside config.yaml file
	if startHeight == 0 {
		startHeight = utils.MaxInt64(1, lastDbBlockHeight)
	}

	if cfg.FastSync {
		ctx.Logger.Info("fast sync is enabled, ignoring all previous blocks", "latest_block_height", latestBlockHeight)
		for _, module := range ctx.Modules {
			if mod, ok := module.(modules.FastSyncModule); ok {
				err := mod.DownloadState(latestBlockHeight)
				if err != nil {
					ctx.Logger.Error("error while performing fast sync",
						"err", err,
						"last_block_height", latestBlockHeight,
						"module", module.Name(),
					)
				}
			}
		}
	} else {
		ctx.Logger.Info("syncing missing blocks...", "latest_block_height", latestBlockHeight)
		for i := startHeight; i <= latestBlockHeight; i++ {
			ctx.Logger.Debug("enqueueing missing block", "height", i)
			exportQueue <- i
		}
	}
}

// enqueueNewBlocks enqueues new block heights onto the provided queue.
func enqueueNewBlocks(exportQueue types.HeightQueue, ctx *parser.Context) {
	currHeight := mustGetLatestHeight(ctx)

	// Enqueue upcoming heights
	for {
		latestBlockHeight := mustGetLatestHeight(ctx)

		// Enqueue all heights from the current height up to the latest height
		for ; currHeight <= latestBlockHeight; currHeight++ {
			ctx.Logger.Debug("enqueueing new block", "height", currHeight)
			exportQueue <- currHeight
		}
		time.Sleep(config.GetAvgBlockTime())
	}
}

// mustGetLatestHeight tries getting the latest height from the RPC client.
// If after 50 tries no latest height can be found, it returns 0.
func mustGetLatestHeight(ctx *parser.Context) int64 {
	for retryCount := 0; retryCount < 50; retryCount++ {
		latestBlockHeight, err := ctx.Node.LatestHeight()
		if err == nil {
			return latestBlockHeight
		}

		ctx.Logger.Error("failed to get last block from RPCConfig client",
			"err", err,
			"retry interval", config.GetAvgBlockTime(),
			"retry count", retryCount)

		time.Sleep(config.GetAvgBlockTime())
	}

	return 0
}

// trapSignal will listen for any OS signal and invoke Done on the main
// WaitGroup allowing the main process to gracefully exit.
func trapSignal(ctx *parser.Context) {
	var sigCh = make(chan os.Signal, 1)

	signal.Notify(sigCh, syscall.SIGTERM)
	signal.Notify(sigCh, syscall.SIGINT)

	go func() {
		sig := <-sigCh
		ctx.Logger.Info("caught signal; shutting down...", "signal", sig.String())
		defer ctx.Node.Stop()
		defer ctx.Database.Close()
		defer waitGroup.Done()
	}()
}
