package parse

import (
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/desmos-labs/juno/types/logging"

	"github.com/go-co-op/gocron"
	tmtypes "github.com/tendermint/tendermint/types"

	"github.com/desmos-labs/juno/modules"
	"github.com/desmos-labs/juno/types"
	"github.com/desmos-labs/juno/worker"

	"github.com/spf13/cobra"
)

var (
	waitGroup sync.WaitGroup
)

// ParseCmd returns the command that should be run when we want to start parsing a chain state.
func ParseCmd(cmdCfg *Config) *cobra.Command {
	return &cobra.Command{
		Use:     "parse",
		Short:   "Start parsing the blockchain data",
		PreRunE: ReadConfig(cmdCfg),
		RunE: func(cmd *cobra.Command, args []string) error {
			context, err := GetParsingContext(cmdCfg)
			if err != nil {
				return err
			}

			go StartPrometheus()

			return StartParsing(context)
		},
	}
}

// StartPrometheus allows to start a Telemetry server used to expose useful metrics
func StartPrometheus() {
	cfg := types.Cfg.GetTelemetryConfig()
	if !cfg.IsEnabled() {
		return
	}

	router := mux.NewRouter()
	router.Handle("/metrics", promhttp.Handler())

	err := http.ListenAndServe(fmt.Sprintf(":%d", cfg.GetPort()), router)
	if err != nil {
		panic(err)
	}
}

// StartParsing represents the function that should be called when the parse command is executed
func StartParsing(ctx *Context) error {
	// Get the config
	cfg := types.Cfg.GetParsingConfig()
	logging.StartHeight.Add(float64(cfg.GetStartHeight()))

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
	workerCtx := worker.NewContext(ctx.EncodingConfig, ctx.Proxy, ctx.Database, ctx.Logger, exportQueue, ctx.Modules)
	workers := make([]worker.Worker, cfg.GetWorkers(), cfg.GetWorkers())
	for i := range workers {
		workers[i] = worker.NewWorker(i, workerCtx)
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

	if cfg.ShouldParseGenesis() {
		// Add the genesis to the queue if requested
		exportQueue <- 0
	}

	if cfg.ShouldParseOldBlocks() {
		go enqueueMissingBlocks(exportQueue, ctx)
	}

	if cfg.ShouldParseNewBlocks() {
		go startNewBlockListener(exportQueue, ctx)
	}

	// Block main process (signal capture will call WaitGroup's Done)
	waitGroup.Wait()
	return nil
}

// enqueueMissingBlocks enqueues jobs (block heights) for missed blocks starting
// at the startHeight up until the latest known height.
func enqueueMissingBlocks(exportQueue types.HeightQueue, ctx *Context) {
	// Get the config
	cfg := types.Cfg.GetParsingConfig()

	// Get the latest height
	latestBlockHeight, err := ctx.Proxy.LatestHeight()
	if err != nil {
		panic(fmt.Errorf("failed to get last block from RPC client: %s", err))
	}

	if cfg.UseFastSync() {
		ctx.Logger.Info("fast sync is enabled, ignoring all previous blocks", "latest_block_height", latestBlockHeight)
		for _, module := range ctx.Modules {
			if mod, ok := module.(modules.FastSyncModule); ok {
				err = mod.DownloadState(latestBlockHeight)
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
		for i := cfg.GetStartHeight(); i <= latestBlockHeight; i++ {
			ctx.Logger.Debug("enqueueing missing block", "height", i)
			exportQueue <- i
		}
	}
}

// startNewBlockListener subscribes to new block events via the Tendermint RPC
// and enqueues each new block height onto the provided queue. It blocks as new
// blocks are incoming.
func startNewBlockListener(exportQueue types.HeightQueue, ctx *Context) {
	eventCh, cancel, err := ctx.Proxy.SubscribeNewBlocks(types.Cfg.GetRPCConfig().GetClientName() + "-blocks")
	defer cancel()

	if err != nil {
		panic(fmt.Errorf("failed to subscribe to new blocks: %s", err))
	}

	ctx.Logger.Info("listening for new block events...")

	for e := range eventCh {
		newBlock := e.Data.(tmtypes.EventDataNewBlock).Block
		height := newBlock.Header.Height

		ctx.Logger.Debug("enqueueing new block", "height", height)
		exportQueue <- height
	}
}

// trapSignal will listen for any OS signal and invoke Done on the main
// WaitGroup allowing the main process to gracefully exit.
func trapSignal(ctx *Context) {
	var sigCh = make(chan os.Signal)

	signal.Notify(sigCh, syscall.SIGTERM)
	signal.Notify(sigCh, syscall.SIGINT)

	go func() {
		sig := <-sigCh
		ctx.Logger.Info("caught signal; shutting down...", "signal", sig.String())
		defer ctx.Proxy.Stop()
		defer ctx.Database.Close()
		defer waitGroup.Done()
	}()
}
