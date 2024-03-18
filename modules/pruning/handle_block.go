package pruning

import (
	"fmt"

	"github.com/forbole/juno/v5/interfaces"
)

var _ interfaces.BlockModule = &Module{}

// HandleBlock implements modules.BlockModule
func (m *Module) HandleBlock(
	block interfaces.Block,
) error {
	if block.Height()%m.cfg.Interval != 0 {
		// Not an interval height, so just skip
		return nil
	}

	// Get last pruned height
	var height, err = m.db.GetLastPruned()
	if err != nil {
		return err
	}

	// Iterate from last pruned height until (current block height - keep recent) to
	// avoid pruning the recent blocks that should be kept
	for ; height < block.Height()-m.cfg.KeepRecent; height++ {

		if height%m.cfg.KeepRecent == 0 {
			// The height should be kept, so just skip
			continue
		}

		// Prune the height
		m.logger.Debug("pruning", "module", "pruning", "height", height)
		err = m.db.Prune(height)
		if err != nil {
			return fmt.Errorf("error while pruning height %d: %s", height, err.Error())
		}
	}

	return m.db.StoreLastPruned(height)
}
