package main

import (
	"os"

	"github.com/saifullah619/juno/v3/cmd/parse/types"

	"github.com/saifullah619/juno/v3/modules/messages"
	"github.com/saifullah619/juno/v3/modules/registrar"

	"github.com/saifullah619/juno/v3/cmd"
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
