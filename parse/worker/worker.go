package worker

import (
	"encoding/json"
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/desmos-labs/juno/db"
	"github.com/desmos-labs/juno/parse/client"
	"github.com/desmos-labs/juno/types"
	"github.com/rs/zerolog/log"
	tmctypes "github.com/tendermint/tendermint/rpc/core/types"
	tmtypes "github.com/tendermint/tendermint/types"
)

var (
	genesisHandlers []GenesisHandler
	blockHandlers   []BlockHandler
	txHandlers      []TxHandler
	msgHandlers     []MsgHandler
)

// GenesisHandler represents a function that allows to handle the genesis state.
// For convenience of use, the entire current codec, the GenesisDoc, the already-unmarshalled AppState
// and the currently used database will be passed to it.
type GenesisHandler func(codec *codec.Codec, genesisDoc *tmtypes.GenesisDoc, appState map[string]json.RawMessage, w Worker) error

// RegisterGenesisHandler allows to register a new GenesisHandler to be called when a new block is parsed.
// All the registered handlers will be called in order as they are registered (First-In-First-Served).
// Later handlers will not execute if a previous handler returns an error.
func RegisterGenesisHandler(handler GenesisHandler) {
	genesisHandlers = append(genesisHandlers, handler)
}

// BlockHandler represents a function that allows to handle a single block.
// For convenience of use, all the transactions present inside the given block
// and the currently used database will be passed as well.
type BlockHandler func(block *tmctypes.ResultBlock, txs []types.Tx, vals *tmctypes.ResultValidators, w Worker) error

// RegisterBlockHandler allows to register a new BlockHandler to be called when a new block is parsed.
// All the registered handlers will be called in order as they are registered (First-In-First-Served).
// Later handlers will not execute if a previous handler returns an error.
func RegisterBlockHandler(handler BlockHandler) {
	blockHandlers = append(blockHandlers, handler)
}

// TxHandler represents a function that allows to handle a single transaction.
// For convenience of use, the currently used database will be passed as well.
type TxHandler func(tx types.Tx, w Worker) error

// RegisterTxHandler allows to register a new TxHandler to be called when a new transaction is parsed.
// All the registered handlers will be called in order as they are registered (First-In-First-Served).
// Later handlers will not execute if a previous handler returns an error.
func RegisterTxHandler(handler TxHandler) {
	txHandlers = append(txHandlers, handler)
}

// MsgHandler represents a function that allows to handle a single transaction message.
// In order to be able to get the logs of that message, or other useful information, the transaction
// that contains it as well as the index of such message inside the transaction itself will be passed too.
// For convenience of use, the currently used database will be passed too.
type MsgHandler func(tx types.Tx, index int, msg sdk.Msg, w Worker) error

// RegisterMsgHandler allows to register a new MsgHandler to be called when a new message is parsed.
// All the registered handlers will be called in order as they are registered (First-In-First-Served).
// Later handlers will not execute if a previous handler returns an error.
func RegisterMsgHandler(handler MsgHandler) {
	msgHandlers = append(msgHandlers, handler)
}

// Worker defines a job consumer that is responsible for getting and
// aggregating block and associated data and exporting it to a database.
type Worker struct {
	Cdc         *codec.Codec
	ClientProxy client.ClientProxy
	queue       types.Queue
	Db          db.Database
}

// NewWorker allows to create a new Worker implementation.
func NewWorker(cdc *codec.Codec, cp client.ClientProxy, q types.Queue, db db.Database) Worker {
	return Worker{cdc, cp, q, db}
}

// Start starts a worker by listening for new jobs (block heights) from the
// given worker queue. Any failed job is logged and re-enqueued.
func (w Worker) Start() {
	for i := range w.queue {
		log.Debug().Int64("height", i).Msg("processing block")

		if err := w.process(i); err != nil {
			// re-enqueue any failed job
			// TODO: Implement exponential backoff or max retries for a block height.
			go func() {
				log.Error().Err(err).Int64("height", i).Msg("re-enqueueing failed block")
				w.queue <- i
			}()
		}
	}
}

// process defines the job consumer workflow. It will fetch a block for a given
// height and associated metadata and export it to a database. It returns an
// error if any export process fails.
func (w Worker) process(height int64) error {
	exists, err := w.Db.HasBlock(height)
	if err != nil {
		return err
	}

	if exists {
		log.Debug().Int64("height", height).Msg("skipping already exported block")
		return nil
	}

	if height == 1 {
		log.Debug().Msg("Parse response")
		response, err := w.ClientProxy.Genesis()
		if err != nil {
			return err
		}

		return w.HandleGenesis(response.Genesis)
	}

	block, err := w.ClientProxy.Block(height)
	if err != nil {
		log.Error().Err(err).Int64("height", height).Msg("failed to get block")
		return err
	}

	txs, err := w.ClientProxy.Txs(block)
	if err != nil {
		log.Error().Err(err).Int64("height", height).Msg("failed to get transactions for block")
		return err
	}

	// Convert the transaction to a more easy-to-handle type
	var txData = make([]types.Tx, len(txs))
	for index, tx := range txs {
		convTx, err := types.NewTx(tx)
		if err != nil {
			return fmt.Errorf("error handleTx")
		}
		txData[index] = *convTx
	}

	vals, err := w.ClientProxy.Validators(block.Block.LastCommit.Height)
	if err != nil {
		log.Error().Err(err).Int64("height", height).Msg("failed to get validators for block")
		return err
	}

	if err := w.ExportPreCommits(block.Block.LastCommit, vals); err != nil {
		return err
	}

	return w.ExportBlock(block, txData, vals)
}

// HandleGenesis accepts a GenesisDoc and calls all the registered genesis handlers
// in the order in which they have been registered.
func (w Worker) HandleGenesis(genesis *tmtypes.GenesisDoc) error {
	var appState map[string]json.RawMessage
	w.Cdc.MustUnmarshalJSON(genesis.AppState, &appState)

	// Call the block handlers
	for _, handler := range genesisHandlers {
		if err := handler(w.Cdc, genesis, appState, w); err != nil {
			return err
		}
	}

	return nil
}

// ExportPreCommits accepts a block commitment and a corresponding set of
// validators for the commitment and persists them to the database. An error is
// returned if any write fails or if there is any missing aggregated data.
func (w Worker) ExportPreCommits(commit *tmtypes.Commit, vals *tmctypes.ResultValidators) error {
	// persist all validators and pre-commits
	for _, commitSig := range commit.Signatures {
		// Avoid empty commits
		if commitSig.Signature == nil {
			continue
		}

		valAddr := sdk.ConsAddress(commitSig.ValidatorAddress).String()

		val := findValidatorByAddr(valAddr, vals)
		if val == nil {
			err := fmt.Errorf("failed to find validator")
			log.Error().Str("validator", valAddr).Time("commit_timestamp", commitSig.Timestamp).Err(err).Send()
			return err
		}

		if err := w.ExportValidator(val); err != nil {
			return err
		}

		if err := w.Db.SaveCommitSig(commitSig, val.VotingPower, val.ProposerPriority); err != nil {
			log.Error().Err(err).Str("validator", valAddr).Msg("failed to persist validator pre-commit")
			return err
		}
	}

	return nil
}

// ExportValidator persists a Tendermint validator with an address and a
// consensus public key. An error is returned if the public key cannot be Bech32
// encoded or if the DB write fails.
func (w Worker) ExportValidator(val *tmtypes.Validator) error {
	valAddr := sdk.ConsAddress(val.Address).String()

	consPubKey, err := sdk.Bech32ifyPubKey(sdk.Bech32PubKeyTypeConsPub, val.PubKey)
	if err != nil {
		log.Error().Err(err).Str("validator", valAddr).Msg("failed to convert validator public key")
		return err
	}

	if err := w.Db.SaveValidator(valAddr, consPubKey); err != nil {
		log.Error().Err(err).Str("validator", valAddr).Msg("failed to persist validator")
		return err
	}

	return nil
}

// ExportBlock accepts a finalized block and a corresponding set of transactions
// and persists them to the database along with attributable metadata. An error
// is returned if the write fails.
func (w Worker) ExportBlock(b *tmctypes.ResultBlock, txs []types.Tx, vals *tmctypes.ResultValidators) error {
	totalGas := sumGasTxs(txs)
	preCommits := uint64(len(b.Block.LastCommit.Signatures))

	// Set the block's proposer if it does not already exist. This may occur if
	// the proposer has never signed before.
	proposerAddr := sdk.ConsAddress(b.Block.ProposerAddress).String()

	val := findValidatorByAddr(proposerAddr, vals)
	if val == nil {
		err := fmt.Errorf("failed to find validator")
		log.Error().Str("validator", proposerAddr).Int64("height", b.Block.Height).Err(err).Send()
		return err
	}

	if err := w.ExportValidator(val); err != nil {
		return err
	}

	// Save the block
	if err := w.Db.SaveBlock(b, totalGas, preCommits); err != nil {
		log.Error().Err(err).Int64("height", b.Block.Height).Msg("failed to persist block")
		return err
	}

	// Call the block handlers
	for _, handler := range blockHandlers {
		if err := handler(b, txs, vals, w); err != nil {
			return err
		}
	}

	// Export the transactions
	return w.ExportTxs(txs)
}

// ExportTxs accepts a slice of transactions and persists then inside the database.
// An error is returned if the write fails.
func (w Worker) ExportTxs(txs []types.Tx) error {
	// Handle all the transactions inside the block
	for _, tx := range txs {
		// Save the transaction itself
		if err := w.Db.SaveTx(tx); err != nil {
			log.Error().Err(err).Str("hash", tx.TxHash).Msg("failed to handle transaction")
			return err
		}

		// Call the tx handlers
		for _, handler := range txHandlers {
			if err := handler(tx, w); err != nil {
				return err
			}
		}

		// Handle all the messages contained inside the transaction
		for i, msg := range tx.Messages {
			// Call the handlers
			for _, handler := range msgHandlers {
				if err := handler(tx, i, msg, w); err != nil {
					return err
				}
			}
		}
	}

	return nil
}
