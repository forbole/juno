package parser

import (
	"github.com/forbole/juno/v6/database"
	"github.com/forbole/juno/v6/logging"
	"github.com/forbole/juno/v6/modules"
	"github.com/forbole/juno/v6/node"
)

// Context represents the context that is shared among different workers
type Context struct {
	Node        node.Node
	Database    database.Database
	Logger      logging.Logger
	Modules     []modules.Module
	SyncManager *ModuleSyncManager
}

// NewContext builds a new Context instance
func NewContext(
	proxy node.Node, db database.Database,
	logger logging.Logger, modules []modules.Module,
) *Context {
	return &Context{
		Node:        proxy,
		Database:    db,
		Modules:     modules,
		Logger:      logger,
		SyncManager: NewModuleSyncManager(db),
	}
}
