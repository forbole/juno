package registrar

import (
	"github.com/cosmos/cosmos-sdk/simapp/params"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/saifullah619/juno/v3/node"

	"github.com/saifullah619/juno/v3/modules/telemetry"

	"github.com/saifullah619/juno/v3/logging"

	"github.com/saifullah619/juno/v3/types/config"

	"github.com/saifullah619/juno/v3/modules/pruning"

	"github.com/saifullah619/juno/v3/modules"
	"github.com/saifullah619/juno/v3/modules/messages"

	"github.com/saifullah619/juno/v3/database"
)

// Context represents the context of the modules registrar
type Context struct {
	JunoConfig     config.Config
	SDKConfig      *sdk.Config
	EncodingConfig *params.EncodingConfig
	Database       database.Database
	Proxy          node.Node
	Logger         logging.Logger
}

// NewContext allows to build a new Context instance
func NewContext(
	parsingConfig config.Config, sdkConfig *sdk.Config, encodingConfig *params.EncodingConfig,
	database database.Database, proxy node.Node, logger logging.Logger,
) Context {
	return Context{
		JunoConfig:     parsingConfig,
		SDKConfig:      sdkConfig,
		EncodingConfig: encodingConfig,
		Database:       database,
		Proxy:          proxy,
		Logger:         logger,
	}
}

// Registrar represents a modules registrar. This allows to build a list of modules that can later be used by
// specifying their names inside the TOML configuration file.
type Registrar interface {
	BuildModules(context Context) modules.Modules
}

// ------------------------------------------------------------------------------------------------------------------

var (
	_ Registrar = &EmptyRegistrar{}
)

// EmptyRegistrar represents a Registrar which does not register any custom module
type EmptyRegistrar struct{}

// BuildModules implements Registrar
func (*EmptyRegistrar) BuildModules(_ Context) modules.Modules {
	return nil
}

// ------------------------------------------------------------------------------------------------------------------

var (
	_ Registrar = &DefaultRegistrar{}
)

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
func (r *DefaultRegistrar) BuildModules(ctx Context) modules.Modules {
	return modules.Modules{
		pruning.NewModule(ctx.JunoConfig, ctx.Database, ctx.Logger),
		messages.NewModule(r.parser, ctx.EncodingConfig.Codec, ctx.Database),
		telemetry.NewModule(ctx.JunoConfig),
	}
}

// ------------------------------------------------------------------------------------------------------------------

// GetModules returns the list of module implementations based on the given module names.
// For each module name that is specified but not found, a warning log is printed.
func GetModules(mods modules.Modules, names []string, logger logging.Logger) []modules.Module {
	var modulesImpls []modules.Module
	for _, name := range names {
		module, found := mods.FindByName(name)
		if found {
			modulesImpls = append(modulesImpls, module)
		} else {
			logger.Error("Module is required but not registered. Be sure to register it using registrar.RegisterModule", "module", name)
		}
	}
	return modulesImpls
}
