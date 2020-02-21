package db

import (
	"github.com/angelorc/desmos-parser/config"
	"github.com/angelorc/desmos-parser/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/rs/zerolog/log"
	tmctypes "github.com/tendermint/tendermint/rpc/core/types"
)

// Database represents an abstract database that can be used to save data inside it
type Database interface {
	// HasBlock tells whether or not the database has already stored the block having the given height.
	HasBlock(height int64) bool

	// SaveBlock will be called when a new block is parsed.
	// NOTE. For each transaction contained inside the block, SaveTx will be called as well.
	SaveBlock(b *tmctypes.ResultBlock) error

	// SaveTx will be called to save each transaction contained inside a block
	// NOTE. For each message present inside this transaction, SaveMsg will called as well.
	SaveTx(tx types.Tx) error

	// SaveMsg will be called to save each message present inside a transaction
	SaveMsg(tx types.Tx, index int, msg sdk.Msg) error
}

// Builder represents a method that allows to build any database from a given codec and configuration
type Builder func(config.Config, *codec.Codec) (*Database, error)

// SaveBlock allows to handle the given block which contains the given txs,
// delegating the work to the given database
func SaveBlock(database Database, block *tmctypes.ResultBlock, txs []sdk.TxResponse) error {
	// Handle the block itself
	if err := database.SaveBlock(block); err != nil {
		log.Error().Err(err).Int64("height", block.Block.Height).Msg("failed to handle block")
		return err
	}

	return nil
}

func SaveTx(database Database, txData types.Tx) error {
	// Handle the transaction itself
	if err := database.SaveTx(txData); err != nil {
		log.Error().Err(err).Str("hash", txData.TxHash).Msg("failed to handle transaction")
		return err
	}

	return nil
}
