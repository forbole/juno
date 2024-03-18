package enqueuer

import (
	"github.com/forbole/juno/v5/interfaces"
)

var _ interfaces.Enqueuer = &NewBlockEnqueuer{}

type NewBlockEnqueuer struct{}

func NewNewBlockEnqueuer() *NewBlockEnqueuer {
	return &NewBlockEnqueuer{}
}

func (n *NewBlockEnqueuer) ListenAndEnqueueBlocks(ctx interfaces.Context, queue interfaces.BlockQueue) {
	blockCh := ctx.BlockNode().SubscribeBlocks()
	for {
		select {
		case block := <-blockCh:
			queue.Enqueue(block)

		case <-ctx.Done():
			return
		}
	}
}
