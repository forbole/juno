package main

import (
	"fmt"
	"strconv"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/desmos-labs/desmos/x/posts"
	"github.com/desmos-labs/juno/db"
	"github.com/desmos-labs/juno/db/mongo"
	"github.com/desmos-labs/juno/types"
	"github.com/rs/zerolog/log"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func msgHandler(tx types.Tx, index int, msg sdk.Msg, db db.Database) error {
	if !tx.Logs[index].Success {
		log.Info().Msg(fmt.Sprintf("Skipping message at index %d of tx hash %s as it was not successull",
			index, tx.TxHash))
		return nil
	}

	// MsgCreatePost
	if createPostMsg, ok := msg.(posts.MsgCreatePost); ok {
		log.Info().Str("tx_hash", tx.TxHash).Int("msg_index", index).Msg("Found MsgCreatePost")

		var postID uint64

		// Get the post id
		// TODO: test with multiple MsgCreatePost
		for _, ev := range tx.Events {
			for _, attr := range ev.Attributes {
				if attr.Key == "post_id" {
					postID, _ = strconv.ParseUint(attr.Value, 10, 64)
				}
			}
		}

		mongoDb, ok := db.(mongo.Db)
		if !ok {
			return fmt.Errorf("database is not a MongoDB instance")
		}

		if err := handleMsgCreatePost(postID, createPostMsg, mongoDb); err != nil {
			return err
		}
	}

	// TODO: Add other types as well

	return nil
}

// handleMsgCreatePost handles a MsgCreatePost and saves the post inside the database
func handleMsgCreatePost(postID uint64, msg posts.MsgCreatePost, db mongo.Db) error {
	ctx, cancel := mongo.BuildCtx()
	defer cancel()

	// Convert the post to a BSON document
	post := convertPostToBSONSetDocument(types.NewPostFromMsg(postID, msg))

	collection := db.Mongo.Collection("posts")
	filter := bson.D{{"post_id", postID}}

	_, err := collection.UpdateOne(ctx, filter, post, options.Update().SetUpsert(true))
	if err != nil {
		return err
	}

	return nil
}

// convertPostToBSONSetDocument converts the given post to a BSON document allowing
// it to be saved inside a Mongo collection
func convertPostToBSONSetDocument(post posts.Post) bson.D {
	return bson.D{
		{"$set", post},
	}
}
