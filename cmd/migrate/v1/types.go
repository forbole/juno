package v1

type Config struct {
	RPC       *RPCConfig       `toml:"rpc"`
	Grpc      *GrpcConfig      `toml:"grpc"`
	Cosmos    *CosmosConfig    `toml:"cosmos"`
	Database  *DatabaseConfig  `toml:"database"`
	Logging   *LoggingConfig   `toml:"logging"`
	Parsing   *ParsingConfig   `toml:"parsing"`
	Pruning   *PruningConfig   `toml:"pruning"`
	Telemetry *TelemetryConfig `toml:"telemetry"`
}

type RPCConfig struct {
	ClientName     string `toml:"client_name"`
	Address        string `toml:"address"`
	MaxConnections int    `toml:"max_connections"`
}

type GrpcConfig struct {
	Address  string `toml:"address"`
	Insecure bool   `toml:"insecure"`
}

type CosmosConfig struct {
	Prefix  string   `toml:"prefix"`
	Modules []string `toml:"modules"`
}

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

type LoggingConfig struct {
	LogLevel  string `toml:"level"`
	LogFormat string `toml:"format"`
}

type ParsingConfig struct {
	GenesisFilePath string `toml:"genesis_file_path"`
	Workers         int64  `toml:"workers"`
	StartHeight     int64  `toml:"start_height"`
	ParseNewBlocks  bool   `toml:"listen_new_blocks"`
	ParseOldBlocks  bool   `toml:"parse_old_blocks"`
	ParseGenesis    bool   `toml:"parse_genesis"`
	FastSync        bool   `toml:"fast_sync"`
}

type PruningConfig struct {
	KeepRecent int64 `toml:"keep_recent"`
	KeepEvery  int64 `toml:"keep_every"`
	Interval   int64 `toml:"interval"`
}

type TelemetryConfig struct {
	Enabled bool `toml:"enabled"`
	Port    uint `toml:"port"`
}
