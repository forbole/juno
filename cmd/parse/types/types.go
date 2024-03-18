package types

import (
	"github.com/cosmos/cosmos-sdk/std"

	"github.com/forbole/juno/v5/logging"
	"github.com/forbole/juno/v5/types/config"
	"github.com/forbole/juno/v5/types/params"

	"github.com/forbole/juno/v5/interfaces"
	"github.com/forbole/juno/v5/modules/registrar"
)

// Config contains all the configuration for the "parse" command
type Config struct {
	registrar             registrar.Registrar
	configParser          config.Parser
	encodingConfigBuilder EncodingConfigBuilder
	setupCfg              SdkConfigSetup
	logger                interfaces.Logger
}

// NewConfig allows to build a new Config instance
func NewConfig() *Config {
	return &Config{}
}

// WithRegistrar sets the modules registrar to be used
func (cfg *Config) WithRegistrar(r registrar.Registrar) *Config {
	cfg.registrar = r
	return cfg
}

// GetRegistrar returns the modules registrar to be used
func (cfg *Config) GetRegistrar() registrar.Registrar {
	if cfg.registrar == nil {
		return &registrar.EmptyRegistrar{}
	}
	return cfg.registrar
}

// WithConfigParser sets the configuration parser to be used
func (cfg *Config) WithConfigParser(p config.Parser) *Config {
	cfg.configParser = p
	return cfg
}

// GetConfigParser returns the configuration parser to be used
func (cfg *Config) GetConfigParser() config.Parser {
	if cfg.configParser == nil {
		return config.DefaultConfigParser
	}
	return cfg.configParser
}

// WithEncodingConfigBuilder sets the configurations builder to be used
func (cfg *Config) WithEncodingConfigBuilder(b EncodingConfigBuilder) *Config {
	cfg.encodingConfigBuilder = b
	return cfg
}

// GetEncodingConfigBuilder returns the encoding config builder to be used
func (cfg *Config) GetEncodingConfigBuilder() EncodingConfigBuilder {
	if cfg.encodingConfigBuilder == nil {
		return func() params.EncodingConfig {
			encodingConfig := params.MakeTestEncodingConfig()
			std.RegisterLegacyAminoCodec(encodingConfig.Amino)
			std.RegisterInterfaces(encodingConfig.InterfaceRegistry)
			ModuleBasics.RegisterLegacyAminoCodec(encodingConfig.Amino)
			ModuleBasics.RegisterInterfaces(encodingConfig.InterfaceRegistry)
			return encodingConfig
		}
	}
	return cfg.encodingConfigBuilder
}

// WithSetupConfig sets the SDK setup configurator to be used
func (cfg *Config) WithSetupConfig(s SdkConfigSetup) *Config {
	cfg.setupCfg = s
	return cfg
}

// GetSetupConfig returns the SDK configuration builder to use
func (cfg *Config) GetSetupConfig() SdkConfigSetup {
	if cfg.setupCfg == nil {
		return DefaultConfigSetup
	}
	return cfg.setupCfg
}

// WithLogger sets the logger to be used while parsing the data
func (cfg *Config) WithLogger(logger interfaces.Logger) *Config {
	cfg.logger = logger
	return cfg
}

// GetLogger returns the logger to be used when parsing the data
func (cfg *Config) GetLogger() interfaces.Logger {
	if cfg.logger == nil {
		return logging.DefaultLogger()
	}
	return cfg.logger
}
