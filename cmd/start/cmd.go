package start

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"

	parsecmdtypes "github.com/forbole/juno/v5/cmd/parse/types"
	"github.com/forbole/juno/v5/interfaces"

	"github.com/forbole/juno/v5/logging"

	"github.com/forbole/juno/v5/types/config"
	"github.com/forbole/juno/v5/types/utils"

	"github.com/forbole/juno/v5/enqueuer"
	"github.com/forbole/juno/v5/operator"
	"github.com/forbole/juno/v5/parser"
	"github.com/forbole/juno/v5/types"

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
			infrastructures, err := parsecmdtypes.GetInfrastructures(config.Cfg, cmdCfg)
			if err != nil {
				return err
			}

			ctx := parser.NewContext(context.Background(), infrastructures.Node, infrastructures.Database, infrastructures.Logger, infrastructures.Modules)

			// Register all the additional operations modules
			operator := operator.NewAdditionalOperator()
			for _, module := range ctx.Modules() {
				if module, ok := module.(interfaces.AdditionalOperationsModule); ok {
					operator.Register(module)
				}
			}
			if err := operator.Start(); err != nil {
				return err
			}

			// Listen for and trap any OS signal to gracefully shutdown and exit
			trapSignal(infrastructures)
			waitGroup.Add(1)

			cfg := config.Cfg.Parser
			logging.StartHeight.Add(float64(cfg.StartHeight))
			if err := startParsing(ctx, cfg.StartHeight, cfg.Workers, cfg.ParseOldBlocks, cfg.ParseNewBlocks); err != nil {
				return err
			}

			// Block main process (signal capture will call WaitGroup's Done)
			waitGroup.Wait()
			return nil
		},
	}
}

// startParsing represents the function that should be called when the parse command is executed
func startParsing(ctx interfaces.Context, startHeight int64, workerAmount int64, parseOldBlocks bool, parseNewBlocks bool) error {
	// Create a queue that will collect, aggregate, and export blocks and metadata
	exportQueue := types.NewQueue(25)

	// Create workers
	workers := make([]parser.Worker, workerAmount)
	for i := range workers {
		workers[i] = parser.NewWorker(i)
	}

	// Start each blocking worker in a go-routine where the worker consumes jobs
	// off of the export queue.
	for i, w := range workers {
		ctx.Logger().Debug("starting worker...", "number", i+1)
		go w.Start(ctx, exportQueue)
	}

	// Start new block enqueuer
	// NOTE: we start the new block enqueuer before the missing block enqueuer to avoid from missing blocks
	if parseNewBlocks {
		go enqueuer.NewNewBlockEnqueuer().ListenAndEnqueueBlocks(ctx, exportQueue)
	}

	// Start missing block enqueuer
	latestBlock, err := ctx.BlockNode().LatestBlock()
	if err != nil {
		return fmt.Errorf("error while getting latest block: %s", err)
	}
	if parseOldBlocks {
		start := utils.MaxInt64(1, startHeight)
		go enqueuer.NewMissingBlockEnqueuer(start, latestBlock.Height()).ListenAndEnqueueBlocks(ctx, exportQueue)
	}

	return nil
}

// trapSignal will listen for any OS signal and invoke Done on the main
// WaitGroup allowing the main process to gracefully exit.
func trapSignal(ctx *parsecmdtypes.Infrastructures) {
	var sigCh = make(chan os.Signal, 1)

	signal.Notify(sigCh, syscall.SIGTERM)
	signal.Notify(sigCh, syscall.SIGINT)

	go func() {
		sig := <-sigCh
		ctx.Logger.Info("caught signal; shutting down...", "signal", sig.String())
		ctx.Database.Close()
		ctx.Node.Stop()
		defer waitGroup.Done()
	}()
}
