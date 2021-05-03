package registrar

import (
	"github.com/cosmos/cosmos-sdk/simapp/params"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/rs/zerolog/log"

	"github.com/desmos-labs/juno/types"

	"github.com/desmos-labs/juno/modules/pruning"

	"github.com/desmos-labs/juno/modules"
	"github.com/desmos-labs/juno/modules/messages"

	"github.com/desmos-labs/juno/client"
	"github.com/desmos-labs/juno/db"
)

// Registrar represents a modules registrar. This allows to build a list of modules that can later be used by
// specifying their names inside the TOML configuration file.
type Registrar interface {
	BuildModules(types.Config, *params.EncodingConfig, *sdk.Config, db.Database, *client.Proxy) modules.Modules
}

// ------------------------------------------------------------------------------------------------------------------

var _ Registrar = &EmptyRegistrar{}

// EmptyRegistrar represents a Registrar which does not register any custom module
type EmptyRegistrar struct{}

// BuildModules implements Registrar
func (*EmptyRegistrar) BuildModules(
	types.Config, *params.EncodingConfig, *sdk.Config, db.Database, *client.Proxy,
) modules.Modules {
	return nil
}

// ------------------------------------------------------------------------------------------------------------------

// DefaultRegistrar represents a registrar that allows to handle the default Juno modules
type DefaultRegistrar struct {
	parser messages.MessageAddressesParser
}

// NewDefaultRegistrar builds a new DefaultRegistrar
func NewDefaultRegistrar(parser messages.MessageAddressesParser) *DefaultRegistrar {
	return &DefaultRegistrar{
		parser: parser,
	}
}

// BuildModules implements Registrar
func (r *DefaultRegistrar) BuildModules(
	cfg types.Config, encodingCfg *params.EncodingConfig, _ *sdk.Config, db db.Database, _ *client.Proxy,
) modules.Modules {
	return modules.Modules{
		pruning.NewModule(cfg.GetPruningConfig(), db),
		messages.NewModule(r.parser, encodingCfg.Marshaler, db),
	}
}

// ------------------------------------------------------------------------------------------------------------------

// GetModules returns the list of module implementations based on the given module names.
// For each module name that is specified but not found, a warning log is printed.
func GetModules(mods modules.Modules, names []string) []modules.Module {
	var modulesImpls []modules.Module
	for _, name := range names {
		module, found := mods.FindByName(name)
		if found {
			modulesImpls = append(modulesImpls, module)
		} else {
			log.Warn().Msgf("%s module is required but not registered. Be sure to register it using registrar.RegisterModule", name)
		}
	}
	return modulesImpls
}
