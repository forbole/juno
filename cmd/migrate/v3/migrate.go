package v3

import (
	"fmt"
	"io/ioutil"

	"gopkg.in/yaml.v3"

	v2 "github.com/forbole/juno/v2/cmd/migrate/v2"
	"github.com/forbole/juno/v2/cmd/parse"
	"github.com/forbole/juno/v2/database"
	databaseconfig "github.com/forbole/juno/v2/database/config"
	v3db "github.com/forbole/juno/v2/database/legacy/v3"
	"github.com/forbole/juno/v2/database/postgresql"
	"github.com/forbole/juno/v2/types/config"
)

// RunMigration runs the migrations from v2 to v3
func RunMigration(parseConfig *parse.Config) error {
	// Migrate the config
	cfg, err := migrateConfig()
	if err != nil {
		return fmt.Errorf("error while migrating config: %s", err)
	}

	bz, err := yaml.Marshal(&cfg)
	if err != nil {
		return fmt.Errorf("error while serializing config: %s", err)
	}

	err = ioutil.WriteFile(config.GetConfigFilePath(), bz, 0666)
	if err != nil {
		return fmt.Errorf("error while writing v3 config: %s", err)
	}

	// Migrate the database
	err = migrateDb(cfg, parseConfig)
	if err != nil {
		return fmt.Errorf("error while migrating database: %s", err)
	}

	return nil
}

func migrateConfig() (Config, error) {
	cfg, err := v2.GetConfig()
	if err != nil {
		return Config{}, fmt.Errorf("error while reading v2 config: %s", err)
	}

	return Config{
		Node:  cfg.Node,
		Chain: cfg.Chain,
		Database: databaseconfig.Config{
			Name:               cfg.Database.Name,
			Host:               cfg.Database.Host,
			Port:               cfg.Database.Port,
			User:               cfg.Database.User,
			Password:           cfg.Database.Password,
			SSLMode:            cfg.Database.SSLMode,
			Schema:             cfg.Database.Schema,
			MaxOpenConnections: cfg.Database.MaxOpenConnections,
			MaxIdleConnections: cfg.Database.MaxIdleConnections,
			PartitionSize:      0,
			PartitionBatchSize: 0,
		},
		Parser:    cfg.Parser,
		Logging:   cfg.Logging,
		Telemetry: cfg.Telemetry,
		Pruning:   cfg.Pruning,
	}, nil
}

func migrateDb(cfg Config, parseConfig *parse.Config) error {
	// Build the codec
	encodingConfig := parseConfig.GetEncodingConfigBuilder()()

	// Get the db
	databaseCtx := database.NewContext(cfg.Database, &encodingConfig, parseConfig.GetLogger())
	db, err := postgresql.Builder(databaseCtx)
	if err != nil {
		return fmt.Errorf("error while building the db: %s", err)
	}

	// Build the migrator and perform the migrations
	migrator := v3db.NewMigrator(db.(*postgresql.Database))
	return migrator.Migrate()

}
