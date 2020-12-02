package config

import (
	"github.com/BurntSushi/toml"
)

// Config defines all necessary juno configuration parameters.
type Config struct {
	RPCNode        string         `toml:"rpc_node"`
	ClientNode     string         `toml:"client_node"`
	Modules        []string       `toml:"modules"`
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
