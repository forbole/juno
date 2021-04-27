package main

import (
	"os"

	"github.com/desmos-labs/juno/types"

	"github.com/desmos-labs/juno/modules/messages"
	"github.com/desmos-labs/juno/modules/registrar"

	"github.com/cosmos/cosmos-sdk/simapp"

	"github.com/desmos-labs/juno/cmd"
	stddb "github.com/desmos-labs/juno/db/builder"
)

func main() {
	// Build the exec
	exec := cmd.BuildDefaultExecutor(
		"juno",
		registrar.NewDefaultRegistrar(
			messages.CosmosMessageAddressesParser,
		),
		types.DefaultConfigParser,
		types.DefaultConfigSetup,
		simapp.MakeTestEncodingConfig,
		stddb.Builder,
	)

	// Run the commands and panic on any error
	err := exec.Execute()
	if err != nil {
		os.Exit(1)
	}
}
