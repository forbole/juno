package modules

import (
	"github.com/cosmos/cosmos-sdk/simapp/params"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/rs/zerolog/log"

	"github.com/desmos-labs/juno/client"
	"github.com/desmos-labs/juno/config"
	"github.com/desmos-labs/juno/db"
)

// Registrar represents a modules registrar. This allows to build a list of modules that can later be used by
// specifying their names inside the TOML configuration file.
type Registrar interface {
	BuildModules(*config.Config, *params.EncodingConfig, *sdk.Config, db.Database, *client.Proxy) Modules
}

// ------------------------------------------------------------------------------------------------------------------

// EmptyRegistrar represents a Registrar which does not register any custom module
type EmptyRegistrar struct{}

// BuildModules implements Registrar
func (*EmptyRegistrar) BuildModules(
	*config.Config, *params.EncodingConfig, *sdk.Config, db.Database, *client.Proxy,
) Modules {
	return nil
}

// ------------------------------------------------------------------------------------------------------------------

// GetModules returns the list of module implementations based on the given module names.
// For each module name that is specified but not found, a warning log is printed.
func GetModules(modules Modules, names []string) []Module {
	var modulesImpls []Module
	for _, name := range names {
		module, found := modules.FindByName(name)
		if found {
			modulesImpls = append(modulesImpls, module)
		} else {
			log.Warn().Msgf("%s module is required but not registered. Be sure to register it using registrar.RegisterModule", name)
		}
	}
	return modulesImpls
}
