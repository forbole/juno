package cosmos

import "github.com/forbole/juno/v5/types"

// CosmosRepository represents an abstract database that can be used to save data inside it
type CosmosRepository interface {
	// HasBlock tells whether or not the database has already stored the block having the given height.
	// An error is returned if the operation fails.
	HasBlock(height int64) (bool, error)

	// GetLastBlockHeight returns the last block height stored in database..
	// An error is returned if the operation fails.
	GetLastBlockHeight() (int64, error)

	// GetMissingHeights returns a slice of missing block heights between startHeight and endHeight
	GetMissingHeights(startHeight, endHeight int64) []int64

	// GetTotalBlocks returns total number of blocks stored in database.
	GetTotalBlocks() int64

	// SaveTx will be called to save each transaction contained inside a block.
	// An error is returned if the operation fails.
	SaveTx(tx *types.Tx) error

	// HasValidator returns true if a given validator by consensus address exists.
	// An error is returned if the operation fails.
	HasValidator(address string) (bool, error)

	// SaveValidators stores a list of validators if they do not already exist.
	// An error is returned if the operation fails.
	SaveValidators(validators []*types.Validator) error

	// SaveCommitSignatures stores a  slice of validator commit signatures.
	// An error is returned if the operation fails.
	SaveCommitSignatures(signatures []*types.CommitSig) error

	// SaveMessage stores a single message.
	// An error is returned if the operation fails.
	SaveMessage(msg *types.Message) error
}
