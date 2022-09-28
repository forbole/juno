package rawmessages

import (
	"github.com/cosmos/cosmos-sdk/codec"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/forbole/juno/v3/database"
	"github.com/forbole/juno/v3/modules"
	"github.com/forbole/juno/v3/types"
)

var _ modules.Module = &Module{}

// Module represents the module allowing to store rawmessages properly inside a dedicated table
type Module struct {
	cdc codec.Codec
	db  database.Database
}

func NewModule(cdc codec.Codec, db database.Database) *Module {
	return &Module{
		cdc: cdc,
		db:  db,
	}
}

// Name implements modules.Module
func (m *Module) Name() string {
	return "rawmessages"
}

// HandleMsg implements modules.RawMessageModule
func (m *Module) HandleMsg(index int, msg *codectypes.Any, tx *types.Tx) error {
	return HandleMsg(index, msg, tx, m.cdc, m.db)
}
