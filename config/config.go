package config

// Config defines all necessary juno configuration parameters.
type Config struct {
	RPCConfig      *RPCConfig
	APIConfig      *APIConfig
	GrpcConfig     *GrpcConfig
	CosmosConfig   *CosmosConfig
	DatabaseConfig *DatabaseConfig
}

// NewConfig builds a new Config instance
func NewConfig(
	rpcConfig *RPCConfig, apiConfig *APIConfig, grpConfig *GrpcConfig,
	cosmosConfig *CosmosConfig, dbConfig *DatabaseConfig,
) *Config {
	return &Config{
		RPCConfig:      rpcConfig,
		APIConfig:      apiConfig,
		GrpcConfig:     grpConfig,
		CosmosConfig:   cosmosConfig,
		DatabaseConfig: dbConfig,
	}
}

// GrpcConfig contains the configuration of the gRPC endpoint
type GrpcConfig struct {
	Address  string `toml:"address"`
	Insecure bool   `toml:"insecure"`
}

// RPCConfig contains the configuration of the RPC endpoint
type RPCConfig struct {
	Address string `toml:"address"`
}

// APIConfig contains the configuration of the REST API endpoint
type APIConfig struct {
	Address string `toml:"address"`
}

// CosmosConfig contains the data to configure the CosmosConfig SDK
type CosmosConfig struct {
	Prefix  string   `toml:"prefix"`
	Modules []string `toml:"modules"`
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
	Schema   string `toml:"schema"`
}
