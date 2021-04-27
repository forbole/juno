package parse

import (
	"github.com/cosmos/cosmos-sdk/simapp"
	"github.com/cosmos/cosmos-sdk/simapp/params"

	"github.com/desmos-labs/juno/client"
	"github.com/desmos-labs/juno/db"
	"github.com/desmos-labs/juno/db/builder"
	"github.com/desmos-labs/juno/modules"
	"github.com/desmos-labs/juno/modules/registrar"
	"github.com/desmos-labs/juno/types"
)

// Config contains all the configuration for the "parse" command
type Config struct {
	name                  string
	registrar             registrar.Registrar
	configParser          types.ConfigParser
	encodingConfigBuilder types.EncodingConfigBuilder
	setupCfg              types.SdkConfigSetup
	buildDb               db.Builder
}

// NewConfig allows to build a new Config instance
func NewConfig(name string) *Config {
	return &Config{
		name: name,
	}
}

// GetName returns the name of the command
func (config *Config) GetName() string {
	return config.name
}

// WithRegistrar sets the modules registrar to be used
func (config *Config) WithRegistrar(r registrar.Registrar) *Config {
	config.registrar = r
	return config
}

// GetRegistrar returns the modules registrar to be used
func (config *Config) GetRegistrar() registrar.Registrar {
	if config.registrar == nil {
		return &registrar.EmptyRegistrar{}
	}
	return config.registrar
}

// WithConfigParser sets the configuration parser to be used
func (config *Config) WithConfigParser(p types.ConfigParser) *Config {
	config.configParser = p
	return config
}

// GetConfigParser returns the configuration parser to be used
func (config *Config) GetConfigParser() types.ConfigParser {
	if config.configParser == nil {
		return types.DefaultConfigParser
	}
	return config.configParser
}

// WithEncodingConfigBuilder sets the configurations builder to be used
func (config *Config) WithEncodingConfigBuilder(b types.EncodingConfigBuilder) *Config {
	config.encodingConfigBuilder = b
	return config
}

// GetEncodingConfigBuilder returns the encoding config builder to be used
func (config *Config) GetEncodingConfigBuilder() types.EncodingConfigBuilder {
	if config.encodingConfigBuilder == nil {
		return simapp.MakeTestEncodingConfig
	}
	return config.encodingConfigBuilder
}

// WithSetupConfig sets the SDK setup configurator to be used
func (config *Config) WithSetupConfig(s types.SdkConfigSetup) *Config {
	config.setupCfg = s
	return config
}

// GetSetupConfig returns the SDK configuration builder to use
func (config *Config) GetSetupConfig() types.SdkConfigSetup {
	if config.setupCfg == nil {
		return types.DefaultConfigSetup
	}
	return config.setupCfg
}

// WithDBBuilder sets the database builder to be used
func (config *Config) WithDBBuilder(b db.Builder) *Config {
	config.buildDb = b
	return config
}

// GetDBBuilder returns the database builder to be used
func (config *Config) GetDBBuilder() db.Builder {
	if config.buildDb == nil {
		return builder.Builder
	}
	return config.buildDb
}

// --------------------------------------------------------------------------------------------------------------------

// ParserData contains the data that should be used to start parsing the chain
type ParserData struct {
	EncodingConfig *params.EncodingConfig
	Proxy          *client.Proxy
	Database       db.Database
	Modules        []modules.Module
}

// NewParserData builds a new ParserData instance
func NewParserData(
	encodingConfig *params.EncodingConfig, proxy *client.Proxy, db db.Database, modules []modules.Module,
) *ParserData {
	return &ParserData{
		EncodingConfig: encodingConfig,
		Proxy:          proxy,
		Database:       db,
		Modules:        modules,
	}
}
