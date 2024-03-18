package telemetry

import (
	"github.com/forbole/juno/v5/interfaces"
	"github.com/forbole/juno/v5/types/config"
)

const (
	ModuleName = "telemetry"
)

var (
	_ interfaces.Module                     = &Module{}
	_ interfaces.AdditionalOperationsModule = &Module{}
)

// Module represents the telemetry module
type Module struct {
	cfg *Config
}

// NewModule returns a new Module implementation
func NewModule(cfg config.Config) *Module {
	bz, err := cfg.GetBytes()
	if err != nil {
		panic(err)
	}

	telemetryCfg, err := ParseConfig(bz)
	if err != nil {
		panic(err)
	}

	return &Module{
		cfg: telemetryCfg,
	}
}

// Name implements modules.Module
func (m *Module) Name() string {
	return ModuleName
}

// RunAdditionalOperations implements modules.AdditionalOperationsModule
func (m *Module) RunAdditionalOperations() error {
	return RunAdditionalOperations(m.cfg)
}
