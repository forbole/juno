package db

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/angelorc/desmos-parser/config"
	"github.com/angelorc/desmos-parser/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"strconv"

	"github.com/desmos-labs/desmos/x/posts"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	tmctypes "github.com/tendermint/tendermint/rpc/core/types"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"time"

	desmoscdc "github.com/angelorc/desmos-parser/codec"
)

type Database struct {
	*mongo.Database
}

func OpenDB(cfg config.Config) (*Database, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	options := options.Client().ApplyURI(cfg.DB.Uri)
	client, err := mongo.NewClient(options)
	if err != nil {
		return nil, errors.Wrap(err, "failed to mongodb connections")
	}

	if err = client.Connect(ctx); err != nil {
		return nil, errors.Wrap(err, "failed to mongodb connections")
	}

	if err = client.Ping(ctx, nil); err != nil {
		return nil, errors.Wrap(err, "failed to ping mongodb")
	}

	mdb := client.Database(cfg.DB.Name)

	return &Database{mdb}, nil
}

func (db *Database) HasBlock(height int64) (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	collection := db.Collection("blocks")
	if err := collection.FindOne(ctx, bson.D{{"height", height}}).Err(); err != nil {
		return false, err
	}

	return true, nil
}

func (db *Database) ExportBlock(b *tmctypes.ResultBlock, txs []sdk.TxResponse) error {
	if err := db.SetBlock(b); err != nil {
		log.Error().Err(err).Int64("height", b.Block.Height).Msg("failed to persist block")
		return err
	}

	for _, tx := range txs {
		if err := db.SetTx(tx); err != nil {
			log.Error().Err(err).Str("hash", tx.TxHash).Msg("failed to persist transaction")
			return err
		}
	}

	return nil
}

func (db *Database) SetBlock(b *tmctypes.ResultBlock) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	filter := bson.D{
		{"height", b.Block.Height},
		{"hash", b.Block.Hash().String()},
	}

	update := bson.D{
		{"$set", bson.D{
			{"height", b.Block.Height},
			{"hash", b.Block.Hash().String()},
			{"num_txs", b.Block.NumTxs},
			{"proposer_address", b.Block.ProposerAddress.String()},
			{"timestamp", b.Block.Time},
		}},
	}

	collection := db.Collection("blocks")
	if _, err := collection.UpdateOne(ctx, filter, update, options.Update().SetUpsert(true)); err != nil {
		return err
	}

	return nil
}

type signature struct {
	Address   string `json:"address,omitempty"`
	Pubkey    string `json:"pubkey,omitempty"`
	Signature string `json:"signature,omitempty"`
}

func (db *Database) SetTx(tx sdk.TxResponse) error {
	stdTx, ok := tx.Tx.(auth.StdTx)
	if !ok {
		return fmt.Errorf("unsupported tx type: %T", tx.Tx)
	}

	logsBz, err := desmoscdc.Codec.MarshalJSON(tx.Logs)
	var logsData []interface{}

	if err != nil {
		return fmt.Errorf("failed to JSON encode tx logs: %s", err)
	}
	if err := json.Unmarshal(logsBz, &logsData); err != nil {
		return fmt.Errorf("failed to JSON unmarshal tx logs: %s", err)
	}

	eventsBz, err := desmoscdc.Codec.MarshalJSON(tx.Events)
	var eventsData []interface{}
	if err != nil {
		return fmt.Errorf("failed to JSON encode tx events: %s", err)
	}
	if err := json.Unmarshal(eventsBz, &eventsData); err != nil {
		panic(fmt.Sprintf("error"))
	}

	msgsBz, err := desmoscdc.Codec.MarshalJSON(stdTx.GetMsgs())
	if err != nil {
		return fmt.Errorf("failed to JSON encode tx messages: %s", err)
	}
	var msgsData []interface{}
	if err := json.Unmarshal(msgsBz, &msgsData); err != nil {
		panic(fmt.Sprintf("error"))
	}

	feeBz, err := desmoscdc.Codec.MarshalJSON(stdTx.Fee)
	var feesData interface{}
	if err != nil {
		return fmt.Errorf("failed to JSON encode tx fee: %s", err)
	}
	if err := json.Unmarshal(feeBz, &feesData); err != nil {
		panic(err)
	}

	// convert Tendermint signatures into a more human-readable format
	sigs := make([]signature, len(stdTx.GetSignatures()), len(stdTx.GetSignatures()))
	for i, sig := range stdTx.GetSignatures() {
		consPubKey, err := sdk.Bech32ifyConsPub(sig.PubKey) // nolint: typecheck
		if err != nil {
			return fmt.Errorf("failed to convert validator public key %s: %s\n", sig.PubKey, err)
		}

		sigs[i] = signature{
			Address:   sig.Address().String(),
			Signature: base64.StdEncoding.EncodeToString(sig.Signature),
			Pubkey:    consPubKey,
		}
	}

	sigsBz, err := desmoscdc.Codec.MarshalJSON(sigs)
	var sigsData []interface{}
	if err != nil {
		return fmt.Errorf("failed to JSON encode tx signatures: %s", err)
	}
	if err := json.Unmarshal(sigsBz, &sigsData); err != nil {
		panic(fmt.Sprintf("error"))
	}

	filter := bson.D{
		{"height", tx.Height},
		{"tx_hash", tx.TxHash},
	}

	update := bson.D{
		{"$set", bson.D{
			{"timestamp", tx.Timestamp},
			{"gas_wanted", tx.GasWanted},
			{"gas_used", tx.GasUsed},
			{"height", tx.Height},
			{"tx_hash", tx.TxHash},
			{"events", eventsData},
			{"logs", logsData},
			{"messages", msgsData},
			{"fee", feesData},
			{"signatures", sigsData},
		}},
	}

	collection := db.Collection("transactions")
	if _, err := collection.UpdateOne(context.TODO(), filter, update, options.Update().SetUpsert(true)); err != nil {
		return err
	}

	// messages
	msgs := stdTx.GetMsgs()
	for i, msg := range msgs {
		if len(tx.Logs) != len(msgs) {
			log.Error().Msg("msg len is different from logs len")
			return nil
		}

		if !tx.Logs[i].Success {
			return nil
		}

		if msg.Type() == "create_post" {
			if _, ok := msg.(posts.MsgCreatePost); !ok {
				err := fmt.Errorf("failed to get msg create post")
				return err
			}

			if err := db.ExportMsgCreatePost(msg.(posts.MsgCreatePost), tx.Timestamp); err != nil {
				return err
			}
		}
	}

	return nil
}

func (db *Database) ExportMsgCreatePost(msg posts.MsgCreatePost, timestamp string) error {
	// TODO: add post id

	parentId, err := strconv.ParseUint(msg.ParentID.String(), 10, 64)
	if err != nil {
		return fmt.Errorf("error parsing parent id")
	}

	t, err := time.Parse("2006-01-02T15:04:05Z07:00", timestamp)
	if err != nil {
		return fmt.Errorf("error parsing date")
	}

	post := types.Post{
		ID:                primitive.NewObjectID(),
		ParentID:          parentId,
		Message:           msg.Message,
		AllowsComments:    msg.AllowsComments,
		ExternalReference: msg.ExternalReference,
		Owner:             msg.Creator.String(),
		Created:           t,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	collection := db.Collection("posts")
	if _, err := collection.InsertOne(ctx, post); err != nil {
		return err
	}

	return nil
}
