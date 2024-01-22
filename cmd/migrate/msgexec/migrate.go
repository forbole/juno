package msgexec

import (
	"fmt"

	utils "github.com/forbole/juno/v5/cmd/migrate/utils"

	parse "github.com/forbole/juno/v5/cmd/parse/types"
	"github.com/forbole/juno/v5/database"
	msgexecdb "github.com/forbole/juno/v5/database/legacy/msgexec"
	"github.com/forbole/juno/v5/database/postgresql"
)

// RunMigration runs the migrations from v4 to v5
func RunMigration(parseConfig *parse.Config) error {
	cfg, err := GetConfig()
	if err != nil {
		return fmt.Errorf("error while reading config: %s", err)
	}

	// Migrate the database
	err = migrateDb(cfg, parseConfig)
	if err != nil {
		return fmt.Errorf("error while migrating database: %s", err)
	}

	return nil
}

func migrateDb(cfg utils.Config, parseConfig *parse.Config) error {
	// Build the codec
	encodingConfig := parseConfig.GetEncodingConfigBuilder()()

	// Get the db
	databaseCtx := database.NewContext(cfg.Database, encodingConfig, parseConfig.GetLogger())
	db, err := postgresql.Builder(databaseCtx)
	if err != nil {
		return fmt.Errorf("error while building the db: %s", err)
	}

	// Build the migrator and perform the migrations
	migrator := msgexecdb.NewMigrator(db.(*postgresql.Database))
	return migrator.Migrate()
}
