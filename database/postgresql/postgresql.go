package postgresql

import (
	"database/sql"
	"encoding/base64"
	"fmt"
	"strings"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/jmoiron/sqlx"

	"github.com/forbole/juno/v4/logging"

	"github.com/lib/pq"

	"github.com/forbole/juno/v4/database"
	"github.com/forbole/juno/v4/types"
	"github.com/forbole/juno/v4/types/config"
	"github.com/forbole/juno/v4/types/env"
	"github.com/forbole/juno/v4/types/utils"
)

// Builder creates a database connection with the given database connection info
// from config. It returns a database connection handle or an error if the
// connection fails.
func Builder(ctx *database.Context) (database.Database, error) {
	dbURI := utils.GetEnvOr(env.DatabaseURI, ctx.Cfg.URL)
	dbEnableSSL := utils.GetEnvOr(env.DatabaseSSLModeEnable, ctx.Cfg.SSLModeEnable)

	// Configure SSL certificates (optional)
	if dbEnableSSL == "true" {
		dbRootCert := utils.GetEnvOr(env.DatabaseSSLRootCert, ctx.Cfg.SSLRootCert)
		dbCert := utils.GetEnvOr(env.DatabaseSSLCert, ctx.Cfg.SSLCert)
		dbKey := utils.GetEnvOr(env.DatabaseSSLKey, ctx.Cfg.SSLKey)
		dbURI += fmt.Sprintf(" sslmode=require sslrootcert=%s sslcert=%s sslkey=%s",
			dbRootCert, dbCert, dbKey)
	}

	postgresDb, err := sqlx.Open("postgres", dbURI)
	if err != nil {
		return nil, err
	}

	// Set max open connections
	postgresDb.SetMaxOpenConns(ctx.Cfg.MaxOpenConnections)
	postgresDb.SetMaxIdleConns(ctx.Cfg.MaxIdleConnections)

	return &Database{
		Cdc:   ctx.EncodingConfig.Codec,
		Amino: ctx.EncodingConfig.Amino,

		SQL:    postgresDb,
		Logger: ctx.Logger,
	}, nil
}

// type check to ensure interface is properly implemented
var _ database.Database = &Database{}

// Database defines a wrapper around a SQL database and implements functionality
// for data aggregation and exporting.
type Database struct {
	Cdc   codec.Codec
	Amino *codec.LegacyAmino

	SQL    *sqlx.DB
	Logger logging.Logger
}

// CreatePartitionIfNotExists creates a new partition having the given partition id if not existing
func (db *Database) CreatePartitionIfNotExists(table string, partitionID int64) error {
	partitionTable := fmt.Sprintf("%s_%d", table, partitionID)

	stmt := fmt.Sprintf(
		"CREATE TABLE IF NOT EXISTS %s PARTITION OF %s FOR VALUES IN (%d)",
		partitionTable,
		table,
		partitionID,
	)
	_, err := db.SQL.Exec(stmt)

	if err != nil {
		return err
	}

	return nil
}

// -------------------------------------------------------------------------------------------------------------------

// HasBlock implements database.Database
func (db *Database) HasBlock(height int64) (bool, error) {
	var res bool
	err := db.SQL.QueryRow(`SELECT EXISTS(SELECT 1 FROM block WHERE height = $1);`, height).Scan(&res)
	return res, err
}

// GetLastBlockHeight returns the last block height stored inside the database
func (db *Database) GetLastBlockHeight() (int64, error) {
	stmt := `SELECT height FROM block ORDER BY height DESC LIMIT 1;`

	var height int64
	err := db.SQL.QueryRow(stmt).Scan(&height)
	if err != nil {
		if strings.Contains(err.Error(), "no rows in result set") {
			// If no rows stored in block table, return 0 as height
			return 0, nil
		}
		return 0, fmt.Errorf("error while getting last block height, error: %s", err)
	}

	return height, nil
}

// GetMissingHeights returns a slice of missing block heights between startHeight and endHeight
func (db *Database) GetMissingHeights(startHeight, endHeight int64) []int64 {
	var result []int64
	stmt := `SELECT generate_series($1::int,$2::int) EXCEPT SELECT height FROM block ORDER BY 1;`
	err := db.SQL.Select(&result, stmt, startHeight, endHeight)
	if err != nil {
		return nil
	}

	if len(result) == 0 {
		return nil
	}

	return result
}

// SaveBlock implements database.Database
func (db *Database) SaveBlock(block *types.Block) error {
	sqlStatement := `
INSERT INTO block (height, hash, num_txs, total_gas, proposer_address, timestamp)
VALUES ($1, $2, $3, $4, $5, $6) ON CONFLICT DO NOTHING`

	proposerAddress := sql.NullString{Valid: len(block.ProposerAddress) != 0, String: block.ProposerAddress}
	_, err := db.SQL.Exec(sqlStatement,
		block.Height, block.Hash, block.TxNum, block.TotalGas, proposerAddress, block.Timestamp,
	)
	return err
}

// GetTotalBlocks implements database.Database
func (db *Database) GetTotalBlocks() int64 {
	var blockCount int64
	err := db.SQL.QueryRow(`SELECT count(*) FROM block;`).Scan(&blockCount)
	if err != nil {
		return 0
	}

	return blockCount
}

// SaveTx implements database.Database
func (db *Database) SaveTx(tx *types.Tx) error {
	var partitionID int64

	partitionSize := config.Cfg.Database.PartitionSize
	if partitionSize > 0 {
		partitionID = tx.Height / partitionSize
		err := db.CreatePartitionIfNotExists("transaction", partitionID)
		if err != nil {
			return err
		}
	}

	return db.saveTxInsidePartition(tx, partitionID)
}

// saveTxInsidePartition stores the given transaction inside the partition having the given id
func (db *Database) saveTxInsidePartition(tx *types.Tx, partitionID int64) error {
	sqlStatement := `
INSERT INTO transaction 
(hash, height, success, messages, memo, signatures, signer_infos, fee, gas_wanted, gas_used, raw_log, logs, partition_id) 
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13) 
ON CONFLICT (hash, partition_id) DO UPDATE 
	SET height = excluded.height, 
		success = excluded.success, 
		messages = excluded.messages,
		memo = excluded.memo, 
		signatures = excluded.signatures, 
		signer_infos = excluded.signer_infos,
		fee = excluded.fee, 
		gas_wanted = excluded.gas_wanted, 
		gas_used = excluded.gas_used,
		raw_log = excluded.raw_log, 
		logs = excluded.logs`

	var sigs = make([]string, len(tx.Signatures))
	for index, sig := range tx.Signatures {
		sigs[index] = base64.StdEncoding.EncodeToString(sig)
	}

	var msgs = make([]string, len(tx.Body.Messages))
	for index, msg := range tx.Body.Messages {
		bz, err := db.Cdc.MarshalJSON(msg)
		if err != nil {
			return err
		}
		msgs[index] = string(bz)
	}
	msgsBz := fmt.Sprintf("[%s]", strings.Join(msgs, ","))

	feeBz, err := db.Cdc.MarshalJSON(tx.AuthInfo.Fee)
	if err != nil {
		return fmt.Errorf("failed to JSON encode tx fee: %s", err)
	}

	var sigInfos = make([]string, len(tx.AuthInfo.SignerInfos))
	for index, info := range tx.AuthInfo.SignerInfos {
		bz, err := db.Cdc.MarshalJSON(info)
		if err != nil {
			return err
		}
		sigInfos[index] = string(bz)
	}
	sigInfoBz := fmt.Sprintf("[%s]", strings.Join(sigInfos, ","))

	logsBz, err := db.Amino.MarshalJSON(tx.Logs)
	if err != nil {
		return err
	}

	_, err = db.SQL.Exec(sqlStatement,
		tx.TxHash, tx.Height, tx.Successful(),
		msgsBz, tx.Body.Memo, pq.Array(sigs),
		sigInfoBz, string(feeBz),
		tx.GasWanted, tx.GasUsed, tx.RawLog, string(logsBz),
		partitionID,
	)
	return err
}

// HasValidator implements database.Database
func (db *Database) HasValidator(addr string) (bool, error) {
	var res bool
	stmt := `SELECT EXISTS(SELECT 1 FROM validator WHERE consensus_address = $1);`
	err := db.SQL.QueryRow(stmt, addr).Scan(&res)
	return res, err
}

// SaveValidators implements database.Database
func (db *Database) SaveValidators(validators []*types.Validator) error {
	if len(validators) == 0 {
		return nil
	}

	stmt := `INSERT INTO validator (consensus_address, consensus_pubkey) VALUES `

	var vparams []interface{}
	for i, val := range validators {
		vi := i * 2

		stmt += fmt.Sprintf("($%d, $%d),", vi+1, vi+2)
		vparams = append(vparams, val.ConsAddr, val.ConsPubKey)
	}

	stmt = stmt[:len(stmt)-1] // Remove trailing ,
	stmt += " ON CONFLICT DO NOTHING"
	_, err := db.SQL.Exec(stmt, vparams...)
	return err
}

// SaveCommitSignatures implements database.Database
func (db *Database) SaveCommitSignatures(signatures []*types.CommitSig) error {
	if len(signatures) == 0 {
		return nil
	}

	stmt := `INSERT INTO pre_commit (validator_address, height, timestamp, voting_power, proposer_priority) VALUES `

	var sparams []interface{}
	for i, sig := range signatures {
		si := i * 5

		stmt += fmt.Sprintf("($%d, $%d, $%d, $%d, $%d),", si+1, si+2, si+3, si+4, si+5)
		sparams = append(sparams, sig.ValidatorAddress, sig.Height, sig.Timestamp, sig.VotingPower, sig.ProposerPriority)
	}

	stmt = stmt[:len(stmt)-1]
	stmt += " ON CONFLICT (validator_address, timestamp) DO NOTHING"
	_, err := db.SQL.Exec(stmt, sparams...)
	return err
}

// SaveMessage implements database.Database
func (db *Database) SaveMessage(msg *types.Message) error {
	var partitionID int64
	partitionSize := config.Cfg.Database.PartitionSize
	if partitionSize > 0 {
		partitionID = msg.Height / partitionSize
		err := db.CreatePartitionIfNotExists("message", partitionID)
		if err != nil {
			return err
		}
	}

	return db.saveMessageInsidePartition(msg, partitionID)
}

// saveMessageInsidePartition stores the given message inside the partition having the provided id
func (db *Database) saveMessageInsidePartition(msg *types.Message, partitionID int64) error {
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

// Close implements database.Database
func (db *Database) Close() {
	err := db.SQL.Close()
	if err != nil {
		db.Logger.Error("error while closing connection", "err", err)
	}
}

// -------------------------------------------------------------------------------------------------------------------

// GetLastPruned implements database.PruningDb
func (db *Database) GetLastPruned() (int64, error) {
	var lastPrunedHeight int64
	err := db.SQL.QueryRow(`SELECT coalesce(MAX(last_pruned_height),0) FROM pruning LIMIT 1;`).Scan(&lastPrunedHeight)
	return lastPrunedHeight, err
}

// StoreLastPruned implements database.PruningDb
func (db *Database) StoreLastPruned(height int64) error {
	_, err := db.SQL.Exec(`DELETE FROM pruning`)
	if err != nil {
		return err
	}

	_, err = db.SQL.Exec(`INSERT INTO pruning (last_pruned_height) VALUES ($1)`, height)
	return err
}

// Prune implements database.PruningDb
func (db *Database) Prune(height int64) error {
	_, err := db.SQL.Exec(`DELETE FROM pre_commit WHERE height = $1`, height)
	if err != nil {
		return err
	}

	_, err = db.SQL.Exec(`
DELETE FROM message 
USING transaction 
WHERE message.transaction_hash = transaction.hash AND transaction.height = $1
`, height)
	return err
}
