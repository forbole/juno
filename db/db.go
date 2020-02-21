package db

import (
	"context"
	"fmt"
	"github.com/angelorc/desmos-parser/config"
	"github.com/angelorc/desmos-parser/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"strconv"

	"github.com/desmos-labs/desmos/x/posts"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	tmctypes "github.com/tendermint/tendermint/rpc/core/types"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"time"
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

func (db *Database) SetTx(tx sdk.TxResponse) error {
	txData, err := types.NewTx(tx)
	if err != nil {
		return fmt.Errorf("error SetTx")
	}

	filter := bson.D{
		{"height", tx.Height},
		{"tx_hash", tx.TxHash},
	}

	txBSON, err := txData.ToBSON()
	if err != nil {
		return fmt.Errorf("Error TX BSON")
	}

	collection := db.Collection("transactions")
	if _, err := collection.UpdateOne(context.TODO(), filter, txBSON, options.Update().SetUpsert(true)); err != nil {
		return err
	}

	if err := db.SetMsgs(tx, txData.Messages); err != nil {
		log.Error().Err(err).Str("hash", tx.TxHash).Msg("failed to persist messages")
		return err
	}

	return nil
}

func (db *Database) SetMsgs(tx sdk.TxResponse, msgs []sdk.Msg) error {
	for i, msg := range msgs {
		if len(tx.Logs) != len(msgs) {
			log.Error().Msg("msg len is different from logs len")
			return nil
		}

		if !tx.Logs[i].Success {
			return nil
		}

		// MsgCreatePost
		if msg.Type() == "create_post" {
			if _, ok := msg.(posts.MsgCreatePost); !ok {
				err := fmt.Errorf("failed to get msg create post")
				return err
			}

			var postID uint64

			// TODO: test with multiple MsgCreatePost
			for _, ev := range tx.Events {
				for _, attr := range ev.Attributes {
					if attr.Key == "post_id" {
						postID, _ = strconv.ParseUint(attr.Value, 10, 64)
					}
				}
			}

			if err := db.ExportMsgCreatePost(postID, msg.(posts.MsgCreatePost), tx.Timestamp); err != nil {
				return err
			}
		}
	}

	return nil
}

func (db *Database) ExportMsgCreatePost(postID uint64, msg posts.MsgCreatePost, timestamp string) error {
	post, err := types.NewPost(postID, msg, timestamp)
	if err != nil {
		return fmt.Errorf("error on ExportMsgCreatePost")
	}

	filter := bson.D{
		{"post_id", postID},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	collection := db.Collection("posts")
	if _, err := collection.UpdateOne(ctx, filter, post.ToBSON(), options.Update().SetUpsert(true)); err != nil {
		return err
	}

	return nil
}
