package node

import (
	"github.com/forbole/juno/v5/interfaces"
	"github.com/forbole/juno/v5/modules/cosmos"
)

type Node interface {
	interfaces.BlockNode

	cosmos.Source

	// Stop defers the node stop execution to the client.
	Stop()
}
