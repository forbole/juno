package postgresql

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/desmos-labs/juno/config"

	"github.com/cosmos/cosmos-sdk/simapp/params"
	"github.com/lib/pq"

	_ "github.com/lib/pq" // nolint
	tmctypes "github.com/tendermint/tendermint/rpc/core/types"
	tmtypes "github.com/tendermint/tendermint/types"

	"github.com/desmos-labs/juno/db"
	"github.com/desmos-labs/juno/db/utils"
	"github.com/desmos-labs/juno/types"
)

// OpenDB opens a database connection with the given database connection info
// from config. It returns a database connection handle or an error if the
// connection fails.
func Builder(cfg *config.PostgreSQLConfig, encodingConfig *params.EncodingConfig) (db.Database, error) {
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

// LastBlockHeight returns the latest block stored.
func (db *Database) LastBlockHeight() (int64, error) {
	var height int64
	err := db.Sql.QueryRow("SELECT coalesce(MAX(height),0) AS height FROM block;").Scan(&height)
	return height, err
}

// HasBlock returns true if a block by height exists. An error should never be
// returned.
func (db *Database) HasBlock(height int64) (bool, error) {
	var res bool
	err := db.Sql.QueryRow(
		"SELECT EXISTS(SELECT 1 FROM block WHERE height = $1);",
		height,
	).Scan(&res)

	return res, err
}

// SetBlock stores a block and returns the resulting record ID. An error is
// returned if the operation fails.
func (db *Database) SaveBlock(block *tmctypes.ResultBlock, totalGas, preCommits uint64) error {
	sqlStatement := `
INSERT INTO block (height, hash, num_txs, total_gas, proposer_address, pre_commits, timestamp)
VALUES ($1, $2, $3, $4, $5, $6, $7) ON CONFLICT DO NOTHING`

	_, err := db.Sql.Exec(sqlStatement,
		block.Block.Height, block.Block.Hash().String(), len(block.Block.Txs),
		totalGas, utils.ConvertValidatorAddressToBech32String(block.Block.ProposerAddress), preCommits, block.Block.Time,
	)
	return err
}

// SetTx stores a transaction and returns the resulting record ID. An error is
// returned if the operation fails.
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

// HasValidator returns true if a given validator by HEX address exists. An
// error should never be returned.
func (db *Database) HasValidator(addr string) (bool, error) {
	var res bool
	stmt := `SELECT EXISTS(SELECT 1 FROM validator WHERE consensus_address = $1);`
	err := db.Sql.QueryRow(stmt, addr).Scan(&res)
	return res, err
}

// SetValidator stores a validator if it does not already exist. An error is
// returned if the operation fails.
func (db *Database) SaveValidator(addr, pk string) error {
	stmt := `INSERT INTO validator (consensus_address, consensus_pubkey) VALUES ($1, $2) ON CONFLICT DO NOTHING;`
	_, err := db.Sql.Exec(stmt, addr, pk)
	return err
}

// SetPreCommit stores a validator's pre-commit and returns the resulting record
// ID. An error is returned if the operation fails.
func (db *Database) SaveCommitSig(height int64, pc tmtypes.CommitSig, votingPower, proposerPriority int64) error {
	sqlStatement := `INSERT INTO pre_commit (validator_address, height, timestamp, voting_power, proposer_priority)
					 VALUES ($1, $2, $3, $4, $5);`

	address := utils.ConvertValidatorAddressToBech32String(pc.ValidatorAddress)
	_, err := db.Sql.Exec(sqlStatement, address, height, pc.Timestamp, votingPower, proposerPriority)
	return err
}
