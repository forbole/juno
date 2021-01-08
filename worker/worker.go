package worker

import (
	"encoding/json"
	"fmt"

	"github.com/desmos-labs/juno/logging"
	"github.com/desmos-labs/juno/modules"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/rs/zerolog/log"
	tmctypes "github.com/tendermint/tendermint/rpc/core/types"
	tmtypes "github.com/tendermint/tendermint/types"

	"github.com/desmos-labs/juno/client"
	"github.com/desmos-labs/juno/db"
	"github.com/desmos-labs/juno/types"
)

// Worker defines a job consumer that is responsible for getting and
// aggregating block and associated data and exporting it to a database.
type Worker struct {
	queue   types.Queue
	cdc     *codec.Codec
	cp      *client.Proxy
	db      db.Database
	modules []modules.Module
}

// NewWorker allows to create a new Worker implementation.
func NewWorker(cdc *codec.Codec, q types.Queue, cp *client.Proxy, db db.Database, modules []modules.Module) Worker {
	return Worker{cdc: cdc, cp: cp, queue: q, db: db, modules: modules}
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
	exists, err := w.db.HasBlock(height)
	if err != nil {
		return err
	}

	if exists {
		log.Debug().Int64("height", height).Msg("skipping already exported block")
		return nil
	}

	if height == 1 {
		log.Debug().Msg("Getting genesis")
		response, err := w.cp.Genesis()
		if err != nil {
			return err
		}

		return w.HandleGenesis(response.Genesis)
	}

	block, err := w.cp.Block(height)
	if err != nil {
		log.Error().Err(err).Int64("height", height).Msg("failed to get block")
		return err
	}

	txs, err := w.cp.Txs(block)
	if err != nil {
		log.Error().Err(err).Int64("height", height).Msg("failed to get transactions for block")
		return err
	}

	// Convert the transaction to a more easy-to-handle type
	var txData = make([]*types.Tx, len(txs))
	for index, tx := range txs {
		convTx, err := types.NewTx(tx)
		if err != nil {
			return fmt.Errorf("error handleTx")
		}
		txData[index] = convTx
	}

	vals, err := w.cp.Validators(block.Block.LastCommit.Height)
	if err != nil {
		log.Error().Err(err).Int64("height", height).Msg("failed to get validators for block")
		return err
	}

	err = w.ExportPreCommits(block.Block.LastCommit, vals)
	if err != nil {
		return err
	}

	return w.ExportBlock(block, txData, vals)
}

// HandleGenesis accepts a GenesisDoc and calls all the registered genesis handlers
// in the order in which they have been registered.
func (w Worker) HandleGenesis(genesis *tmtypes.GenesisDoc) error {
	var appState map[string]json.RawMessage
	w.cdc.MustUnmarshalJSON(genesis.AppState, &appState)

	// Call the block handlers
	for _, module := range w.modules {
		if err := module.HandleGenesis(genesis, appState, w.cdc, w.cp, w.db); err != nil {
			logging.LogGenesisError(err)
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

		valAddr := sdk.ConsAddress(commitSig.ValidatorAddress)
		val := findValidatorByAddr(valAddr.String(), vals)
		if val == nil {
			err := fmt.Errorf("failed to find validator")
			log.Error().
				Err(err).
				Int64("height", commit.Height).
				Str("validator_hex", commitSig.ValidatorAddress.String()).
				Str("validator_bech32", valAddr.String()).
				Time("commit_timestamp", commitSig.Timestamp).
				Send()
			return err
		}

		err := w.ExportValidator(val)
		if err != nil {
			return err
		}

		err = w.db.SaveCommitSig(commitSig, val.VotingPower, val.ProposerPriority)
		if err != nil {
			log.Error().
				Err(err).
				Int64("height", commit.Height).
				Str("validator_hex", commitSig.ValidatorAddress.String()).
				Str("validator_bech32", valAddr.String()).
				Msg("failed to persist validator pre-commit")
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

	err = w.db.SaveValidator(valAddr, consPubKey)
	if err != nil {
		log.Error().Err(err).Str("validator", valAddr).Msg("failed to persist validator")
		return err
	}

	return nil
}

// ExportBlock accepts a finalized block and a corresponding set of transactions
// and persists them to the database along with attributable metadata. An error
// is returned if the write fails.
func (w Worker) ExportBlock(b *tmctypes.ResultBlock, txs []*types.Tx, vals *tmctypes.ResultValidators) error {
	totalGas := sumGasTxs(txs)
	preCommits := uint64(len(b.Block.LastCommit.Signatures))

	// Set the block's proposer if it does not already exist. This may occur if
	// the proposer has never signed before.
	proposerAddr := sdk.ConsAddress(b.Block.ProposerAddress)

	val := findValidatorByAddr(proposerAddr.String(), vals)
	if val == nil {
		err := fmt.Errorf("failed to find validator")
		log.Error().
			Err(err).
			Int64("height", b.Block.Height).
			Str("validator_hex", b.Block.ProposerAddress.String()).
			Str("validator_bech32", proposerAddr.String()).
			Time("commit_timestamp", b.Block.Time).
			Send()
		return err
	}

	err := w.ExportValidator(val)
	if err != nil {
		return err
	}

	// Save the block
	err = w.db.SaveBlock(b, totalGas, preCommits)
	if err != nil {
		log.Error().Err(err).Int64("height", b.Block.Height).Msg("failed to persist block")
		return err
	}

	// Call the block handlers
	for _, module := range w.modules {
		err := module.HandleBlock(b, txs, vals, w.cdc, w.cp, w.db)
		if err != nil {
			logging.LogBlockError(err)
		}
	}

	// Export the transactions
	return w.ExportTxs(txs)
}

// ExportTxs accepts a slice of transactions and persists then inside the database.
// An error is returned if the write fails.
func (w Worker) ExportTxs(txs []*types.Tx) error {
	// Handle all the transactions inside the block
	for _, tx := range txs {
		// Save the transaction itself
		err := w.db.SaveTx(tx)
		if err != nil {
			log.Error().Err(err).Str("hash", tx.TxHash).Msg("failed to handle transaction")
			return err
		}

		// Call the tx handlers
		for _, module := range w.modules {
			err := module.HandleTx(tx, w.cdc, w.cp, w.db)
			if err != nil {
				logging.LogTxError(err)
			}
		}

		// Handle all the messages contained inside the transaction
		for i, msg := range tx.Messages {
			// Call the handlers
			for _, module := range w.modules {
				err := module.HandleMsg(i, msg, tx, w.cdc, w.cp, w.db)
				if err != nil {
					logging.LogMsgError(err)
				}
			}
		}
	}

	return nil
}
