package parser

import (
	"fmt"

	"github.com/forbole/juno/v5/interfaces"
	"github.com/forbole/juno/v5/logging"
	"github.com/forbole/juno/v5/node"
)

var _ interfaces.Worker = &Worker{}

// Worker defines a job consumer that is responsible for getting and
// aggregating block and associated data and exporting it to a database.
type Worker struct {
	index int
}

// NewWorker allows to create a new Worker implementation.
func NewWorker(index int) Worker {
	return Worker{
		index: index,
	}
}

// Start starts a worker by listening for new jobs (block heights) from the
// given worker queue. Any failed job is logged and re-enqueued.
func (w Worker) Start(ctx interfaces.Context, queue interfaces.BlockQueue) {
	logging.WorkerCount.Inc()
	chainID, err := ctx.BlockNode().(node.Node).ChainID()
	if err != nil {
		ctx.Logger().Error("error while getting chain ID from the node ", "err", err)
	}

	for block := range queue.Listen() {
		if err := w.Process(ctx, block); err != nil {
			ctx.Logger().Debug("processing block", "height", block.Height())

			// TODO: Implement exponential backoff or max retries for a block height.
			go func() {
				ctx.Logger().Error("re-enqueueing failed block", "height", block.Height(), "err", err)
				queue.Enqueue(block)
			}()
		}

		ctx.Logger().Info("processed block", "height", block.Height())
		logging.WorkerHeight.WithLabelValues(fmt.Sprintf("%d", w.index), chainID).Set(float64(block.Height()))
	}
}

func (w Worker) Process(ctx interfaces.Context, block interfaces.Block) error {
	// Save the block
	err := ctx.WorkerRepository().SaveBlock(block)
	if err != nil {
		return fmt.Errorf("error while saving block into the database: %s", err)
	}

	for _, module := range ctx.Modules() {
		if m, ok := module.(interfaces.BlockModule); ok {
			if err := m.HandleBlock(block); err != nil {
				return fmt.Errorf("error while handling block with module %s: %s", module.Name(), err)
			}
		}
	}

	return nil
}
