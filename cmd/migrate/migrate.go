package migrate

import (
	"fmt"
	"io/ioutil"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

	v1 "github.com/forbole/juno/v2/cmd/migrate/v1"
	databaseconfig "github.com/forbole/juno/v2/database/config"
	loggingconfig "github.com/forbole/juno/v2/logging/config"
	"github.com/forbole/juno/v2/modules/pruning"
	"github.com/forbole/juno/v2/modules/telemetry"
	nodeconfig "github.com/forbole/juno/v2/node/config"
	"github.com/forbole/juno/v2/node/remote"
	parserconfig "github.com/forbole/juno/v2/parser/config"
	"github.com/forbole/juno/v2/types/config"
)

type Config struct {
	Chain     *config.ChainConfig    `yaml:"chain"`
	Node      *nodeconfig.Config     `yaml:"node"`
	Parser    *parserconfig.Config   `yaml:"parsing"`
	Database  *databaseconfig.Config `yaml:"database"`
	Logging   *loggingconfig.Config  `yaml:"logging"`
	Telemetry *telemetry.Config      `yaml:"telemetry,omitempty"`
	Pruning   *pruning.Config        `yaml:"pruning,omitempty"`
}

// MigrateCmd returns the command that should be run when we want to migrate from v1 to v2
func MigrateCmd() *cobra.Command {
	return &cobra.Command{
		Use: "migrate",
		RunE: func(cmd *cobra.Command, args []string) error {
			v2Config, err := MigrateConfig()
			if err != nil {
				return nil
			}

			bz, err := yaml.Marshal(&v2Config)
			if err != nil {
				return fmt.Errorf("error while serializing v2 config: %s", err)
			}

			v2File := config.GetConfigFilePath()
			return ioutil.WriteFile(v2File, bz, 0666)
		},
	}
}

func MigrateConfig() (Config, error) {
	bz, err := v1.ReadConfig()
	if err != nil {
		return Config{}, nil
	}

	cfg, err := v1.ParseConfig(bz)
	if err != nil {
		return Config{}, fmt.Errorf("error while parsing v1 config: %s", err)
	}

	v2Cfg := config.NewConfig(
		nodeconfig.NewConfig(
			nodeconfig.TypeRemote,
			remote.NewDetails(
				remote.NewRPCConfig(
					cfg.RPC.ClientName,
					cfg.RPC.Address,
					cfg.RPC.MaxConnections,
				),
				remote.NewGrpcConfig(
					cfg.Grpc.Address,
					cfg.Grpc.Insecure,
				),
			),
		),
		config.NewChainConfig(
			cfg.Cosmos.Prefix,
			cfg.Cosmos.Modules,
		),
		databaseconfig.NewDatabaseConfig(
			cfg.Database.Name,
			cfg.Database.Host,
			cfg.Database.Port,
			cfg.Database.User,
			cfg.Database.Password,
			cfg.Database.SSLMode,
			cfg.Database.Schema,
			cfg.Database.MaxOpenConnections,
			cfg.Database.MaxIdleConnections,
		),
		parserconfig.NewParsingConfig(
			cfg.Parsing.Workers,
			cfg.Parsing.ParseNewBlocks,
			cfg.Parsing.ParseOldBlocks,
			cfg.Parsing.ParseGenesis,
			cfg.Parsing.GenesisFilePath,
			cfg.Parsing.StartHeight,
			cfg.Parsing.FastSync,
		),
		loggingconfig.NewLoggingConfig(
			cfg.Logging.LogLevel,
			cfg.Logging.LogFormat,
		),
	)

	var telemetryConfig *telemetry.Config
	if cfg.Telemetry != nil {
		telemetryConfig = telemetry.NewConfig(cfg.Telemetry.Port)

		if cfg.Telemetry.Enabled {
			v2Cfg.Chain.Modules = appendModuleIfNotExisting(v2Cfg.Chain.Modules, telemetry.ModuleName)
		}
	}

	var pruningConfig *pruning.Config
	if cfg.Pruning != nil {
		pruningConfig = pruning.NewConfig(
			cfg.Pruning.KeepRecent,
			cfg.Pruning.KeepEvery,
			cfg.Pruning.Interval,
		)
	}

	return Config{
		Chain:     &v2Cfg.Chain,
		Node:      &v2Cfg.Node,
		Parser:    &v2Cfg.Parser,
		Database:  &v2Cfg.Database,
		Logging:   &v2Cfg.Logging,
		Telemetry: telemetryConfig,
		Pruning:   pruningConfig,
	}, nil
}

func appendModuleIfNotExisting(modules []string, module string) []string {
	var found = false
	for _, m := range modules {
		if m == module {
			found = true
		}
	}

	if !found {
		return append(modules, module)
	}

	return modules
}
