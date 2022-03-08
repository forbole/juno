package v2

import (
	loggingconfig "github.com/forbole/juno/v3/logging/config"
	"github.com/forbole/juno/v3/modules/pruning"
	"github.com/forbole/juno/v3/modules/telemetry"
	nodeconfig "github.com/forbole/juno/v3/node/config"
	parserconfig "github.com/forbole/juno/v3/parser/config"
	"github.com/forbole/juno/v3/types/config"
)

type Config struct {
	Chain    config.ChainConfig   `yaml:"chain"`
	Node     nodeconfig.Config    `yaml:"node"`
	Parser   parserconfig.Config  `yaml:"parsing"`
	Database DatabaseConfig       `yaml:"database"`
	Logging  loggingconfig.Config `yaml:"logging"`

	// The following are there to support modules which config are present if they are enabled

	Telemetry *telemetry.Config `yaml:"telemetry,omitempty"`
	Pruning   *pruning.Config   `yaml:"pruning,omitempty"`
}

type DatabaseConfig struct {
	Name               string `yaml:"name"`
	Host               string `yaml:"host"`
	Port               int64  `yaml:"port"`
	User               string `yaml:"user"`
	Password           string `yaml:"password"`
	SSLMode            string `yaml:"ssl_mode,omitempty"`
	Schema             string `yaml:"schema,omitempty"`
	MaxOpenConnections int    `yaml:"max_open_connections"`
	MaxIdleConnections int    `yaml:"max_idle_connections"`
}
