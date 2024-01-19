package v5

import (
	"encoding/json"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/lib/pq"
	"github.com/rs/zerolog/log"

	dbtypes "github.com/forbole/juno/v5/database/migrate/utils"
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
	log.Info().Msg("migrating transactions")
	log.Debug().Int("tx count", len(msgTypes)).Msg("processing total transactions")

	for _, msgType := range msgTypes {
		log.Debug().Str("tx hash", msgType.TransactionHash).Msg("migrating transaction....")

		tx, err := db.getMsgExecTransactionsFromDatabase(msgType.TransactionHash)
		if err != nil {
			return fmt.Errorf("error while getting transaction row: %s", err)
		}

		if tx.Success == "false" {
			skipped++
			continue
		}

		var msgs sdk.ABCIMessageLogs
		err = json.Unmarshal([]byte(tx.Logs), &msgs)
		if err != nil {
			return fmt.Errorf("error while unmarshaling messages: %s", err)
		}

		var addresses []string

		for _, msg1 := range msgs {
			for _, event := range msg1.Events {
				for _, attribute := range event.Attributes {
					fmt.Printf("\t attribute %v\n", attribute)
					fmt.Printf("\t attribute value %v\n", attribute.Value)

					// Try parsing the address as a validator address
					validatorAddress, err := sdk.ValAddressFromBech32(attribute.Value)
					if err != nil {
						fmt.Printf("\t error %s \n ", err)
					}
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
		involvedAddresses := db.removeDuplicates(addresses)

		fmt.Printf("\n ADDRESSES BEFORE %s", msgType.InvolvedAccountsAddresses)
		fmt.Printf("\n ADDRESSES AFTER %s \n", involvedAddresses)

		return db.updateInvolvedAddressesInsideMessageTable(types.NewMessage(msgType.TransactionHash,
			int(msgType.Index),
			msgType.Type,
			msgType.Value,
			involvedAddresses,
			msgType.Height), msgType.PartitionID)

	}

	fmt.Printf("\n TOTAL SKIPPED %d \n", skipped)

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

// function to remove duplicate values
func (db *Migrator) removeDuplicates(s []string) []string {
	bucket := make(map[string]bool)
	var result []string
	for _, str := range s {
		if _, ok := bucket[str]; !ok {
			bucket[str] = true
			result = append(result, str)
		}
	}
	return result
}

// migrateMsgTypes stores the given message type inside the database
func (db *Migrator) updateInvolvedAddressesInsideMessageTable(msg *types.Message, partitionID int64) error {
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
