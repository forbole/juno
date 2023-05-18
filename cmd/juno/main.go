package main

import (
	"os"

	"github.com/emrahm/juno/v5/cmd/parse/types"

	"github.com/emrahm/juno/v5/modules/messages"
	"github.com/emrahm/juno/v5/modules/registrar"

	"github.com/emrahm/juno/v5/cmd"
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
