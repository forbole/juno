package config

import (
	"fmt"
	"io/ioutil"

	"github.com/BurntSushi/toml"
	"github.com/pkg/errors"
)

// Config defines all necessary juno configuration parameters.
type Config struct {
	RPCNode        string         `toml:"rpc_node"`
	ClientNode     string         `toml:"client_node"`
	DatabaseConfig DatabaseConfig `toml:"database"`
}

// DatabaseConfig represents a generic database configuration
type DatabaseConfig struct {
	Type   string      `toml:"type"`
	Config interface{} `toml:"config"`
}

// MongoDBConfig defines all database connection configuration
// parameters for a MongoDB database
type MongoDBConfig struct {
	Name string `toml:"name"`
	Uri  string `toml:"uri"`
}

// PostgreSQLConfig defines all database connection configuration
// parameters for a PostgreSQL database
type PostgreSQLConfig struct {
	Name     string `toml:"name"`
	Host     string `toml:"host"`
	Port     uint64 `toml:"port"`
	User     string `toml:"user"`
	Password string `toml:"password"`
	SSLMode  string `toml:"ssl_mode"`
}

// ____________________________________________________________

type configToml struct {
	RPCNode    string           `toml:"rpc_node"`
	ClientNode string           `toml:"client_node"`
	DB         databaseInfoToml `toml:"database"`
}

type databaseInfoToml struct {
	Name   string         `toml:"name"`
	Type   string         `toml:"type"`
	Config toml.Primitive `toml:"config"`
}

// ____________________________________________________________

// ParseConfig attempts to read and parse a Juno config from the given file path.
// An error reading or parsing the config results in a panic.
func ParseConfig(configPath string) (*Config, error) {
	if configPath == "" {
		return nil, fmt.Errorf("invalid configuration file")
	}

	configData, err := ioutil.ReadFile(configPath)
	if err != nil {
		return nil, errors.Wrap(err, "failed to read config")
	}

	return ParseConfigString(configData)
}

// ParseConfigString attempts to read and parse a Juno config from the given string bytes.
// An error reading or parsing the config results in a panic.
func ParseConfigString(configData []byte) (*Config, error) {
	var cfg configToml
	md, err := toml.Decode(string(configData), &cfg)
	if err != nil {
		return nil, errors.Wrap(err, "failed to decode config")
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
