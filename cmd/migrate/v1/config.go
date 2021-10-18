package v1

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"

	"github.com/pelletier/go-toml"

	"github.com/forbole/juno/v2/types/config"
)

// GetConfigFilePath returns the path to the configuration file given the executable name
func GetConfigFilePath() string {
	return path.Join(config.HomePath, "config.toml")
}

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

// ReadConfig reads the config.toml file contents
func ReadConfig() ([]byte, error) {
	v1File := GetConfigFilePath()

	// Make sure the path exists
	if _, err := os.Stat(v1File); os.IsNotExist(err) {
		return nil, fmt.Errorf("config v1File does not exist")
	}

	bz, err := ioutil.ReadFile(v1File)
	if err != nil {
		return nil, fmt.Errorf("error while reading v1 config files: %s", err)
	}

	return bz, nil
}

// ParseConfig attempts to read and parse a Juno Config from the given string bytes.
// An error reading or parsing the Config results in a panic.
func ParseConfig(configData []byte) (Config, error) {
	var cfg Config
	err := toml.Unmarshal(configData, &cfg)
	return cfg, err
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
	Workers         int64  `toml:"workers"`
	ParseNewBlocks  bool   `toml:"listen_new_blocks"`
	ParseOldBlocks  bool   `toml:"parse_old_blocks"`
	GenesisFilePath string `toml:"genesis_file_path"`
	ParseGenesis    bool   `toml:"parse_genesis"`
	StartHeight     int64  `toml:"start_height"`
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
