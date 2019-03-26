package main

import (
	"database/sql"
	"fmt"
	"log"

	sdk "github.com/cosmos/cosmos-sdk/types"
	_ "github.com/lib/pq" // nolint
	tmctypes "github.com/tendermint/tendermint/rpc/core/types"
	tmtypes "github.com/tendermint/tendermint/types"
)

// database defines a wrapper around a SQL database and implements functionality
// for data aggregation and exporting.
type database struct {
	*sql.DB
}

func openDB(cfg config) (*database, error) {
	connStr := fmt.Sprintf(
		"host=%s port=%d dbname=%s user=%s password=%s sslmode=require",
		cfg.DB.Host, cfg.DB.Port, cfg.DB.Name, cfg.DB.User, cfg.DB.Password,
	)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}

	return &database{db}, nil
}

// lastBlockHeight returns the latest block stored.
func (db *database) lastBlockHeight() (int64, error) {
	var height int64
	err := db.QueryRow("SELECT coalesce(MAX(height),0) AS height FROM block;").Scan(&height)
	return height, err
}

// hasBlock returns true if a block by height exists. An error should never be
// returned.
func (db *database) hasBlock(height int64) (bool, error) {
	var res bool
	err := db.QueryRow(
		"SELECT EXISTS(SELECT 1 FROM block WHERE height = $1);",
		height,
	).Scan(&res)

	return res, err
}

// hasValidator returns true if a given validator by HEX address exists. An
// error should never be returned.
func (db *database) hasValidator(ah string) (bool, error) {
	var res bool
	err := db.QueryRow(
		"SELECT EXISTS(SELECT 1 FROM validator WHERE address = $1);",
		ah,
	).Scan(&res)

	return res, err
}

// setValidator stores a validator and returns the resulting record ID. An error
// is returned if the operation fails.
func (db *database) setValidator(ah, pk string) (uint64, error) {
	var id uint64

	err := db.QueryRow(
		"INSERT INTO validator (address, consensus_pubkey) VALUES ($1, $2) RETURNING id;",
		ah, pk,
	).Scan(&id)

	return id, err
}

// setPreCommit stores a validator's pre-commit and returns the resulting record
// ID. An error is returned if the operation fails.
func (db *database) setPreCommit(pc *tmtypes.CommitSig, vp, pp int64) (uint64, error) {
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

// setBlock stores a block and returns the resulting record ID. An error is
// returned if the operation fails.
func (db *database) setBlock(b *tmctypes.ResultBlock, tg, pc uint64) (uint64, error) {
	var id uint64

	sqlStatement := `
	INSERT INTO block (height, hash, num_txs, total_gas, proposer_address, pre_commits)
	VALUES ($1, $2, $3, $4, $5, $6)
	RETURNING id;
	`

	err := db.QueryRow(
		sqlStatement,
		b.Block.Height, b.Block.Hash().String(), b.Block.NumTxs, tg, b.Block.ProposerAddress.String(), pc,
	).Scan(&id)

	return id, err
}

func (db *database) exportBlock(b *tmctypes.ResultBlock, txs []*tmctypes.ResultTx) error {
	totalGas := sumGasTxs(txs)
	preCommits := uint64(len(b.Block.LastCommit.Precommits))

	if _, err := db.setBlock(b, totalGas, preCommits); err != nil {
		log.Printf("failed to persist block #%d: %s\n", b.Block.Height, err)
		return err
	}

	return nil
}

func (db *database) exportPreCommits(block *tmctypes.ResultBlock, vals *tmctypes.ResultValidators) error {
	// persist all validators and pre-commits
	for _, pc := range block.Block.LastCommit.Precommits {
		if pc != nil {
			valAddr := pc.ValidatorAddress.String()
			ok, err := db.hasValidator(valAddr)
			if err != nil {
				log.Printf("failed to query for validator %s: %s\n", valAddr, err)
				return err
			}

			val := findValidatorByAddr(valAddr, vals)
			if val == nil {
				err := fmt.Errorf("failed to find validator by address %s for block #%d\n", valAddr, block.Block.Height)
				log.Println(err)
				return err
			}

			// persist the validator if we have not seen them before
			if !ok {
				consPubKey, err := sdk.Bech32ifyConsPub(val.PubKey) // nolint: typecheck
				if err != nil {
					log.Printf("failed to convert validator public key %s: %s\n", valAddr, err)
					return err
				}

				if _, err := db.setValidator(valAddr, consPubKey); err != nil {
					log.Printf("failed to persist validator %s: %s\n", valAddr, err)
					return err
				}
			}

			if _, err := db.setPreCommit(pc, val.VotingPower, val.ProposerPriority); err != nil {
				log.Printf("failed to persist pre-commit for validator %s: %s\n", valAddr, err)
			}
		}
	}

	return nil
}
