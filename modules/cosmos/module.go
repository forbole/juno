package cosmos

import (
	codec "github.com/cosmos/cosmos-sdk/codec"

	"github.com/forbole/juno/v5/interfaces"
)

var _ interfaces.Module = &Module{}

type Module struct {
	source          Source
	db              CosmosRepository
	codec           codec.Codec
	logger          interfaces.Logger
	parseGenesis    bool
	genesisFilePath string
	modules         []interfaces.Module
}

func NewModule(
	source Source,
	db CosmosRepository,
	codec codec.Codec,
	logger interfaces.Logger,
	parseGenesis bool,
	genesisFilePath string,
	subModules ...interfaces.Module,
) *Module {
	return &Module{
		source:  source,
		db:      db,
		codec:   codec,
		logger:  logger,
		modules: subModules,
	}
}

func (m *Module) Name() string {
	return "cosmos"
}
