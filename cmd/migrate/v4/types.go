package v4

import (
	databaseconfig "github.com/0xPellNetwork/juno/v6/database/config"
	loggingconfig "github.com/0xPellNetwork/juno/v6/logging/config"
	"github.com/0xPellNetwork/juno/v6/modules/pruning"
	"github.com/0xPellNetwork/juno/v6/modules/telemetry"
	nodeconfig "github.com/0xPellNetwork/juno/v6/node/config"
	parserconfig "github.com/0xPellNetwork/juno/v6/parser/config"
	pricefeedconfig "github.com/0xPellNetwork/juno/v6/pricefeed"
	"github.com/0xPellNetwork/juno/v6/types/config"
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
