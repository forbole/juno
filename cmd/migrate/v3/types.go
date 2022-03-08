package v3

import (
	databaseconfig "github.com/forbole/juno/v2/database/config"
	loggingconfig "github.com/forbole/juno/v2/logging/config"
	"github.com/forbole/juno/v2/modules/pruning"
	"github.com/forbole/juno/v2/modules/telemetry"
	nodeconfig "github.com/forbole/juno/v2/node/config"
	parserconfig "github.com/forbole/juno/v2/parser/config"
	"github.com/forbole/juno/v2/types/config"
)

// Config defines all necessary juno configuration parameters.
type Config struct {
	Chain    config.ChainConfig    `yaml:"chain"`
	Node     nodeconfig.Config     `yaml:"node"`
	Parser   parserconfig.Config   `yaml:"parsing"`
	Database databaseconfig.Config `yaml:"database"`
	Logging  loggingconfig.Config  `yaml:"logging"`

	// The following are there to support modules which config are present if they are enabled

	Telemetry *telemetry.Config `yaml:"telemetry,omitempty"`
	Pruning   *pruning.Config   `yaml:"pruning,omitempty"`
}
