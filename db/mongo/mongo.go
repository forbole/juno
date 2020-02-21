package mongo

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/angelorc/desmos-parser/config"
	"github.com/angelorc/desmos-parser/db"
	"github.com/angelorc/desmos-parser/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/desmos-labs/desmos/x/posts"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	tmctypes "github.com/tendermint/tendermint/rpc/core/types"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// type check to ensure interface is properly implemented
var _ db.Database = Db{}

// MongoDb represents a MongoDb instance that relies on a MongoDB instance
type Db struct {
	mongo *mongo.Database
	codec *codec.Codec
	ctx   context.Context
}

func Builder(cfg config.Config, codec *codec.Codec) (*db.Database, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return Open(cfg, codec, ctx)
}

// Open allows to open a new MongoDb instance connection using the specified config
func Open(cfg config.Config, codec *codec.Codec, ctx context.Context) (*db.Database, error) {
	opts := options.Client().ApplyURI(cfg.DB.Uri)
	client, err := mongo.NewClient(opts)
	if err != nil {
		return nil, errors.Wrap(err, "failed to mongodb connections")
	}

	if err = client.Connect(ctx); err != nil {
		return nil, errors.Wrap(err, "failed to mongodb connections")
	}

	if err = client.Ping(ctx, nil); err != nil {
		return nil, errors.Wrap(err, "failed to ping mongodb")
	}

	var mdb db.Database = Db{mongo: client.Database(cfg.DB.Name), codec: codec, ctx: ctx}
	return &mdb, nil
}

// HasBlock implements MongoDb
func (db Db) HasBlock(height int64) bool {
	collection := db.mongo.Collection("blocks")

	if err := collection.FindOne(db.ctx, bson.D{{"height", height}}).Err(); err != nil {
		return false
	}
	return true
}

// SaveBlock implements MongoDb
func (db Db) SaveBlock(b *tmctypes.ResultBlock) error {
	var bsonBlock bson.M
	bytes := db.codec.MustMarshalJSON(b.Block)
	if err := json.Unmarshal(bytes, &bsonBlock); err != nil {
		return err
	}

	// NOTE: Why only this data?
	filter := bson.D{
		{"height", b.Block.Height},
		{"hash", b.Block.Hash().String()},
	}

	update := bson.D{
		{"$set", bsonBlock},
	}

	collection := db.mongo.Collection("blocks")
	if _, err := collection.UpdateOne(db.ctx, filter, update, options.Update().SetUpsert(true)); err != nil {
		return err
	}

	return nil
}

// SaveTx implements MongoDb
func (db Db) SaveTx(tx types.Tx) error {
	filter := bson.D{
		{"height", tx.Height},
		{"tx_hash", tx.TxHash},
	}

	txBSON, err := ConvertTxToBSONSetDocument(db.codec, tx)
	if err != nil {
		return fmt.Errorf("error converting the transaction to a BSON document")
	}

	collection := db.mongo.Collection("transactions")
	if _, err := collection.UpdateOne(context.TODO(), filter, txBSON, options.Update().SetUpsert(true)); err != nil {
		return err
	}

	return nil
}

// SaveMsg implements MongoDb
func (db Db) SaveMsg(tx types.Tx, index int, msg sdk.Msg) error {
	if len(tx.Logs) != len(tx.Messages) {
		log.Error().Msg("msg len is different from logs len")
		return nil
	}

	if !tx.Logs[index].Success {
		log.Info().Msg(fmt.Sprintf("Skipping message at index %d of tx hash %s as it was not successull",
			index, tx.TxHash))
		return nil
	}

	// MsgCreatePost
	if createPostMsg, ok := msg.(posts.MsgCreatePost); ok {
		var postID uint64

		// TODO: test with multiple MsgCreatePost
		for _, ev := range tx.Events {
			for _, attr := range ev.Attributes {
				if attr.Key == "post_id" {
					postID, _ = strconv.ParseUint(attr.Value, 10, 64)
				}
			}
		}

		if err := db.handleMsgCreatePost(postID, createPostMsg); err != nil {
			return err
		}
	}

	return nil
}

func (db *Db) handleMsgCreatePost(postID uint64, msg posts.MsgCreatePost) error {
	// Convert the post to a BSON document
	post := ConvertPostToBSONSetDocument(types.NewPostFromMsg(postID, msg))

	collection := db.mongo.Collection("posts")
	filter := bson.D{{"post_id", postID}}

	_, err := collection.UpdateOne(db.ctx, filter, post, options.Update().SetUpsert(true))
	if err != nil {
		return err
	}

	return nil
}
