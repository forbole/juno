package types

import "github.com/pelletier/go-toml"

var (
	// Cfg represents the configuration to be used during the execution
	Cfg Config
)

// ConfigParser represents a function that allows to parse a file contents as a Config object
type ConfigParser = func(fileContents []byte) (Config, error)

// DefaultConfigParser attempts to read and parse a Juno config from the given string bytes.
// An error reading or parsing the config results in a panic.
func DefaultConfigParser(configData []byte) (Config, error) {
	var cfg config
	err := toml.Unmarshal(configData, &cfg)
	return &cfg, err
}

// ---------------------------------------------------------------------------------------------------------------------

// Config represents the configuration to run Juno
type Config interface {
	GetRPCConfig() *RPCConfig
	GetGrpcConfig() *GrpcConfig
	GetCosmosConfig() *CosmosConfig
	GetDatabaseConfig() *DatabaseConfig
	GetLoggingConfig() *LoggingConfig
	GetParsingConfig() *ParsingConfig
	GetPruningConfig() *PruningConfig
}

var _ Config = &config{}

// Config defines all necessary juno configuration parameters.
type config struct {
	RPC      *RPCConfig      `toml:"rpc"`
	Grpc     *GrpcConfig     `toml:"grpc"`
	Cosmos   *CosmosConfig   `toml:"cosmos"`
	Database *DatabaseConfig `toml:"database"`
	Logging  *LoggingConfig  `toml:"logging"`
	Parsing  *ParsingConfig  `toml:"parsing"`
	Pruning  *PruningConfig  `toml:"pruning"`
}

// NewConfig builds a new Config instance
func NewConfig(
	rpcConfig *RPCConfig, grpConfig *GrpcConfig,
	cosmosConfig *CosmosConfig, dbConfig *DatabaseConfig,
	loggingConfig *LoggingConfig, parsingConfig *ParsingConfig,
) Config {
	return &config{
		RPC:      rpcConfig,
		Grpc:     grpConfig,
		Cosmos:   cosmosConfig,
		Database: dbConfig,
		Logging:  loggingConfig,
		Parsing:  parsingConfig,
	}
}

// GetRPCConfig implements Config
func (c *config) GetRPCConfig() *RPCConfig {
	return c.RPC
}

// GetGrpcConfig implements Config
func (c *config) GetGrpcConfig() *GrpcConfig {
	return c.Grpc
}

// GetCosmosConfig implements Config
func (c *config) GetCosmosConfig() *CosmosConfig {
	return c.Cosmos
}

// GetDatabaseConfig implements Config
func (c *config) GetDatabaseConfig() *DatabaseConfig {
	return c.Database
}

// GetLoggingConfig implements Config
func (c *config) GetLoggingConfig() *LoggingConfig {
	return c.Logging
}

// GetParsingConfig implements Config
func (c *config) GetParsingConfig() *ParsingConfig {
	return c.Parsing
}

// GetPruningConfig implements Config
func (c *config) GetPruningConfig() *PruningConfig {
	return c.Pruning
}

// ---------------------------------------------------------------------------------------------------------------------

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

// ---------------------------------------------------------------------------------------------------------------------

// RPCConfig contains the configuration of the RPC endpoint
type RPCConfig struct {
	ClientName string `toml:"client_name"`
	Address    string `toml:"address"`
}

func NewRPCConfig(clientName, address string) *RPCConfig {
	return &RPCConfig{
		ClientName: clientName,
		Address:    address,
	}
}

// ---------------------------------------------------------------------------------------------------------------------

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

// ---------------------------------------------------------------------------------------------------------------------

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

// ---------------------------------------------------------------------------------------------------------------------

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

// ---------------------------------------------------------------------------------------------------------------------

// ParsingConfig represents the configuration of the parsing
type ParsingConfig struct {
	Workers        int64 `toml:"workers"`
	ParseNewBlocks bool  `toml:"listen_new_blocks"`
	ParseOldBlocks bool  `toml:"parse_old_blocks"`
	ParseGenesis   bool  `toml:"parse_genesis"`
	StartHeight    int64 `toml:"start_height"`
	FastSync       bool  `toml:"fast_sync"`
}

func NewParsingConfig(
	workers int64,
	parseNewBlocks, parseOldBlocks bool,
	parseGenesis bool, startHeight int64, fastSync bool,
) *ParsingConfig {
	return &ParsingConfig{
		Workers:        workers,
		ParseOldBlocks: parseOldBlocks,
		ParseNewBlocks: parseNewBlocks,
		ParseGenesis:   parseGenesis,
		StartHeight:    startHeight,
		FastSync:       fastSync,
	}
}

// ---------------------------------------------------------------------------------------------------------------------

// PruningConfig contains the configuration of the pruning strategy
type PruningConfig struct {
	KeepRecent int64 `toml:"keep_recent"`
	KeepEvery  int64 `toml:"keep_every"`
	Interval   int64 `toml:"interval"`
}
