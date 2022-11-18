package types

import (
	"github.com/cosmos/cosmos-sdk/simapp"

	"github.com/saifullah619/juno/v3/logging"
	"github.com/saifullah619/juno/v3/types/config"

	"github.com/saifullah619/juno/v3/database"
	"github.com/saifullah619/juno/v3/database/builder"
	"github.com/saifullah619/juno/v3/modules/registrar"
)

// Config contains all the configuration for the "parse" command
type Config struct {
	registrar             registrar.Registrar
	configParser          config.Parser
	encodingConfigBuilder EncodingConfigBuilder
	setupCfg              SdkConfigSetup
	buildDb               database.Builder
	logger                logging.Logger
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
		return simapp.MakeTestEncodingConfig
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

// WithDBBuilder sets the database builder to be used
func (cfg *Config) WithDBBuilder(b database.Builder) *Config {
	cfg.buildDb = b
	return cfg
}

// GetDBBuilder returns the database builder to be used
func (cfg *Config) GetDBBuilder() database.Builder {
	if cfg.buildDb == nil {
		return builder.Builder
	}
	return cfg.buildDb
}

// WithLogger sets the logger to be used while parsing the data
func (cfg *Config) WithLogger(logger logging.Logger) *Config {
	cfg.logger = logger
	return cfg
}

// GetLogger returns the logger to be used when parsing the data
func (cfg *Config) GetLogger() logging.Logger {
	if cfg.logger == nil {
		return logging.DefaultLogger()
	}
	return cfg.logger
}
