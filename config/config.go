package config

import (
	"fmt"
	"io/ioutil"

	"github.com/BurntSushi/toml"
	"github.com/pkg/errors"
)

// Config defines all necessary juno configuration parameters.
type Config struct {
	RPCNode    string         `toml:"rpc_node"`
	ClientNode string         `toml:"client_node"`
	DB         DatabaseConfig `toml:"database"`
}

// DatabaseConfig defines all database connection configuration parameters.
type DatabaseConfig struct {
	// Common
	Name string `toml:"name"`

	// MongoDB
	Uri string `toml:"uri"`

	// PostgreSQL
	Host     string `toml:"host"`
	Port     uint64 `toml:"port"`
	User     string `toml:"user"`
	Password string `toml:"password"`
	SSLMode  string `toml:"ssl_mode"`
}

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

	var cfg Config
	if _, err := toml.Decode(string(configData), &cfg); err != nil {
		return nil, errors.Wrap(err, "failed to decode config")
	}

	return &cfg, nil
}
