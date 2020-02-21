package mongo

import (
	"context"
	"fmt"
	"strconv"

	"github.com/angelorc/desmos-parser/config"
	"github.com/angelorc/desmos-parser/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"time"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/desmos-labs/desmos/x/posts"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	tmctypes "github.com/tendermint/tendermint/rpc/core/types"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Database represents a Database instance that relies on a MongoDB instance
type Database struct {
	mongo *mongo.Database
	codec *codec.Codec
}

// Open allows to open a new Database instance connection using the specified config
func Open(cfg config.Config, codec *codec.Codec) (*Database, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

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

	mdb := client.Database(cfg.DB.Name)

	return &Database{mongo: mdb, codec: codec}, nil
}

// HasBlock implements Database
func (db *Database) HasBlock(height int64) bool {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	collection := db.mongo.Collection("blocks")
	if err := collection.FindOne(ctx, bson.D{{"height", height}}).Err(); err != nil {
		return false
	}

	return true
}

// HandleBlock implements Database
func (db *Database) HandleBlock(b *tmctypes.ResultBlock) error {
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

	collection := db.mongo.Collection("blocks")
	if _, err := collection.UpdateOne(ctx, filter, update, options.Update().SetUpsert(true)); err != nil {
		return err
	}

	return nil
}

// HandleTx implements Database
func (db *Database) HandleTx(tx types.Tx) error {
	filter := bson.D{
		{"height", tx.Height},
		{"tx_hash", tx.TxHash},
	}

	txBSON, err := ConvertTxToBSONSetDocument(db.codec, tx)
	if err != nil {
		return fmt.Errorf("Error TX BSON")
	}

	collection := db.mongo.Collection("transactions")
	if _, err := collection.UpdateOne(context.TODO(), filter, txBSON, options.Update().SetUpsert(true)); err != nil {
		return err
	}

	return nil
}

// HandleMsg implements Database
func (db *Database) HandleMsg(tx types.Tx, index int, msg sdk.Msg) error {
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

func (db *Database) handleMsgCreatePost(postID uint64, msg posts.MsgCreatePost) error {
	// Convert the post to a BSON document
	post := ConvertPostToBSONSetDocument(types.NewPostFromMsg(postID, msg))

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	collection := db.mongo.Collection("posts")
	filter := bson.D{{"post_id", postID}}
	_, err := collection.UpdateOne(ctx, filter, post, options.Update().SetUpsert(true))
	if err != nil {
		return err
	}

	return nil
}
