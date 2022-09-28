package messages

import (
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/forbole/juno/v3/database"
	"github.com/forbole/juno/v3/modules"
	"github.com/forbole/juno/v3/types"
)

var _ modules.Module = &Module{}

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

// HandleMsg implements modules.MessageModule
func (m *Module) HandleMsg(index int, msg sdk.Msg, tx *types.Tx) error {
	return HandleMsg(index, msg, tx, m.cdc, m.db)
}
