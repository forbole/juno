package db

import (
	"database/sql"
	"encoding/base64"
	"fmt"

	junocdc "github.com/alexanderbez/juno/codec"
	"github.com/alexanderbez/juno/config"
	_ "github.com/lib/pq" // nolint
	"github.com/rs/zerolog/log"
	tmctypes "github.com/tendermint/tendermint/rpc/core/types"
	tmtypes "github.com/tendermint/tendermint/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
)

// Database defines a wrapper around a SQL database and implements functionality
// for data aggregation and exporting.
type Database struct {
	*sql.DB
}

// OpenDB opens a database connection with the given database connection info
// from config. It returns a database connection handle or an error if the
// connection fails.
func OpenDB(cfg config.Config) (*Database, error) {
	sslMode := "disable"
	if cfg.DB.SSLMode != "" {
		sslMode = cfg.DB.SSLMode
	}

	connStr := fmt.Sprintf(
		"host=%s port=%d dbname=%s user=%s sslmode=%s",
		cfg.DB.Host, cfg.DB.Port, cfg.DB.Name, cfg.DB.User, sslMode,
	)

	if cfg.DB.Password != "" {
		connStr += fmt.Sprintf(" password=%s", cfg.DB.Password)
	}

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}

	return &Database{db}, nil
}

// LastBlockHeight returns the latest block stored.
func (db *Database) LastBlockHeight() (int64, error) {
	var height int64
	err := db.QueryRow("SELECT coalesce(MAX(height),0) AS height FROM block;").Scan(&height)
	return height, err
}

// HasBlock returns true if a block by height exists. An error should never be
// returned.
func (db *Database) HasBlock(height int64) (bool, error) {
	var res bool
	err := db.QueryRow(
		"SELECT EXISTS(SELECT 1 FROM block WHERE height = $1);",
		height,
	).Scan(&res)

	return res, err
}

// HasValidator returns true if a given validator by HEX address exists. An
// error should never be returned.
func (db *Database) HasValidator(addr string) (bool, error) {
	var res bool
	err := db.QueryRow(
		"SELECT EXISTS(SELECT 1 FROM validator WHERE address = $1);",
		addr,
	).Scan(&res)

	return res, err
}

// SetValidator stores a validator if it does not already exist. An error is
// returned if the operation fails.
func (db *Database) SetValidator(addr, pk string) error {
	_, err := db.Exec(
		"INSERT INTO validator (address, consensus_pubkey) VALUES ($1, $2) ON CONFLICT DO NOTHING RETURNING id;",
		addr, pk,
	)

	return err
}

// SetPreCommit stores a validator's pre-commit and returns the resulting record
// ID. An error is returned if the operation fails.
func (db *Database) SetPreCommit(pc *tmtypes.CommitSig, vp, pp int64) (uint64, error) {
	var id uint64

	sqlStatement := `
	INSERT INTO pre_commit (height, round, validator_address, timestamp, voting_power, proposer_priority)
	VALUES ($1, $2, $3, $4, $5, $6)
	RETURNING id;
	`

	err := db.QueryRow(
		sqlStatement,
		pc.Height, pc.Round, pc.ValidatorAddress.String(), pc.Timestamp, vp, pp,
	).Scan(&id)

	return id, err
}

// SetBlock stores a block and returns the resulting record ID. An error is
// returned if the operation fails.
func (db *Database) SetBlock(b *tmctypes.ResultBlock, tg, pc uint64) (uint64, error) {
	var id uint64

	sqlStatement := `
	INSERT INTO block (height, hash, num_txs, total_gas, proposer_address, pre_commits, timestamp)
	VALUES ($1, $2, $3, $4, $5, $6, $7)
	RETURNING id;
	`

	err := db.QueryRow(
		sqlStatement,
		b.Block.Height, b.Block.Hash().String(), b.Block.NumTxs,
		tg, b.Block.ProposerAddress.String(), pc, b.Block.Time,
	).Scan(&id)

	return id, err
}

type signature struct {
	Address   string `json:"address,omitempty"`
	Pubkey    string `json:"pubkey,omitempty"`
	Signature string `json:"signature,omitempty"`
}

// SetTx stores a transaction and returns the resulting record ID. An error is
// returned if the operation fails.
func (db *Database) SetTx(tx sdk.TxResponse) (uint64, error) {
	var id uint64

	sqlStatement := `
	INSERT INTO transaction (timestamp, gas_wanted, gas_used, height, txhash, events, messages, fee, signatures, memo)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	RETURNING id;
	`

	stdTx, ok := tx.Tx.(auth.StdTx)
	if !ok {
		return 0, fmt.Errorf("unsupported tx type: %T", tx.Tx)
	}

	eventsBz, err := junocdc.Codec.MarshalJSON(tx.Events)
	if err != nil {
		return 0, fmt.Errorf("failed to JSON encode tx events: %s", err)
	}

	msgsBz, err := junocdc.Codec.MarshalJSON(stdTx.GetMsgs())
	if err != nil {
		return 0, fmt.Errorf("failed to JSON encode tx messages: %s", err)
	}

	feeBz, err := junocdc.Codec.MarshalJSON(stdTx.Fee)
	if err != nil {
		return 0, fmt.Errorf("failed to JSON encode tx fee: %s", err)
	}

	// convert Tendermint signatures into a more human-readable format
	sigs := make([]signature, len(stdTx.GetSignatures()), len(stdTx.GetSignatures()))
	for i, sig := range stdTx.GetSignatures() {
		consPubKey, err := sdk.Bech32ifyConsPub(sig.PubKey) // nolint: typecheck
		if err != nil {
			return 0, fmt.Errorf("failed to convert validator public key %s: %s\n", sig.PubKey, err)
		}

		sigs[i] = signature{
			Address:   sig.Address().String(),
			Signature: base64.StdEncoding.EncodeToString(sig.Signature),
			Pubkey:    consPubKey,
		}
	}

	sigsBz, err := junocdc.Codec.MarshalJSON(sigs)
	if err != nil {
		return 0, fmt.Errorf("failed to JSON encode tx signatures: %s", err)
	}

	err = db.QueryRow(
		sqlStatement,
		tx.Timestamp, tx.GasWanted, tx.GasUsed, tx.Height, tx.TxHash, string(eventsBz),
		string(msgsBz), string(feeBz), string(sigsBz), stdTx.GetMemo(),
	).Scan(&id)

	return id, err
}

// ExportBlock accepts a finalized block and a corresponding set of transactions
// and persists them to the database along with attributable metadata. An error
// is returned if the write fails.
func (db *Database) ExportBlock(b *tmctypes.ResultBlock, txs []sdk.TxResponse, vals *tmctypes.ResultValidators) error {
	totalGas := sumGasTxs(txs)
	preCommits := uint64(len(b.Block.LastCommit.Precommits))

	// Set the block's proposer if it does not already exist. This may occur if
	// the proposer has never signed before.
	proposerAddr := b.Block.ProposerAddress.String()

	val := findValidatorByAddr(proposerAddr, vals)
	if val == nil {
		err := fmt.Errorf("failed to find validator by address %s for block %d", proposerAddr, b.Block.Height)
		log.Error().Str("validator", proposerAddr).Int64("height", b.Block.Height).Msg("failed to find validator by address")
		return err
	}

	if err := db.ExportValidator(val); err != nil {
		return err
	}

	if _, err := db.SetBlock(b, totalGas, preCommits); err != nil {
		log.Error().Err(err).Int64("height", b.Block.Height).Msg("failed to persist block")
		return err
	}

	for _, tx := range txs {
		if _, err := db.SetTx(tx); err != nil {
			log.Error().Err(err).Str("hash", tx.TxHash).Msg("failed to persist transaction")
			return err
		}
	}

	return nil
}

// ExportValidator persists a Tendermint validator with an address and a
// consensus public key. An error is returned if the public key cannot be Bech32
// encoded or if the DB write fails.
func (db *Database) ExportValidator(val *tmtypes.Validator) error {
	valAddr := val.Address.String()

	consPubKey, err := sdk.Bech32ifyConsPub(val.PubKey) // nolint: typecheck
	if err != nil {
		log.Error().Err(err).Str("validator", valAddr).Msg("failed to convert validator public key")
		return err
	}

	if err := db.SetValidator(valAddr, consPubKey); err != nil {
		log.Error().Err(err).Str("validator", valAddr).Msg("failed to persist validator")
		return err
	}

	return nil
}

// ExportPreCommits accepts a block commitment and a coressponding set of
// validators for the commitment and persists them to the database. An error is
// returned if any write fails or if there is any missing aggregated data.
func (db *Database) ExportPreCommits(commit *tmtypes.Commit, vals *tmctypes.ResultValidators) error {
	// persist all validators and pre-commits
	for _, pc := range commit.Precommits {
		if pc != nil {
			valAddr := pc.ValidatorAddress.String()

			val := findValidatorByAddr(valAddr, vals)
			if val == nil {
				err := fmt.Errorf("failed to find validator by address %s for block %d", valAddr, commit.Height())
				log.Error().Msg(err.Error())
				return err
			}

			if err := db.ExportValidator(val); err != nil {
				return err
			}

			if _, err := db.SetPreCommit(pc, val.VotingPower, val.ProposerPriority); err != nil {
				log.Error().Err(err).Str("validator", valAddr).Msg("failed to persist validator pre-commit")
				return err
			}
		}
	}

	return nil
}
