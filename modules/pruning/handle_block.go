package pruning

import (
	"fmt"

	tmctypes "github.com/tendermint/tendermint/rpc/core/types"

	"github.com/saifullah619/juno/v3/database"
	"github.com/saifullah619/juno/v3/types"
)

// HandleBlock implements modules.BlockModule
func (m *Module) HandleBlock(
	block *tmctypes.ResultBlock, _ *tmctypes.ResultBlockResults, _ []*types.Tx, _ *tmctypes.ResultValidators,
) error {
	if block.Block.Height%m.cfg.Interval != 0 {
		// Not an interval height, so just skip
		return nil
	}

	pruningDb, ok := m.db.(database.PruningDb)
	if !ok {
		return fmt.Errorf("pruning is enabled, but your database does not implement PruningDb")
	}

	// Get last pruned height
	var height, err = pruningDb.GetLastPruned()
	if err != nil {
		return err
	}

	// Iterate from last pruned height until (current block height - keep recent) to
	// avoid pruning the recent blocks that should be kept
	for ; height < block.Block.Height-m.cfg.KeepRecent; height++ {

		if height%m.cfg.KeepRecent == 0 {
			// The height should be kept, so just skip
			continue
		}

		// Prune the height
		m.logger.Debug("pruning", "module", "pruning", "height", height)
		err = pruningDb.Prune(height)
		if err != nil {
			return fmt.Errorf("error while pruning height %d: %s", height, err.Error())
		}
	}

	return pruningDb.StoreLastPruned(height)
}
