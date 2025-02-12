package main

import (
	"os"

	"github.com/0xPellNetwork/juno/v6/cmd/parse/types"

	"github.com/0xPellNetwork/juno/v6/modules/messages"
	"github.com/0xPellNetwork/juno/v6/modules/registrar"

	"github.com/0xPellNetwork/juno/v6/cmd"
)

func main() {
	// JunoConfig the runner
	config := cmd.NewConfig("juno").
		WithParseConfig(types.NewConfig().
			WithRegistrar(registrar.NewDefaultRegistrar(
				messages.CosmosMessageAddressesParser,
			)),
		)

	// Run the commands and panic on any error
	exec := cmd.BuildDefaultExecutor(config)
	err := exec.Execute()
	if err != nil {
		os.Exit(1)
	}
}
