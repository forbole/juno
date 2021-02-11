package config

var (
	// Cfg represents the configuration to be used during the execution
	Cfg *Config
)

// Config defines all necessary juno configuration parameters.
type Config struct {
	RPC      *RPCConfig      `toml:"rpc"`
	Grpc     *GrpcConfig     `toml:"grpc"`
	Cosmos   *CosmosConfig   `toml:"cosmos"`
	Database *DatabaseConfig `toml:"database"`
	Logging  *LoggingConfig  `toml:"logging"`
	Parsing  *ParsingConfig  `toml:"parsing"`
}

// NewConfig builds a new Config instance
func NewConfig(
	rpcConfig *RPCConfig, grpConfig *GrpcConfig,
	cosmosConfig *CosmosConfig, dbConfig *DatabaseConfig,
	loggingConfig *LoggingConfig, parsingConfig *ParsingConfig,
) *Config {
	return &Config{
		RPC:      rpcConfig,
		Grpc:     grpConfig,
		Cosmos:   cosmosConfig,
		Database: dbConfig,
		Logging:  loggingConfig,
		Parsing:  parsingConfig,
	}
}

// GrpcConfig contains the configuration of the gRPC endpoint
type GrpcConfig struct {
	Address  string `toml:"address"`
	Insecure bool   `toml:"insecure"`
}

func NewGrpcConfig(address string, insecure bool) *GrpcConfig {
	return &GrpcConfig{
		Address:  address,
		Insecure: insecure,
	}
}

// RPCConfig contains the configuration of the RPC endpoint
type RPCConfig struct {
	Address string `toml:"address"`
}

func NewRPCConfig(address string) *RPCConfig {
	return &RPCConfig{
		Address: address,
	}
}

// CosmosConfig contains the data to configure the CosmosConfig SDK
type CosmosConfig struct {
	Prefix  string   `toml:"prefix"`
	Modules []string `toml:"modules"`
}

func NewCosmosConfig(prefix string, modules []string) *CosmosConfig {
	return &CosmosConfig{
		Prefix:  prefix,
		Modules: modules,
	}
}

// DatabaseConfig represents a generic database configuration
type DatabaseConfig struct {
	Name               string `toml:"name"`
	Host               string `toml:"host"`
	Port               int64  `toml:"port"`
	User               string `toml:"user"`
	Password           string `toml:"password"`
	SSLMode            string `toml:"ssl_mode"`
	Schema             string `toml:"schema"`
	MaxOpenConnections int    `toml:"max_open_connections"`
	MaxIdleConnections int    `toml:"max_idle_connections"`
}

func NewDatabaseConfig(
	name, host string, port int64, user string, password string,
	sslMode string, schema string,
	maxOpenConnections int, maxIdleConnections int,
) *DatabaseConfig {
	return &DatabaseConfig{
		Name:               name,
		Host:               host,
		Port:               port,
		User:               user,
		Password:           password,
		SSLMode:            sslMode,
		Schema:             schema,
		MaxOpenConnections: maxOpenConnections,
		MaxIdleConnections: maxIdleConnections,
	}
}

// LoggingConfig represents the configuration for the logging part
type LoggingConfig struct {
	LogLevel  string `toml:"level"`
	LogFormat string `toml:"format"`
}

func NewLoggingConfig(level, format string) *LoggingConfig {
	return &LoggingConfig{
		LogLevel:  level,
		LogFormat: format,
	}
}

// ParsingConfig represents the configuration of the parsing
type ParsingConfig struct {
	Workers        int64 `toml:"workers"`
	ParseNewBlocks bool  `toml:"listen_new_blocks"`
	ParseOldBlocks bool  `toml:"parse_old_blocks"`
	FastSync       bool  `toml:"fast_sync"`
	StartHeight    int64 `toml:"start_height"`
}

func NewParsingConfig(
	workers int64, parseNewBlocks, parseOldBlocks bool, startHeight int64, fastSync bool,
) *ParsingConfig {
	return &ParsingConfig{
		Workers:        workers,
		ParseOldBlocks: parseOldBlocks,
		ParseNewBlocks: parseNewBlocks,
		StartHeight:    startHeight,
		FastSync:       fastSync,
	}
}
