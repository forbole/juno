package migrate

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/forbole/juno/v2/types/config"
	"github.com/forbole/juno/v2/cmd/parse"
	database "github.com/forbole/juno/v2/database"


)

// MigrateTablesCmd returns a Cobra command that allows to implement table partition
func MigrateTablesCmd(parseConfig *parse.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "migrate-tables",
		Short: "Migrate transaction and message tables to tables with partition",
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
		limit := config.Cfg.Database.Limit
		offset := int64(0)
		

		for {
			// SELECT rows from transaction_old table
			txRows, err := db.SelectRows(limit, offset)
			if err != nil {
				return fmt.Errorf("error while selecting transaction rows: ", err)
			}
			if len(txRows) == 0 {
				break
			}

			fmt.Printf("--- Migrating data from row %v to row %v --- \n", offset, offset+limit)
			// INSERT INTO transaction and message tables
			err = db.InsertTransactions(txRows)
			if err != nil {
				return fmt.Errorf("error while inserting data: ", err)
			}

			offset += limit
		}

		// DROP old messages_by_address function
		err = db.DropMessageByAddressFunc()
		if err != nil {
				return fmt.Errorf("error while dropping messages_by_address function: ", err)
		}

		// CREATE new messages_by_address function
		err = db.CreateMessageByAddressFunc()
		if err != nil {
				return fmt.Errorf("error while creating messages_by_address function: ", err)
		}

		fmt.Println("--- Migration completed ---")

			return nil
		},
	}

	return cmd
}
