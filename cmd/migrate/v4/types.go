package v4

import (
	databaseconfig "github.com/emrahm/juno/v5/database/config"
	loggingconfig "github.com/emrahm/juno/v5/logging/config"
	"github.com/emrahm/juno/v5/modules/pruning"
	"github.com/emrahm/juno/v5/modules/telemetry"
	nodeconfig "github.com/emrahm/juno/v5/node/config"
	parserconfig "github.com/emrahm/juno/v5/parser/config"
	pricefeedconfig "github.com/emrahm/juno/v5/pricefeed"
	"github.com/emrahm/juno/v5/types/config"
)

// Config defines all necessary juno configuration parameters.
type Config struct {
	Chain    config.ChainConfig    `yaml:"chain"`
	Node     nodeconfig.Config     `yaml:"node"`
	Parser   parserconfig.Config   `yaml:"parsing"`
	Database databaseconfig.Config `yaml:"database"`
	Logging  loggingconfig.Config  `yaml:"logging"`

	// The following are there to support modules which config are present if they are enabled

	Telemetry *telemetry.Config       `yaml:"telemetry,omitempty"`
	Pruning   *pruning.Config         `yaml:"pruning,omitempty"`
	PriceFeed *pricefeedconfig.Config `yaml:"pricefeed,omitempty"`
}
