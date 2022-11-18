package v2

import (
	"fmt"
	"io/ioutil"

	parsecmdtypes "github.com/saifullah619/juno/v3/cmd/parse/types"

	"gopkg.in/yaml.v3"

	v1 "github.com/saifullah619/juno/v3/cmd/migrate/v1"

	loggingconfig "github.com/saifullah619/juno/v3/logging/config"
	"github.com/saifullah619/juno/v3/modules/pruning"
	"github.com/saifullah619/juno/v3/modules/telemetry"
	nodeconfig "github.com/saifullah619/juno/v3/node/config"
	"github.com/saifullah619/juno/v3/node/remote"
	"github.com/saifullah619/juno/v3/types/config"
)

// RunMigration runs the migration that migrates the data from v1 to v2
func RunMigration(_ *parsecmdtypes.Config) error {
	v2Config, err := migrateConfig()
	if err != nil {
		return err
	}

	bz, err := yaml.Marshal(&v2Config)
	if err != nil {
		return fmt.Errorf("error while serializing v2 config: %s", err)
	}

	v2File := config.GetConfigFilePath()
	return ioutil.WriteFile(v2File, bz, 0600)
}

func migrateConfig() (Config, error) {
	cfg, err := v1.GetConfig()
	if err != nil {
		return Config{}, fmt.Errorf("error while parsing v1 config: %s", err)
	}

	v2Cfg := &Config{
		Node: nodeconfig.Config{
			Type: nodeconfig.TypeRemote,
			Details: remote.NewDetails(
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
		},
		Chain: config.ChainConfig{
			Bech32Prefix: cfg.Cosmos.Prefix,
			Modules:      cfg.Cosmos.Modules,
		},
		Database: DatabaseConfig{
			Name:               cfg.Database.Name,
			Host:               cfg.Database.Host,
			Port:               cfg.Database.Port,
			User:               cfg.Database.User,
			Password:           cfg.Database.Password,
			SSLMode:            cfg.Database.SSLMode,
			Schema:             cfg.Database.Schema,
			MaxOpenConnections: cfg.Database.MaxOpenConnections,
			MaxIdleConnections: cfg.Database.MaxIdleConnections,
		},
		Parser: ParserConfig{
			Workers:         cfg.Parsing.Workers,
			ParseNewBlocks:  cfg.Parsing.ParseNewBlocks,
			ParseOldBlocks:  cfg.Parsing.ParseOldBlocks,
			ParseGenesis:    cfg.Parsing.ParseGenesis,
			GenesisFilePath: cfg.Parsing.GenesisFilePath,
			StartHeight:     cfg.Parsing.StartHeight,
			FastSync:        cfg.Parsing.FastSync,
		},
		Logging: loggingconfig.Config{
			LogLevel:  cfg.Logging.LogLevel,
			LogFormat: cfg.Logging.LogFormat,
		},
	}

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
		Chain:     v2Cfg.Chain,
		Node:      v2Cfg.Node,
		Parser:    v2Cfg.Parser,
		Database:  v2Cfg.Database,
		Logging:   v2Cfg.Logging,
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
