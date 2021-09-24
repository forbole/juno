package parser

import (
	"github.com/cosmos/cosmos-sdk/codec"

	"github.com/desmos-labs/juno/logging"
	"github.com/desmos-labs/juno/node"

	"github.com/desmos-labs/juno/database"
	"github.com/desmos-labs/juno/modules"
	"github.com/desmos-labs/juno/types"
)

// Context represents the context that is shared among different workers
type Context struct {
	Codec    codec.BinaryMarshaler
	Node     node.Node
	Database database.Database
	Logger   logging.Logger

	Queue   types.HeightQueue
	Modules []modules.Module
}

// NewContext allows to build a new Worker Context instance
func NewContext(
	codec codec.BinaryMarshaler, queue types.HeightQueue,
	node node.Node, db database.Database, logger logging.Logger,
	modules []modules.Module,
) *Context {
	return &Context{
		Codec:    codec,
		Node:     node,
		Database: db,
		Logger:   logger,

		Queue:   queue,
		Modules: modules,
	}
}
