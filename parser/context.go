package parser

import (
	"golang.org/x/net/context"

	"github.com/forbole/juno/v5/interfaces"
)

var _ interfaces.Context = &Context{}

// Context represents the context that is shared among different workers
type Context struct {
	context.Context
	node     interfaces.BlockNode
	database interfaces.WorkerRepository
	logger   interfaces.Logger
	modules  []interfaces.Module
}

func (c *Context) WorkerRepository() interfaces.WorkerRepository {
	return c.database
}

func (c *Context) BlockNode() interfaces.BlockNode {
	return c.node
}

func (c *Context) Modules() []interfaces.Module {
	return c.modules
}

func (c *Context) Logger() interfaces.Logger {
	return c.logger
}

// NewContext builds a new Context instance
func NewContext(
	ctx context.Context, proxy interfaces.BlockNode, db interfaces.WorkerRepository,
	logger interfaces.Logger, modules []interfaces.Module,
) *Context {
	return &Context{
		Context:  ctx,
		node:     proxy,
		database: db,
		modules:  modules,
		logger:   logger,
	}
}
