package msgexec

import (
	"encoding/json"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/lib/pq"
	"github.com/rs/zerolog/log"

	dbtypes "github.com/forbole/juno/v5/database/migrate/utils"
	msgmodule "github.com/forbole/juno/v5/modules/messages"
	"github.com/forbole/juno/v5/types"
)

// Migrate implements database.Migrator
func (db *Migrator) Migrate() error {
	msgTypes, err := db.getAllMsgExecStoredInDatabase()
	if err != nil {
		return fmt.Errorf("error while getting message types rows: %s", err)
	}

	var skipped = 0
	// Migrate the transactions
	log.Info().Msg("** migrating transactions **")
	log.Debug().Int("tx count", len(msgTypes)).Msg("processing total transactions")

	for _, msgType := range msgTypes {
		log.Debug().Str("tx hash", msgType.TransactionHash).Msg("getting transaction....")

		tx, err := db.getMsgExecTransactionsFromDatabase(msgType.TransactionHash)
		if err != nil {
			return fmt.Errorf("error while getting transaction %s: %s", msgType.TransactionHash, err)
		}

		if tx.Success == "true" {
			var msgs sdk.ABCIMessageLogs
			err = json.Unmarshal([]byte(tx.Logs), &msgs)
			if err != nil {
				return fmt.Errorf("error while unmarshaling tx logs: %s", err)
			}

			var addresses []string

			for _, msg := range msgs {
				for _, event := range msg.Events {
					for _, attribute := range event.Attributes {
						// Try parsing the address as a validator address
						validatorAddress, _ := sdk.ValAddressFromBech32(attribute.Value)
						if validatorAddress != nil {
							addresses = append(addresses, validatorAddress.String())
						}

						// Try parsing the address as an account address
						accountAddress, err := sdk.AccAddressFromBech32(attribute.Value)
						if err != nil {
							// Skip if the address is not an account address
							continue
						}

						addresses = append(addresses, accountAddress.String())
					}
				}
			}
			involvedAddresses := msgmodule.RemoveDuplicates(addresses)

			fmt.Printf("\n ADDRESSES BEFORE %s", msgType.InvolvedAccountsAddresses)
			fmt.Printf("\n ADDRESSES AFTER %s \n", involvedAddresses)

			err = db.updateMessage(types.NewMessage(msgType.TransactionHash,
				int(msgType.Index),
				msgType.Type,
				msgType.Value,
				involvedAddresses,
				msgType.Height), msgType.PartitionID)

			if err != nil {
				return fmt.Errorf("error while storing updated message: %s", err)
			}
		} else {
			skipped++
		}

	}

	log.Debug().Int("*** Total Skipped ***", skipped)

	return nil

}

// getMsgTypesFromMessageTable retrieves messages types stored in database inside message table
func (db *Migrator) getAllMsgExecStoredInDatabase() ([]dbtypes.MessageRow, error) {
	const msgType = "cosmos.authz.v1beta1.MsgExec"
	var rows []dbtypes.MessageRow
	err := db.SQL.Select(&rows, `SELECT * FROM message WHERE type = $1`, msgType)
	if err != nil {
		return nil, err
	}

	return rows, nil
}

// getMsgTypesFromMessageTable retrieves messages types stored in database inside message table
func (db *Migrator) getMsgExecTransactionsFromDatabase(txHash string) (dbtypes.TransactionRow, error) {
	var rows []dbtypes.TransactionRow
	err := db.SQL.Select(&rows, `SELECT * FROM transaction WHERE hash = $1`, txHash)
	if err != nil {
		return dbtypes.TransactionRow{}, err
	}

	return rows[0], nil
}

// updateMessage stores updated message inside the database
func (db *Migrator) updateMessage(msg *types.Message, partitionID int64) error {
	stmt := `
INSERT INTO message(transaction_hash, index, type, value, involved_accounts_addresses, height, partition_id) 
VALUES ($1, $2, $3, $4, $5, $6, $7) 
ON CONFLICT (transaction_hash, index, partition_id) DO UPDATE 
	SET height = excluded.height, 
		type = excluded.type,
		value = excluded.value,
		involved_accounts_addresses = excluded.involved_accounts_addresses`

	_, err := db.SQL.Exec(stmt, msg.TxHash, msg.Index, msg.Type, msg.Value, pq.Array(msg.Addresses), msg.Height, partitionID)
	return err

}
