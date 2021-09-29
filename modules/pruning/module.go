package pruning

import (
	"fmt"

	"github.com/desmos-labs/juno/v2/types/config"

	"github.com/desmos-labs/juno/v2/logging"

	tmctypes "github.com/tendermint/tendermint/rpc/core/types"

	"github.com/desmos-labs/juno/v2/database"
	"github.com/desmos-labs/juno/v2/modules"
	"github.com/desmos-labs/juno/v2/types"
)

var (
	_ modules.Module                     = &Module{}
	_ modules.BlockModule                = &Module{}
	_ modules.AdditionalOperationsModule = &Module{}
)

// Module represents the pruning module allowing to clean the database periodically
type Module struct {
	cfg    *Config
	db     database.Database
	logger logging.Logger
}

// NewModule builds a new Module instance
func NewModule(cfg config.Config, db database.Database, logger logging.Logger) *Module {
	pruningCfg, err := ParseConfig(cfg.GetBytes())
	if err != nil {
		panic(err)
	}

	return &Module{
		cfg:    pruningCfg,
		db:     db,
		logger: logger,
	}
}

// Name implements modules.Module
func (m *Module) Name() string {
	return "pruning"
}

// RunAdditionalOperations implements
func (m *Module) RunAdditionalOperations() error {
	return RunAdditionalOperations(m.cfg)
}

// HandleBlock implements modules.BlockModule
func (m *Module) HandleBlock(block *tmctypes.ResultBlock, _ []*types.Tx, _ *tmctypes.ResultValidators) error {
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
