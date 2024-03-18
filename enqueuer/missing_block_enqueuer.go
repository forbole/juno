package enqueuer

import (
	"time"

	"github.com/forbole/juno/v5/interfaces"
)

var _ interfaces.Enqueuer = &MissingBlockEnqueuer{}

type MissingBlockEnqueuer struct {
	start int64
	end   int64
}

func NewMissingBlockEnqueuer(start int64, end int64) *MissingBlockEnqueuer {
	return &MissingBlockEnqueuer{
		start: start,
		end:   end,
	}
}

func (m *MissingBlockEnqueuer) ListenAndEnqueueBlocks(ctx interfaces.Context, queue interfaces.BlockQueue) {
	for current := m.start; current <= m.end; {
		select {
		case <-ctx.Done():
			return

		default:
			block, err := ctx.BlockNode().Block(current)
			if err != nil {
				ctx.Logger().Error("failed to get block", "height", current, "err", err)
				time.Sleep(2 * time.Second)
				continue
			}
			queue.Enqueue(block)
			current += 1
		}
	}
}
