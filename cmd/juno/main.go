package main

import (
	"github.com/cosmos/cosmos-sdk/simapp"
	dbbuilder "github.com/desmos-labs/juno/db/builder"
	"github.com/desmos-labs/juno/executor"
	"github.com/desmos-labs/juno/types"
)

func main() {
	// Build the exec
	exec := executor.BuildDefaultExecutor("juno", types.EmptySetup, simapp.MakeCodec, dbbuilder.DatabaseBuilder)

	// Run the commands and panic on any error
	err := exec.Execute()
	if err != nil {
		panic(err)
	}
}
