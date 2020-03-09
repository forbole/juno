package main

import (
	"github.com/cosmos/cosmos-sdk/simapp"
	"github.com/desmos-labs/juno/db/postgresql"
	"github.com/desmos-labs/juno/executor"
	"github.com/desmos-labs/juno/types"
)

func main() {
	// Build the exec
	exec := executor.BuildDefaultExecutor("juno", types.EmptySetup, simapp.MakeCodec, postgresql.Builder)

	// Run the commands and panic on any error
	err := exec.Execute()
	if err != nil {
		panic(err)
	}
}
