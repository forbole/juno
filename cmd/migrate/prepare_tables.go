package migrate

import (
	"fmt"
	"github.com/forbole/juno/v2/cmd/parse"
	"github.com/spf13/cobra"
	database "github.com/forbole/juno/v2/database"
	"github.com/forbole/juno/v2/types/config"
)

// PrepareTablesCmd returns a Cobra command that allows to prepare transaction and message tables for postgresql partition
func PrepareTablesCmd(parseConfig *parse.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "prepare-tables",
		Short: "Prepare transaction and message tables for postgresql partition",
		RunE: func(cmd *cobra.Command, args []string) error {
		// Get config		
		cfg := config.Cfg

		// Build the codec
		encodingConfig := parseConfig.GetEncodingConfigBuilder()()
		
		// Get the db	
		databaseCtx := database.NewContext(cfg.Database, &encodingConfig, parseConfig.GetLogger())
		db, err := parseConfig.GetDBBuilder()(databaseCtx)
		if err != nil {
			return fmt.Errorf("Error while getting the db: %s", err)
		}

		fmt.Println("--- Preparing tables ---")

		// ALTER tables and indexes to add "_old" tags
		err = db.AlterTables()
		if err != nil {
			return fmt.Errorf("Error while altering tables: %s", err)
		}

		// CREATE new tables with new indexes
		err = db.CreateTables()
		if err != nil {
			return fmt.Errorf("Error while creating tables:  %s", err)
		}

		fmt.Println("--- Preparing tables completed ---")

			return nil
		},
	}

	return cmd
}
