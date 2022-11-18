package v3

import (
	"fmt"
	"io/ioutil"
	"time"

	parsecmdtypes "github.com/saifullah619/juno/v3/cmd/parse/types"
	parserconfig "github.com/saifullah619/juno/v3/parser/config"

	"gopkg.in/yaml.v3"

	v2 "github.com/saifullah619/juno/v3/cmd/migrate/v2"
	"github.com/saifullah619/juno/v3/database"
	databaseconfig "github.com/saifullah619/juno/v3/database/config"
	v3db "github.com/saifullah619/juno/v3/database/legacy/v3"
	"github.com/saifullah619/juno/v3/database/postgresql"
	"github.com/saifullah619/juno/v3/types/config"
)

// RunMigration runs the migrations from v2 to v3
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

	// Get the new fields added inside the various configurations
	var partitionSize int64
	if cfg.Database.PartitionSize != nil {
		partitionSize = *cfg.Database.PartitionSize
	}

	var partitionBatchSize int64
	if cfg.Database.PartitionBatchSize != nil {
		partitionBatchSize = *cfg.Database.PartitionBatchSize
	}

	var averageBlockTime = 3 * time.Second
	if cfg.Parser.AvgBlockTime != nil {
		averageBlockTime = *cfg.Parser.AvgBlockTime
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
			PartitionSize:      partitionSize,
			PartitionBatchSize: partitionBatchSize,
		},
		Parser: parserconfig.Config{
			Workers:         cfg.Parser.Workers,
			ParseNewBlocks:  cfg.Parser.ParseNewBlocks,
			ParseOldBlocks:  cfg.Parser.ParseOldBlocks,
			GenesisFilePath: cfg.Parser.GenesisFilePath,
			ParseGenesis:    cfg.Parser.ParseGenesis,
			StartHeight:     cfg.Parser.StartHeight,
			FastSync:        cfg.Parser.FastSync,
			AvgBlockTime:    &averageBlockTime,
		},
		Logging:   cfg.Logging,
		Telemetry: cfg.Telemetry,
		Pruning:   cfg.Pruning,
		PriceFeed: cfg.PriceFeed,
	}, nil
}

func migrateDb(cfg Config, parseConfig *parsecmdtypes.Config) error {
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
