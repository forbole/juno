package cosmos

import (
	"fmt"

	tmctypes "github.com/cometbft/cometbft/rpc/core/types"
	tmtypes "github.com/cometbft/cometbft/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/authz"
	"github.com/cosmos/gogoproto/proto"
	"github.com/forbole/juno/v5/interfaces"
	"github.com/forbole/juno/v5/logging"
	"github.com/forbole/juno/v5/types"
)

var _ interfaces.BlockModule = &Module{}

func (m *Module) HandleBlock(block interfaces.Block) error {
	return m.Process(block)
}

// Process fetches  a block for a given height and associated metadata and export it to a database.
// It returns an error if any export process fails.
func (m *Module) Process(block interfaces.Block) error {
	height := block.Height()
	resultBlock, ok := block.Value().(*tmctypes.ResultBlock)
	if !ok {
		return fmt.Errorf("invalid block type: %T", block)
	}

	events, err := m.source.BlockResults(height)
	if err != nil {
		return fmt.Errorf("failed to get block results from node: %s", err)
	}

	txs, err := m.source.Txs(resultBlock)
	if err != nil {
		return fmt.Errorf("failed to get transactions for block: %s", err)
	}

	vals, err := m.source.Validators(height)
	if err != nil {
		return fmt.Errorf("failed to get validators for block: %s", err)
	}

	return m.ExportBlock(resultBlock, events, txs, vals)
}

// ProcessTransactions fetches transactions for a given height and stores them into the database.
// It returns an error if the export process fails.
func (m *Module) ProcessTransactions(height int64) error {
	block, err := m.source.ResultBlock(height)
	if err != nil {
		return fmt.Errorf("failed to get block from node: %s", err)
	}

	txs, err := m.source.Txs(block)
	if err != nil {
		return fmt.Errorf("failed to get transactions for block: %s", err)
	}

	return m.ExportTxs(txs)
}

// SaveValidators persists a list of Tendermint validators with an address and a
// consensus public key. An error is returned if the public key cannot be Bech32
// encoded or if the DB write fails.
func (m *Module) SaveValidators(vals []*tmtypes.Validator) error {
	var validators = make([]*types.Validator, len(vals))
	for index, val := range vals {
		consAddr := sdk.ConsAddress(val.Address).String()

		consPubKey, err := types.ConvertValidatorPubKeyToBech32String(val.PubKey)
		if err != nil {
			return fmt.Errorf("failed to convert validator public key for validators %s: %s", consAddr, err)
		}

		validators[index] = types.NewValidator(consAddr, consPubKey)
	}

	err := m.db.SaveValidators(validators)
	if err != nil {
		return fmt.Errorf("error while saving validators: %s", err)
	}

	return nil
}

// ExportBlock accepts a finalized block and a corresponding set of transactions
// and persists them to the database along with attributable metadata. An error
// is returned if the write fails.
func (m *Module) ExportBlock(
	rb *tmctypes.ResultBlock, r *tmctypes.ResultBlockResults, txs []*types.Tx, vals *tmctypes.ResultValidators,
) error {
	// Save all validators
	err := m.SaveValidators(vals.Validators)
	if err != nil {
		return err
	}

	// Make sure the proposer exists
	proposerAddr := sdk.ConsAddress(rb.Block.ProposerAddress)
	val := findValidatorByAddr(proposerAddr.String(), vals)
	if val == nil {
		return fmt.Errorf("failed to find validator by proposer address %s: %s", proposerAddr.String(), err)
	}

	// Save the commits
	err = m.ExportCommit(rb.Block.LastCommit, vals)
	if err != nil {
		return err
	}

	// Export the transactions
	return m.ExportTxs(txs)
}

// ExportCommit accepts a block commitment and a corresponding set of
// validators for the commitment and persists them to the database. An error is
// returned if any write fails or if there is any missing aggregated data.
func (m *Module) ExportCommit(commit *tmtypes.Commit, vals *tmctypes.ResultValidators) error {
	var signatures []*types.CommitSig
	for _, commitSig := range commit.Signatures {
		// Avoid empty commits
		if commitSig.Signature == nil {
			continue
		}

		valAddr := sdk.ConsAddress(commitSig.ValidatorAddress)
		val := findValidatorByAddr(valAddr.String(), vals)
		if val == nil {
			return fmt.Errorf("failed to find validator by commit validator address %s", valAddr.String())
		}

		signatures = append(signatures, types.NewCommitSig(
			types.ConvertValidatorAddressToBech32String(commitSig.ValidatorAddress),
			val.VotingPower,
			val.ProposerPriority,
			commit.Height,
			commitSig.Timestamp,
		))
	}

	err := m.db.SaveCommitSignatures(signatures)
	if err != nil {
		return fmt.Errorf("error while saving commit signatures: %s", err)
	}

	return nil
}

// saveTx accepts the transaction and persists it inside the database.
// An error is returned if the write fails.
func (m *Module) saveTx(tx *types.Tx) error {
	err := m.db.SaveTx(tx)
	if err != nil {
		return fmt.Errorf("failed to handle transaction with hash %s: %s", tx.TxHash, err)
	}
	return nil
}

// handleTx accepts the transaction and calls the tx handlers.
func (m *Module) handleTx(tx *types.Tx) {
	// Call the tx handlers
	for _, module := range m.modules {
		if transactionModule, ok := module.(TransactionModule); ok {
			err := transactionModule.HandleTx(tx)
			if err != nil {
				m.logger.Error("error while handling transaction",
					"err", err,
					"module", module.Name(),
					"height", tx.Height,
					"tx_hash", tx.TxHash,
				)
			}
		}
	}
}

// handleMessage accepts the transaction and handles messages contained
// inside the transaction.
func (m *Module) handleMessage(index int, msg sdk.Msg, tx *types.Tx) {
	// Allow modules to handle the message
	for _, module := range m.modules {
		if messageModule, ok := module.(MessageModule); ok {
			err := messageModule.HandleMsg(index, msg, tx)
			if err != nil {
				m.logger.Error("error while handling message",
					"err", err,
					"module", module.Name(),
					"height", tx.Height,
					"tx_hash", tx.TxHash,
					"msg_type", proto.MessageName(msg),
				)
			}
		}
	}

	// If it's a MsgExecute, we need to make sure the included messages are handled as well
	if msgExec, ok := msg.(*authz.MsgExec); ok {
		for authzIndex, msgAny := range msgExec.Msgs {
			var executedMsg sdk.Msg
			err := m.codec.UnpackAny(msgAny, &executedMsg)
			if err != nil {
				m.logger.Error("unable to unpack MsgExec inner message", "index", authzIndex, "error", err)
			}

			for _, module := range m.modules {
				if messageModule, ok := module.(AuthzMessageModule); ok {
					err = messageModule.HandleMsgExec(index, msgExec, authzIndex, executedMsg, tx)
					if err != nil {
						m.logger.Error("error while handling message",
							"err", err,
							"module", module.Name(),
							"height", tx.Height,
							"tx_hash", tx.TxHash,
							"msg_type", proto.MessageName(executedMsg),
						)
					}
				}
			}
		}
	}
}

// ExportTxs accepts a slice of transactions and persists then inside the database.
// An error is returned if the write fails.
func (m *Module) ExportTxs(txs []*types.Tx) error {
	// handle all transactions inside the block
	for _, tx := range txs {
		// save the transaction
		err := m.saveTx(tx)
		if err != nil {
			return fmt.Errorf("error while storing txs: %s", err)
		}

		// call the tx handlers
		m.handleTx(tx)

		// handle all messages contained inside the transaction
		sdkMsgs := make([]sdk.Msg, len(tx.Body.Messages))
		for i, msg := range tx.Body.Messages {
			var stdMsg sdk.Msg
			err := m.codec.UnpackAny(msg, &stdMsg)
			if err != nil {
				return err
			}
			sdkMsgs[i] = stdMsg
		}

		// call the msg handlers
		for i, sdkMsg := range sdkMsgs {
			m.handleMessage(i, sdkMsg, tx)
		}
	}

	totalBlocks := m.db.GetTotalBlocks()
	logging.DbBlockCount.WithLabelValues("total_blocks_in_db").Set(float64(totalBlocks))

	dbLatestHeight, err := m.db.GetLastBlockHeight()
	if err != nil {
		return err
	}
	logging.DbLatestHeight.WithLabelValues("db_latest_height").Set(float64(dbLatestHeight))

	return nil
}
