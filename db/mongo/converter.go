package mongo

import (
	"encoding/json"
	"fmt"

	"github.com/angelorc/desmos-parser/types"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/desmos-labs/desmos/x/posts"
	"go.mongodb.org/mongo-driver/bson"
)

// ConvertTxToBSONSetDocument converts the given tx into a BSON document
// that allows it to be saved inside a collection
func ConvertTxToBSONSetDocument(codec *codec.Codec, tx types.Tx) (bson.D, error) {
	msgsBz, err := codec.MarshalJSON(tx.Messages)
	if err != nil {
		return nil, fmt.Errorf("failed to JSON encode tx messages: %s", err)
	}

	var msgsData []interface{}
	if err := json.Unmarshal(msgsBz, &msgsData); err != nil {
		panic(fmt.Sprintf("error"))
	}

	txb := bson.D{
		{"timestamp", tx.Timestamp},
		{"gas_wanted", tx.GasWanted},
		{"gas_used", tx.GasUsed},
		{"height", tx.Height},
		{"tx_hash", tx.TxHash},
		{"event", tx.Events},
		{"logs", tx.Logs},
		{"messages", msgsData},
		{"fee", tx.Fee},
		{"signatures", tx.Signatures},
	}

	return bson.D{
		{"$set", txb},
	}, nil
}

// ConvertPostToBSONSetDocument converts the given post to a BSON document allowing
// it to be saved inside a Mongo collection
func ConvertPostToBSONSetDocument(post posts.Post) bson.D {
	return bson.D{
		{"$set", post},
	}
}
