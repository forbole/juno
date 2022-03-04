package migrate

import (
	"github.com/spf13/cobra"
)

// MigrateTablesCmd returns a Cobra command that allows to implement table partition
func MigrateTablesCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "migrate-tables",
		Short: "Migrate transaction and message tables to tables with partition",
		RunE: func(cmd *cobra.Command, args []string) error {
	
		limit := db.Config.Database.Limit
		offset := int64(0)

		for {
			// SELECT rows from transaction_old table
			txRows, err := db.SelectRows(limit, offset)
			if err != nil {
				log.Fatal("error while selecting transaction rows: ", err)
				return
			}
			if len(txRows) == 0 {
				break
			}

			fmt.Printf("--- Migrating data from row %v to row %v --- \n", offset, offset+limit)
			// INSERT INTO transaction and message tables
			err = db.InsertTransactions(txRows)
			if err != nil {
				log.Fatal("error while inserting data: ", err)
				return
			}

			offset += limit
		}

		// DROP old messages_by_address function
		err := db.DropMessageByAddressFunc()
		if err != nil {
			log.Fatal("error while dropping messages_by_address function: ", err)
			return
		}

		// CREATE new messages_by_address function
		err = db.CreateMessageByAddressFunc()
		if err != nil {
			log.Fatal("error while creating messages_by_address function: ", err)
			return
		}

		fmt.Println("--- Migration completed ---")
		return

			return nil
		},
	}

	return cmd
}
