package config

import (
	"fmt"
	"io/ioutil"

	"github.com/BurntSushi/toml"
	"github.com/rs/zerolog/log"
)

// SetupConfig takes the path to a configuration file and returns the properly parsed configuration
func Read(configPath string) (*Config, error) {
	if configPath == "" {
		return nil, fmt.Errorf("empty configuration path")
	}

	log.Debug().Msg("reading config file")

	configData, err := ioutil.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config: %s", err)
	}

	return ParseString(configData)
}

// ParseString attempts to read and parse a Juno config from the given string bytes.
// An error reading or parsing the config results in a panic.
func ParseString(configData []byte) (*Config, error) {
	var cfg configToml
	md, err := toml.Decode(string(configData), &cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to decode config: %s", err)
	}

	var config interface{}
	switch cfg.DB.Type {
	case "mongodb":
		config = new(MongoDBConfig)
	case "postgresql":
		config = new(PostgreSQLConfig)
	default:
		return nil, fmt.Errorf("unknown type %q", cfg.DB.Type)
	}
	if err := md.PrimitiveDecode(cfg.DB.Config, config); err != nil {
		return nil, err
	}

	return &Config{
		RPCNode:    cfg.RPCNode,
		ClientNode: cfg.ClientNode,
		DatabaseConfig: DatabaseConfig{
			Type:   cfg.DB.Type,
			Config: config,
		},
	}, nil
}
