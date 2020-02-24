package mongo

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/angelorc/desmos-parser/config"
	"github.com/angelorc/desmos-parser/db"
	"github.com/angelorc/desmos-parser/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/cosmos/cosmos-sdk/codec"
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
	Mongo *mongo.Database
	Codec *codec.Codec
}

func BuildCtx() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), 5*time.Second)
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

	var mdb db.Database = Db{Mongo: client.Database(cfg.DB.Name), Codec: codec}
	return &mdb, nil
}

// HasBlock implements Database
func (db Db) HasBlock(height int64) bool {
	ctx, cancel := BuildCtx()
	defer cancel()

	collection := db.Mongo.Collection("blocks")

	if err := collection.FindOne(ctx, bson.D{{"height", height}}).Err(); err != nil {
		return false
	}
	return true
}

// SaveBlock implements Database
func (db Db) SaveBlock(b *tmctypes.ResultBlock, _ []types.Tx) error {
	ctx, cancel := BuildCtx()
	defer cancel()

	var bsonBlock bson.M
	bytes := db.Codec.MustMarshalJSON(b.Block)
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

	collection := db.Mongo.Collection("blocks")
	if _, err := collection.UpdateOne(ctx, filter, update, options.Update().SetUpsert(true)); err != nil {
		return err
	}

	return nil
}

// SaveTx implements Database
func (db Db) SaveTx(tx types.Tx) error {
	ctx, cancel := BuildCtx()
	defer cancel()

	filter := bson.D{
		{"height", tx.Height},
		{"tx_hash", tx.TxHash},
	}

	txBSON, err := ConvertTxToBSONSetDocument(db.Codec, tx)
	if err != nil {
		return fmt.Errorf("error converting the transaction to a BSON document")
	}

	collection := db.Mongo.Collection("transactions")
	if _, err := collection.UpdateOne(ctx, filter, txBSON, options.Update().SetUpsert(true)); err != nil {
		return err
	}

	return nil
}

// SaveMsg implements Database
func (db Db) SaveMsg(tx types.Tx, index int, msg sdk.Msg) error {
	if len(tx.Logs) != len(tx.Messages) {
		log.Error().Msg("msg len is different from logs len")
		return nil
	}

	return nil
}
