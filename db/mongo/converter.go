package mongo

import (
	"encoding/json"
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/desmos-labs/juno/types"
	tmctypes "github.com/tendermint/tendermint/rpc/core/types"
	tmtypes "github.com/tendermint/tendermint/types"
	"go.mongodb.org/mongo-driver/bson"
)

func ConvertBlockToBSONSetDocument(block *tmctypes.ResultBlock, totalGas, preCommits uint64) bson.D {
	return bson.D{
		{"$set", bson.D{
			{"height", block.Block.Height},
			{"hash", block.Block.Hash().String()},
			{"num_txs", len(block.Block.Txs)},
			{"total_gas", totalGas},
			{"proposer_address", block.Block.ProposerAddress.String()},
			{"pre_commits", preCommits},
			{"timestamp", block.Block.Time},
		}},
	}
}

// ConvertTxToBSONSetDocument converts the given tx into a BSON document
// that allows it to be saved inside a collection
func ConvertTxToBSONSetDocument(codec *codec.Codec, tx *types.Tx) (bson.D, error) {
	msgsBz, err := codec.MarshalJSON(tx.Messages)
	if err != nil {
		return nil, fmt.Errorf("failed to JSON encode tx messages: %s", err)
	}

	var msgsData []interface{}
	if err := json.Unmarshal(msgsBz, &msgsData); err != nil {
		panic(fmt.Sprintf("error"))
	}

	return bson.D{
		{"$set", bson.D{
			{"timestamp", tx.Timestamp},
			{"gas_wanted", tx.GasWanted},
			{"gas_used", tx.GasUsed},
			{"height", tx.Height},
			{"tx_hash", tx.TxHash},
			{"logs", tx.Logs},
			{"messages", msgsData},
			{"fee", tx.Fee},
			{"signatures", tx.Signatures},
			{"memo", tx.Memo},
		}},
	}, nil
}

func ConvertValidatorToBSONSetDocument(address, publicKey string) bson.D {
	return bson.D{
		{"$set", bson.D{
			{"address", address},
			{"public_key", publicKey},
		}},
	}
}

func ConvertPrecommitToBSONSetDocument(commitSig tmtypes.CommitSig, votingPower, proposerPriority int64) bson.D {
	return bson.D{{
		"$set", bson.D{
			{"validator_address", commitSig.ValidatorAddress.String()},
			{"timestamp", commitSig.Timestamp},
			{"voting_power", votingPower},
			{"proposer_priority", proposerPriority},
		}},
	}
}
