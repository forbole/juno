package worker

import (
	"encoding/json"
	"fmt"

	"github.com/cosmos/cosmos-sdk/simapp/params"

	"github.com/desmos-labs/juno/modules"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/rs/zerolog/log"
	tmctypes "github.com/tendermint/tendermint/rpc/core/types"
	tmtypes "github.com/tendermint/tendermint/types"

	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"

	"github.com/desmos-labs/juno/client"
	"github.com/desmos-labs/juno/db"
	"github.com/desmos-labs/juno/types"
)

// Worker defines a job consumer that is responsible for getting and
// aggregating block and associated data and exporting it to a database.
type Worker struct {
	queue          types.HeightQueue
	encodingConfig *params.EncodingConfig
	cp             *client.Proxy
	db             db.Database
	modules        []modules.Module
}

// NewWorker allows to create a new Worker implementation.
func NewWorker(config *Config) Worker {
	return Worker{
		encodingConfig: config.EncodingConfig,
		cp:             config.ClientProxy,
		queue:          config.Queue,
		db:             config.Database,
		modules:        config.Modules,
	}
}

// Start starts a worker by listening for new jobs (block heights) from the
// given worker queue. Any failed job is logged and re-enqueued.
func (w Worker) Start() {
	for i := range w.queue {
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
		log.Debug().Msg("getting genesis")
		response, err := w.cp.Genesis()
		if err != nil {
			log.Error().Err(err).Msg("failed to get genesis")
			return err
		}

		return w.HandleGenesis(response.Genesis)
	}

	log.Debug().Int64("height", height).Msg("processing block")

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

	vals, err := w.cp.Validators(block.Block.LastCommit.Height)
	if err != nil {
		log.Error().Err(err).Int64("height", height).Msg("failed to get validators for block")
		return err
	}

	return w.ExportBlock(block, txs, vals)
}

// HandleGenesis accepts a GenesisDoc and calls all the registered genesis handlers
// in the order in which they have been registered.
func (w Worker) HandleGenesis(genesis *tmtypes.GenesisDoc) error {
	log.Debug().Str("module", "worker").Msg("handling genesis")

	var appState map[string]json.RawMessage
	if err := json.Unmarshal(genesis.AppState, &appState); err != nil {
		return fmt.Errorf("error unmarshalling genesis doc %s: %s", appState, err.Error())
	}

	// Store a new block with height 1
	// Since the genesis has no proposer, we simply take the first validator and use its address as the proposer
	// Also, the number of transactions will be the length of the genesis transactions slice
	var genUtilState genutiltypes.GenesisState
	if err := json.Unmarshal(appState[genutiltypes.ModuleName], &genUtilState); err != nil {
		return fmt.Errorf("error unmarshaling gentuil genesis state: %s", err)
	}

	err := w.db.SaveBlock(types.NewBlock(
		genesis.InitialHeight,
		genesis.AppHash.String(),
		len(genUtilState.GenTxs),
		0,
		"",
		genesis.GenesisTime,
	))
	if err != nil {
		return fmt.Errorf("error while saving genesis block: %s", err)
	}

	// Call the genesis handlers
	for _, module := range w.modules {
		if module, ok := module.(modules.GenesisModule); ok {
			if err := module.HandleGenesis(genesis, appState); err != nil {
				types.LogGenesisError(err)
			}
		}
	}

	return nil
}

// SaveValidators persists a list of Tendermint validators with an address and a
// consensus public key. An error is returned if the public key cannot be Bech32
// encoded or if the DB write fails.
func (w Worker) SaveValidators(vals []*tmtypes.Validator) error {
	var validators = make([]*types.Validator, len(vals))
	for index, val := range vals {
		consAddr := sdk.ConsAddress(val.Address).String()

		consPubKey, err := types.ConvertValidatorPubKeyToBech32String(val.PubKey)
		if err != nil {
			log.Error().Err(err).Str("validator", consAddr).Msg("failed to convert validator public key")
			return err
		}

		validators[index] = types.NewValidator(consAddr, consPubKey)
	}

	err := w.db.SaveValidators(validators)
	if err != nil {
		return fmt.Errorf("error while saving validators: %s", err)
	}

	return nil
}

// ExportBlock accepts a finalized block and a corresponding set of transactions
// and persists them to the database along with attributable metadata. An error
// is returned if the write fails.
func (w Worker) ExportBlock(b *tmctypes.ResultBlock, txs []*types.Tx, vals *tmctypes.ResultValidators) error {
	// Save all validators
	err := w.SaveValidators(vals.Validators)
	if err != nil {
		return err
	}

	// Make sure the proposer exists
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

	// Save the block
	err = w.db.SaveBlock(types.NewBlockFromTmBlock(b, sumGasTxs(txs)))
	if err != nil {
		log.Error().Err(err).Int64("height", b.Block.Height).Msg("failed to persist block")
		return err
	}

	// Save the commits
	err = w.ExportCommit(b.Block.LastCommit, vals)
	if err != nil {
		return err
	}

	// Call the block handlers
	for _, module := range w.modules {
		if module, ok := module.(modules.BlockModule); ok {
			err := module.HandleBlock(b, txs, vals)
			if err != nil {
				types.LogBlockError(err)
			}
		}
	}

	// Export the transactions
	return w.ExportTxs(txs)
}

// ExportCommit accepts a block commitment and a corresponding set of
// validators for the commitment and persists them to the database. An error is
// returned if any write fails or if there is any missing aggregated data.
func (w Worker) ExportCommit(commit *tmtypes.Commit, vals *tmctypes.ResultValidators) error {
	var signatures []*types.CommitSig
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

		signatures = append(signatures, types.NewCommitSig(
			types.ConvertValidatorAddressToBech32String(commitSig.ValidatorAddress),
			val.VotingPower,
			val.ProposerPriority,
			commit.Height,
			commitSig.Timestamp,
		))
	}

	err := w.db.SaveCommitSignatures(signatures)
	if err != nil {
		return fmt.Errorf("error while saving commit signatures: %s", err)
	}

	return nil
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
			if module, ok := module.(modules.TransactionModule); ok {
				err := module.HandleTx(tx)
				if err != nil {
					types.LogTxError(err)
				}
			}
		}

		// Handle all the messages contained inside the transaction
		for i, msg := range tx.Body.Messages {
			var stdMsg sdk.Msg
			err = w.encodingConfig.Marshaler.UnpackAny(msg, &stdMsg)
			if err != nil {
				return err
			}

			// Call the handlers
			for _, module := range w.modules {
				if module, ok := module.(modules.MessageModule); ok {
					err = module.HandleMsg(i, stdMsg, tx)
					if err != nil {
						types.LogMsgError(err)
					}
				}
			}
		}
	}

	return nil
}
