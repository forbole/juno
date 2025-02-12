package messages

import (
	"github.com/0xPellNetwork/juno/v6/database"
	"github.com/0xPellNetwork/juno/v6/modules"
	"github.com/0xPellNetwork/juno/v6/types"
)

var _ modules.Module = &Module{}

// Module represents the module allowing to store messages properly inside a dedicated table
type Module struct {
	parser MessageAddressesParser

	db database.Database
}

func NewModule(parser MessageAddressesParser, db database.Database) *Module {
	return &Module{
		parser: parser,
		db:     db,
	}
}

// Name implements modules.Module
func (m *Module) Name() string {
	return "messages"
}

// HandleMsg implements modules.MessageModule
func (m *Module) HandleMsg(index int, msg types.Message, tx *types.Transaction) error {
	return HandleMsg(index, msg, tx, m.parser, m.db)
}
