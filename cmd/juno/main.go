package main

import (
	"os"

	"github.com/desmos-labs/juno/cmd/parse"

	"github.com/desmos-labs/juno/modules/messages"
	"github.com/desmos-labs/juno/modules/registrar"

	"github.com/desmos-labs/juno/cmd"
)

func main() {
	// Config the runner
	config := cmd.NewConfig("juno").
		WithParseConfig(parse.NewConfig().
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
