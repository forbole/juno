package v3

import (
	"encoding/json"
	"fmt"

	"github.com/rs/zerolog/log"

	"github.com/forbole/juno/v3/types/config"

	types "github.com/forbole/juno/v3/database/migrate/utils"
)

// Migrate implements database.Migrator
func (db *Migrator) Migrate() error {
	batchSize := config.Cfg.Database.PartitionBatchSize
	if batchSize == 0 {
		log.Info().Msg("partition batch size is set to 0, skipping migration")
		return nil
	}

	partitionSize := config.Cfg.Database.PartitionSize
	if partitionSize == 0 {
		log.Info().Msg("partition size is set to 0, skipping migration")
		return nil
	}

	// Prepare the migration
	log.Info().Msg("preparing the tables for the migration")
	err := db.PrepareMigration()
	if err != nil {
		return err
	}

	// Migrate the transactions
	log.Info().Msg("migrating transactions")
	var offset int64
	for {
		rows, err := db.getOldTransactions(batchSize, offset)
		if err != nil {
			return fmt.Errorf("error while getting old transaction rows: %s", err)
		}

		// Stop migrating if there are no more rows
		if len(rows) == 0 {
			break
		}

		// Perform migration
		log.Debug().Int64("start row", offset).Int64("end row", offset+batchSize).Msg("migrating transactions")
		err = db.migrateTransactions(rows, partitionSize)
		if err != nil {
			return fmt.Errorf("error while inserting data: %s", err)
		}

		offset += batchSize
	}

	return nil
}

func (db *Migrator) getOldTransactions(batchSize int64, offset int64) ([]types.TransactionRow, error) {
	stmt := fmt.Sprintf("SELECT * FROM transaction_old ORDER BY height LIMIT %v OFFSET %v", batchSize, offset)

	var rows []types.TransactionRow
	err := db.SQL.Select(&rows, stmt)
	if err != nil {
		return nil, err
	}

	return rows, nil
}

func (db *Migrator) createPartitionTable(table string, partitionID int64) error {
	partitionTable := fmt.Sprintf("%s_%v", table, partitionID)

	stmt := fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %s PARTITION OF %s FOR VALUES IN (%v)`, partitionTable, table, partitionID)
	_, err := db.SQL.Exec(stmt)
	return err
}

func (db *Migrator) migrateTransactions(rows []types.TransactionRow, partitionSize int64) error {
	stmt := `INSERT INTO transaction 
(hash, height, success, messages, memo, signatures, signer_infos, fee, gas_wanted, gas_used, raw_log, logs, partition_id) VALUES 
`
	var params []interface{}
	for i, tx := range rows {
		// Create transaction partition table if not exists
		partitionID := tx.Height / partitionSize
		err := db.createPartitionTable("transaction", partitionID)
		if err != nil {
			return fmt.Errorf("error while creating transaction partition table: %s", err)
		}

		log.Debug().Int64("tx height", tx.Height).Msg("processing transactions")

		// Append params
		params = append(params, tx.Hash, tx.Height, tx.Success, tx.Messages, tx.Memo, tx.Signatures,
			tx.SignerInfos, tx.Fee, tx.GasWanted, tx.GasUsed, tx.RawLog, tx.Logs, partitionID)

		// Add columns to stmt
		ai := i * 13
		stmt += fmt.Sprintf("($%v, $%v, $%v, $%v, $%v, $%v, $%v, $%v, $%v, $%v, $%v, $%v, $%v),",
			ai+1, ai+2, ai+3, ai+4, ai+5, ai+6, ai+7, ai+8, ai+9, ai+10, ai+11, ai+12, ai+13)
	}
	stmt = stmt[:len(stmt)-1] // remove trailing ,
	stmt += " ON CONFLICT DO NOTHING"

	_, err := db.SQL.Exec(stmt, params...)
	if err != nil {
		return fmt.Errorf("error while inserting transaction: %s", err)
	}

	for _, tx := range rows {
		log.Debug().Int64("tx height", tx.Height).Msg("processing transaction messages")

		// Insert the messages of this transaction
		err = db.insertTransactionMessages(tx, partitionSize)
		if err != nil {
			return fmt.Errorf("error while inserting messages: %s", err)
		}
	}

	return nil
}

func (db *Migrator) insertTransactionMessages(tx types.TransactionRow, partitionSize int64) error {
	partitionID := tx.Height / partitionSize

	// Create message partition table if not exists
	err := db.createPartitionTable("message", partitionID)
	if err != nil {
		return fmt.Errorf("error while creating message partition table: %s", err)
	}

	// Prepare stmt
	stmt := `INSERT INTO message 
(transaction_hash, index, type, value, involved_accounts_addresses, height, partition_id) VALUES `

	// Prepare params
	var params []interface{}

	// Unmarshal messages
	var msgs []map[string]interface{}
	err = json.Unmarshal([]byte(tx.Messages), &msgs)
	if err != nil {
		return fmt.Errorf("error while unmarshaling messages: %s", err)
	}

	for i, msg := range msgs {
		msg := msg

		// Append params
		msgType := msg["@type"].(string)[1:] // remove head "/"
		involvedAddresses := types.MessageParser(msg)
		delete(msg, "@type")

		mBz, err := json.Marshal(&msg)
		if err != nil {
			return fmt.Errorf("error while marshaling msg value to json: %s", err)
		}
		params = append(params, tx.Hash, i, msgType, string(mBz), involvedAddresses, tx.Height, partitionID)

		// Add columns to stmt
		ai := i * 7
		stmt += fmt.Sprintf("($%v, $%v, $%v, $%v, $%v, $%v, $%v),",
			ai+1, ai+2, ai+3, ai+4, ai+5, ai+6, ai+7)
	}

	stmt = stmt[:len(stmt)-1] // remove trailing ","
	stmt += " ON CONFLICT DO NOTHING"

	_, err = db.SQL.Exec(stmt, params...)
	return err
}
