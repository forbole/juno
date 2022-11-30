package v4

import (
	"fmt"
	"io/ioutil"

	parsecmdtypes "github.com/forbole/juno/v4/cmd/parse/types"

	"gopkg.in/yaml.v3"

	v3 "github.com/forbole/juno/v4/cmd/migrate/v3"
	databaseconfig "github.com/forbole/juno/v4/database/config"
	"github.com/forbole/juno/v4/types/config"
)

// RunMigration runs the migrations from v3 to v4
func RunMigration(parseConfig *parsecmdtypes.Config) error {
	// Migrate the config
	cfg, err := migrateConfig()
	if err != nil {
		return fmt.Errorf("error while migrating config: %s", err)
	}

	// Refresh the global configuration
	err = parsecmdtypes.UpdatedGlobalCfg(parseConfig)
	if err != nil {
		return err
	}

	bz, err := yaml.Marshal(&cfg)
	if err != nil {
		return fmt.Errorf("error while serializing config: %s", err)
	}

	err = ioutil.WriteFile(config.GetConfigFilePath(), bz, 0600)
	if err != nil {
		return fmt.Errorf("error while writing v4 config: %s", err)
	}

	return nil
}

func migrateConfig() (Config, error) {
	cfg, err := v3.GetConfig()
	if err != nil {
		return Config{}, fmt.Errorf("error while reading v3 config: %s", err)
	}

	sslMode := cfg.Database.SSLMode
	if sslMode == "" {
		sslMode = "disable"
	}

	schema := cfg.Database.Schema
	if schema == "" {
		schema = "public"
	}

	return Config{
		Node:  cfg.Node,
		Chain: cfg.Chain,
		Database: databaseconfig.Config{
			URL: fmt.Sprintf("postgresql://%s:%s@%s:%d/%s?sslmode=%s&search_path=%s",
				cfg.Database.User,
				cfg.Database.Password,
				cfg.Database.Host,
				cfg.Database.Port,
				cfg.Database.Name,
				sslMode,
				schema,
			),
			MaxOpenConnections: cfg.Database.MaxOpenConnections,
			MaxIdleConnections: cfg.Database.MaxIdleConnections,
			PartitionSize:      cfg.Database.PartitionSize,
			PartitionBatchSize: cfg.Database.PartitionBatchSize,
		},
		Parser:    cfg.Parser,
		Logging:   cfg.Logging,
		Telemetry: cfg.Telemetry,
		Pruning:   cfg.Pruning,
		PriceFeed: cfg.PriceFeed,
	}, nil
}
