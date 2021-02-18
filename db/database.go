package db

import (
	"github.com/cosmos/cosmos-sdk/simapp/params"
	tmctypes "github.com/tendermint/tendermint/rpc/core/types"
	tmtypes "github.com/tendermint/tendermint/types"

	"github.com/desmos-labs/juno/config"
	"github.com/desmos-labs/juno/types"
)

// Database represents an abstract database that can be used to save data inside it
type Database interface {
	// HasBlock tells whether or not the database has already stored the block having the given height.
	// An error is returned if the operation fails.
	HasBlock(height int64) (bool, error)

	// SaveBlock will be called when a new block is parsed, passing the block itself
	// and the transactions contained inside that block.
	// An error is returned if the operation fails.
	// NOTE. For each transaction inside txs, SaveTx will be called as well.
	SaveBlock(block *tmctypes.ResultBlock, totalGas uint64) error

	// SaveTx will be called to save each transaction contained inside a block.
	// An error is returned if the operation fails.
	SaveTx(tx *types.Tx) error

	// HasValidator returns true if a given validator by consensus address exists.
	// An error is returned if the operation fails.
	HasValidator(address string) (bool, error)

	// SetValidator stores a validator if it does not already exist.
	// An error is returned if the operation fails.
	// The address should be the consensus address of the validator.
	SaveValidator(address, publicKey string) error

	// SaveCommitSig stores a validator's commit signature.
	// An error is returned if the operation fails.
	SaveCommitSig(height int64, commitSig tmtypes.CommitSig, votingPower, proposerPriority int64) error

	// SaveMessage stores a single message.
	// An error is returned if the operation fails.
	SaveMessage(msg *types.Message) error

	// Close closes the connection to the database
	Close()
}

// PruningDb represents a database that supports pruning properly
type PruningDb interface {
	// Prune prunes the data for the given height, returning any error
	Prune(height int64) error

	// StoreLastPruned saves the last height at which the database was pruned
	StoreLastPruned(height int64) error

	// GetLastPruned returns the last height at which the database was pruned
	GetLastPruned() (int64, error)
}

// Create represents a method that allows to build any database from a given codec and configuration
type Builder func(cfg *config.Config, encodingConfig *params.EncodingConfig) (Database, error)
