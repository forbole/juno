package postgresql

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/rs/zerolog/log"

	"github.com/desmos-labs/juno/config"

	"github.com/cosmos/cosmos-sdk/simapp/params"
	"github.com/lib/pq"

	_ "github.com/lib/pq" // nolint

	"github.com/desmos-labs/juno/db"
	"github.com/desmos-labs/juno/types"
)

// Builder creates a database connection with the given database connection info
// from config. It returns a database connection handle or an error if the
// connection fails.
func Builder(cfg *config.DatabaseConfig, encodingConfig *params.EncodingConfig) (db.Database, error) {
	sslMode := "disable"
	if cfg.SSLMode != "" {
		sslMode = cfg.SSLMode
	}

	schema := "public"
	if cfg.Schema != "" {
		schema = cfg.Schema
	}

	connStr := fmt.Sprintf(
		"host=%s port=%d dbname=%s user=%s sslmode=%s search_path=%s",
		cfg.Host, cfg.Port, cfg.Name, cfg.User, sslMode, schema,
	)

	if cfg.Password != "" {
		connStr += fmt.Sprintf(" password=%s", cfg.Password)
	}

	postgresDb, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}

	// Set max open connections
	postgresDb.SetMaxOpenConns(cfg.MaxOpenConnections)
	postgresDb.SetMaxIdleConns(cfg.MaxIdleConnections)

	return &Database{Sql: postgresDb, EncodingConfig: encodingConfig}, nil
}

// type check to ensure interface is properly implemented
var _ db.Database = &Database{}

// Database defines a wrapper around a SQL database and implements functionality
// for data aggregation and exporting.
type Database struct {
	Sql            *sql.DB
	EncodingConfig *params.EncodingConfig
}

// LastBlockHeight implements db.Database
func (db *Database) LastBlockHeight() (int64, error) {
	var height int64
	err := db.Sql.QueryRow(`SELECT coalesce(MAX(height),0) AS height FROM block;`).Scan(&height)
	return height, err
}

// HasBlock implements db.Database
func (db *Database) HasBlock(height int64) (bool, error) {
	var res bool
	err := db.Sql.QueryRow(`SELECT EXISTS(SELECT 1 FROM block WHERE height = $1);`, height).Scan(&res)
	return res, err
}

// SaveBlock implements db.Database
func (db *Database) SaveBlock(block *types.Block) error {
	sqlStatement := `
INSERT INTO block (height, hash, num_txs, total_gas, proposer_address, timestamp)
VALUES ($1, $2, $3, $4, $5, $6) ON CONFLICT DO NOTHING`

	proposerAddress := sql.NullString{Valid: len(block.ProposerAddress) != 0, String: block.ProposerAddress}
	_, err := db.Sql.Exec(sqlStatement,
		block.Height, block.Hash, block.TxNum, block.TotalGas, proposerAddress, block.Timestamp,
	)
	return err
}

// SaveTx implements db.Database
func (db *Database) SaveTx(tx *types.Tx) error {
	sqlStatement := `
INSERT INTO transaction 
    (hash, height, success, messages, memo, signatures, signer_infos, fee, gas_wanted, gas_used, raw_log, logs) 
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12) ON CONFLICT DO NOTHING`

	var sigs = make([]string, len(tx.Signatures))
	for index, sig := range sigs {
		sigs[index] = sig
	}

	var msgs = make([]string, len(tx.Body.Messages))
	for index, msg := range tx.Body.Messages {
		bz, err := db.EncodingConfig.Marshaler.MarshalJSON(msg)
		if err != nil {
			return err
		}
		msgs[index] = string(bz)
	}
	msgsBz := fmt.Sprintf("[%s]", strings.Join(msgs, ","))

	feeBz, err := db.EncodingConfig.Marshaler.MarshalJSON(tx.AuthInfo.Fee)
	if err != nil {
		return fmt.Errorf("failed to JSON encode tx fee: %s", err)
	}

	var sigInfos = make([]string, len(tx.AuthInfo.SignerInfos))
	for index, info := range tx.AuthInfo.SignerInfos {
		bz, err := db.EncodingConfig.Marshaler.MarshalJSON(info)
		if err != nil {
			return err
		}
		sigInfos[index] = string(bz)
	}
	sigInfoBz := fmt.Sprintf("[%s]", strings.Join(sigInfos, ","))

	logsBz, err := db.EncodingConfig.Amino.MarshalJSON(tx.Logs)
	if err != nil {
		return err
	}

	_, err = db.Sql.Exec(sqlStatement,
		tx.TxHash, tx.Height, tx.Successful(),
		msgsBz, tx.Body.Memo, pq.Array(sigs),
		sigInfoBz, string(feeBz),
		tx.GasWanted, tx.GasUsed, tx.RawLog, string(logsBz),
	)
	return err
}

// HasValidator implements db.Database
func (db *Database) HasValidator(addr string) (bool, error) {
	var res bool
	stmt := `SELECT EXISTS(SELECT 1 FROM validator WHERE consensus_address = $1);`
	err := db.Sql.QueryRow(stmt, addr).Scan(&res)
	return res, err
}

// SaveValidator implements db.Database
func (db *Database) SaveValidator(addr, pk string) error {
	stmt := `INSERT INTO validator (consensus_address, consensus_pubkey) VALUES ($1, $2) ON CONFLICT DO NOTHING;`
	_, err := db.Sql.Exec(stmt, addr, pk)
	return err
}

// SaveCommitSig implements db.Database
func (db *Database) SaveCommitSig(pc *types.CommitSig) error {
	sqlStatement := `
INSERT INTO pre_commit (validator_address, height, timestamp, voting_power, proposer_priority)
VALUES ($1, $2, $3, $4, $5) ON CONFLICT (validator_address, timestamp) DO NOTHING;`

	_, err := db.Sql.Exec(sqlStatement, pc.ValidatorAddress, pc.Height, pc.Timestamp, pc.VotingPower, pc.ProposerPriority)
	return err
}

// SaveMessage implements db.Database
func (db *Database) SaveMessage(msg *types.Message) error {
	stmt := `
INSERT INTO message(transaction_hash, index, type, value, involved_accounts_addresses) 
VALUES ($1, $2, $3, $4, $5)`

	_, err := db.Sql.Exec(stmt, msg.TxHash, msg.Index, msg.Type, msg.Value, pq.Array(msg.Addresses))
	return err
}

// Close implements db.Database
func (db *Database) Close() {
	err := db.Sql.Close()
	if err != nil {
		log.Error().Str("module", "psql database").Err(err).Msg("error while closing connection")
	}
}

// -------------------------------------------------------------------------------------------------------------------

// GetLastPruned implements db.PruningDb
func (db *Database) GetLastPruned() (int64, error) {
	var lastPrunedHeight int64
	err := db.Sql.QueryRow(`SELECT coalesce(MAX(last_pruned_height),0) FROM pruning LIMIT 1;`).Scan(&lastPrunedHeight)
	return lastPrunedHeight, err
}

// StoreLastPruned implements db.PruningDb
func (db *Database) StoreLastPruned(height int64) error {
	_, err := db.Sql.Exec(`DELETE FROM pruning`)
	if err != nil {
		return err
	}

	_, err = db.Sql.Exec(`INSERT INTO pruning (last_pruned_height) VALUES ($1)`, height)
	return err
}

// Prune implements db.PruningDb
func (db *Database) Prune(height int64) error {
	_, err := db.Sql.Exec(`DELETE FROM pre_commit WHERE height = $1`, height)
	if err != nil {
		return err
	}

	_, err = db.Sql.Exec(`
DELETE FROM message 
USING transaction 
WHERE message.transaction_hash = transaction.hash AND transaction.height = $1
`, height)
	return err
}
