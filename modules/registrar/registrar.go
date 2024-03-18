package registrar

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/forbole/juno/v5/interfaces"
	"github.com/forbole/juno/v5/node"
	"github.com/forbole/juno/v5/types/params"

	"github.com/forbole/juno/v5/modules/telemetry"

	"github.com/forbole/juno/v5/types/config"

	"github.com/forbole/juno/v5/modules/pruning"

	"github.com/forbole/juno/v5/modules/messages"

	"github.com/forbole/juno/v5/modules/cosmos"

	"github.com/forbole/juno/v5/database/postgresql"
)

// Context represents the context of the modules registrar
type Context struct {
	JunoConfig     config.Config
	SDKConfig      *sdk.Config
	EncodingConfig params.EncodingConfig
	Database       *postgresql.Database
	Proxy          node.Node
	Logger         interfaces.Logger
}

// NewContext allows to build a new Context instance
func NewContext(
	parsingConfig config.Config, sdkConfig *sdk.Config, encodingConfig params.EncodingConfig,
	database *postgresql.Database, proxy node.Node, logger interfaces.Logger,
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
	BuildModules(context Context) interfaces.Modules
}

// ------------------------------------------------------------------------------------------------------------------

var (
	_ Registrar = &EmptyRegistrar{}
)

// EmptyRegistrar represents a Registrar which does not register any custom module
type EmptyRegistrar struct{}

// BuildModules implements Registrar
func (*EmptyRegistrar) BuildModules(_ Context) interfaces.Modules {
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
func (r *DefaultRegistrar) BuildModules(ctx Context) interfaces.Modules {
	modules := interfaces.Modules{
		pruning.NewModule(ctx.JunoConfig, ctx.Database, ctx.Logger),
		messages.NewModule(r.parser, ctx.EncodingConfig.Codec, ctx.Database),
		telemetry.NewModule(ctx.JunoConfig),
	}

	cosmosModule := cosmos.NewModule(ctx.Proxy, ctx.Database, ctx.EncodingConfig.Codec, ctx.Logger, ctx.JunoConfig.Parser.ParseGenesis, ctx.JunoConfig.Parser.GenesisFilePath, modules...)
	return append(modules, cosmosModule)
}

// ------------------------------------------------------------------------------------------------------------------

// GetModules returns the list of module implementations based on the given module names.
// For each module name that is specified but not found, a warning log is printed.
func GetModules(mods interfaces.Modules, names []string, logger interfaces.Logger) interfaces.Modules {
	var modulesImpls []interfaces.Module
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
