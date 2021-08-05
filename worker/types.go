package worker

import (
	"github.com/cosmos/cosmos-sdk/simapp/params"

	"github.com/desmos-labs/juno/types/logging"

	"github.com/desmos-labs/juno/client"
	"github.com/desmos-labs/juno/db"
	"github.com/desmos-labs/juno/modules"
	"github.com/desmos-labs/juno/types"
)

// Context represents the context that is shared among different workers
type Context struct {
	EncodingConfig *params.EncodingConfig
	ClientProxy    *client.Proxy
	Database       db.Database
	Logger         logging.Logger

	Queue   types.HeightQueue
	Modules []modules.Module
}

// NewContext allows to build a new Worker Context instance
func NewContext(
	encodingConfig *params.EncodingConfig,
	clientProxy *client.Proxy,
	db db.Database,
	logger logging.Logger,
	queue types.HeightQueue,
	modules []modules.Module,
) *Context {
	return &Context{
		EncodingConfig: encodingConfig,
		ClientProxy:    clientProxy,
		Database:       db,
		Logger:         logger,

		Queue:   queue,
		Modules: modules,
	}
}
