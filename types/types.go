package types

import (
	"encoding/base64"
	"fmt"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/desmos-labs/desmos/x/posts"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"strconv"
	"time"
)

type Tx struct {
	Timestamp string `json:"timestamps"`
	GasWanted int64 `json:"gas_wanted"`
	GasUsed int64 `json:"gas_used"`
	Height int64 `json:"height"`
	TxHash string `json:"tx_hash"`
	Events sdk.StringEvents `json:"events"`
	Logs sdk.ABCIMessageLogs `json:"logs"`
	Messages []sdk.Msg `json:"messages"`
	Fee auth.StdFee `json:"fee"`
	Signatures []Signature `json:"signatures"`
}

func NewTx(tx sdk.TxResponse) (*Tx, error) {
	stdTx, ok := tx.Tx.(auth.StdTx)
	if !ok {
		return nil, fmt.Errorf("unsupported tx type: %T", tx.Tx)
	}

	// convert Tendermint signatures into a more human-readable format
	sigs := make([]Signature, len(stdTx.GetSignatures()), len(stdTx.GetSignatures()))
	for i, sig := range stdTx.GetSignatures() {
		consPubKey, err := sdk.Bech32ifyConsPub(sig.PubKey) // nolint: typecheck
		if err != nil {
			return nil, fmt.Errorf("failed to convert validator public key %s: %s\n", sig.PubKey, err)
		}

		sigs[i] = Signature{
			Address:   sig.Address().String(),
			Signature: base64.StdEncoding.EncodeToString(sig.Signature),
			Pubkey:    consPubKey,
		}
	}

	msgs := stdTx.GetMsgs()

	return &Tx{
		Timestamp:  tx.Timestamp,
		GasWanted:  tx.GasWanted,
		GasUsed:    tx.GasUsed,
		Height:     tx.Height,
		TxHash:     tx.TxHash,
		Events:     tx.Events,
		Logs:       tx.Logs,
		Messages:   msgs,
		Fee:        stdTx.Fee,
		Signatures: sigs,
	}, nil
}

func (tx *Tx) ToBSON() bson.D {
	return bson.D{
		{"$set", tx},
	}
}

type Signature struct {
	Address   string `json:"address,omitempty"`
	Pubkey    string `json:"pubkey,omitempty"`
	Signature string `json:"signature,omitempty"`
}

type Post struct {
	ID                primitive.ObjectID `json:"_id" bson:"_id"`
	PostID            uint64             `json:"post_id" bson:"post_id"`                       // Unique id
	ParentID          uint64             `json:"parent_id" bson:"parent_id"`                   // Post of which this one is a comment
	Message           string             `json:"message" bson:"message"`                       // Message contained inside the post
	Created           time.Time          `json:"created" bson:"created"`                       // Block height at which the post has been created
	LastEdited        uint64             `json:"last_edited" bson:"last_edited"`               // Block height at which the post has been edited the last time
	AllowsComments    bool               `json:"allows_comments" bson:"allows_commets"`        // Tells if users can reference this PostID as the parent
	ExternalReference string             `json:"external_reference" bson:"external_reference"` // Used to know when to display this post
	Owner             string             `json:"owner" bson:"owner"`                           // Creator of the Post
}

func NewPost(postID uint64, msg posts.MsgCreatePost, ts string) (*Post, error) {
	parentId, err := strconv.ParseUint(msg.ParentID.String(), 10, 64)
	if err != nil {
		return nil, fmt.Errorf("error parsing parent id")
	}

	timestamp, err := time.Parse("2006-01-02T15:04:05Z07:00", ts)
	if err != nil {
		return nil, fmt.Errorf("error parsing date")
	}

	return &Post{
		PostID:            postID,
		ParentID:          parentId,
		Message:           msg.Message,
		AllowsComments:    msg.AllowsComments,
		ExternalReference: msg.ExternalReference,
		Owner:             msg.Creator.String(),
		Created:           timestamp,
	}, nil
}

func (p *Post) ToBSON() bson.D {
	return bson.D{
		{"$set", p},
	}
}