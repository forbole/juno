package db

import (
	"fmt"

	"github.com/angelorc/desmos-parser/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/rs/zerolog/log"
	tmctypes "github.com/tendermint/tendermint/rpc/core/types"
)

// Database represents an abstract database that can be used to save data inside it
type Database interface {
	// HasBlock tells whether or not the database has already stored the block having the given height.
	HasBlock(height int64) bool

	// HandleBlock will be called when a new block is parsed.
	// NOTE. For each transaction contained inside the block, HandleTx will be called.
	HandleBlock(b *tmctypes.ResultBlock) error

	// HandleTx will be called to handle each transaction contained inside a block
	// NOTE. For each message present inside this transaction, HandleMsg will called instead.
	HandleTx(tx types.Tx) error

	// HandleMsg will be called to handle each message present inside a transaction
	HandleMsg(tx types.Tx, index int, msg sdk.Msg) error
}

// HandleBlock allows to handle the given block which contains the given txs,
// delegating the work to the given database
func HandleBlock(database Database, block *tmctypes.ResultBlock, txs []sdk.TxResponse) error {
	// Handle the block itself
	if err := database.HandleBlock(block); err != nil {
		log.Error().Err(err).Int64("height", block.Block.Height).Msg("failed to handle block")
		return err
	}

	// Handle all the transactions
	for _, tx := range txs {
		if err := HandleTx(database, tx); err != nil {
			return err
		}
	}

	return nil
}

func HandleTx(database Database, tx sdk.TxResponse) error {
	// Convert the transaction to a more easy-to-handle type
	txData, err := types.NewTx(tx)
	if err != nil {
		return fmt.Errorf("error handleTx")
	}

	// Handle the transaction itself
	if err := database.HandleTx(*txData); err != nil {
		log.Error().Err(err).Str("hash", tx.TxHash).Msg("failed to handle transaction")
		return err
	}

	// Handle all the messages contained inside the transaction
	if err := HandleMsgs(database, *txData); err != nil {
		log.Error().Err(err).Str("hash", txData.TxHash).Msg("failed to handle messages")
		return err
	}

	return nil
}

func HandleMsgs(database Database, tx types.Tx) error {
	for i, msg := range tx.Messages {
		if err := database.HandleMsg(tx, i, msg); err != nil {
			return err
		}
	}

	return nil
}
