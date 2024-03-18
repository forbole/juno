package pruning

import (
	"github.com/forbole/juno/v5/types/config"

	"github.com/forbole/juno/v5/interfaces"
)

var (
	_ interfaces.Module                     = &Module{}
	_ interfaces.AdditionalOperationsModule = &Module{}
)

// Module represents the pruning module allowing to clean the database periodically
type Module struct {
	cfg    *Config
	db     PruningRepository
	logger interfaces.Logger
}

// NewModule builds a new Module instance
func NewModule(cfg config.Config, db PruningRepository, logger interfaces.Logger) *Module {
	bz, err := cfg.GetBytes()
	if err != nil {
		panic(err)
	}

	pruningCfg, err := ParseConfig(bz)
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
