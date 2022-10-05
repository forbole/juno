package messages

import (
	"github.com/cosmos/cosmos-sdk/codec"

	"github.com/forbole/juno/v3/database"
	"github.com/forbole/juno/v3/modules"
)

var (
	_ modules.Module           = &Module{}
	_ modules.RawMessageModule = &Module{}
)

// Module represents the module allowing to store messages properly inside a dedicated table
type Module struct {
	cdc codec.Codec
	db  database.Database
}

func NewModule(
	cdc codec.Codec, db database.Database) *Module {
	return &Module{
		cdc: cdc,
		db:  db,
	}
}

// Name implements modules.Module
func (m *Module) Name() string {
	return "messages"
}
