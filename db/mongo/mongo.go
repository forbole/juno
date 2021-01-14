package mongo

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	tmtypes "github.com/tendermint/tendermint/types"

	"github.com/desmos-labs/juno/config"
	"github.com/desmos-labs/juno/db"
	"github.com/desmos-labs/juno/types"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/pkg/errors"
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

// Builder allows to create a new MongoDB connection from the given config and codec
func Builder(cfg *config.MongoDBConfig, codec *codec.Codec) (db.Database, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return Open(cfg, codec, ctx)
}

// Open allows to open a new MongoDb instance connection using the specified config
func Open(cfg *config.MongoDBConfig, codec *codec.Codec, ctx context.Context) (db.Database, error) {
	opts := options.Client().ApplyURI(cfg.Uri)
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

	return &Db{Mongo: client.Database(cfg.Name), Codec: codec}, nil
}

// BuildCtx returns a new Mongo context
func BuildCtx() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), 5*time.Second)
}

// HasBlock implements Database
func (db Db) HasBlock(height int64) (bool, error) {
	ctx, cancel := BuildCtx()
	defer cancel()

	collection := db.Mongo.Collection("blocks")
	if err := collection.FindOne(ctx, bson.D{{"height", height}}).Err(); err != nil {
		return false, nil
	}

	return true, nil
}

// SaveBlock implements Database
func (db Db) SaveBlock(b *tmctypes.ResultBlock, totalGas, preCommits uint64) error {
	ctx, cancel := BuildCtx()
	defer cancel()

	var bsonBlock bson.M
	bytes := db.Codec.MustMarshalJSON(b.Block)
	if err := json.Unmarshal(bytes, &bsonBlock); err != nil {
		return err
	}

	filter := bson.D{
		{"height", b.Block.Height},
		{"hash", b.Block.Hash().String()},
	}
	update := ConvertBlockToBSONSetDocument(b, totalGas, preCommits)

	collection := db.Mongo.Collection("blocks")
	if _, err := collection.UpdateOne(ctx, filter, update, options.Update().SetUpsert(true)); err != nil {
		return err
	}

	return nil
}

// SaveTx implements Database
func (db Db) SaveTx(tx *types.Tx) error {
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

// HasValidator returns true if a given validator by HEX address exists. An
// error should never be returned.
func (db Db) HasValidator(addr string) (bool, error) {
	ctx, cancel := BuildCtx()
	defer cancel()

	collection := db.Mongo.Collection("validators")
	if err := collection.FindOne(ctx, bson.D{{"address", addr}}).Err(); err != nil {
		return false, nil
	}

	return true, nil
}

// SetValidator stores a validator if it does not already exist. An error is
// returned if the operation fails.
func (db Db) SaveValidator(addr, pk string) error {
	ctx, cancel := BuildCtx()
	defer cancel()

	filter := bson.D{
		{"address", addr},
	}
	update := ConvertValidatorToBSONSetDocument(addr, pk)

	collection := db.Mongo.Collection("validators")
	if _, err := collection.UpdateOne(ctx, filter, update, options.Update().SetUpsert(true)); err != nil {
		return err
	}

	return nil
}

// SetPreCommit stores a validator's pre-commit and returns the resulting record
// ID. An error is returned if the operation fails.
func (db Db) SaveCommitSig(height int64, commitSig tmtypes.CommitSig, votingPower, proposerPriority int64) error {
	ctx, cancel := BuildCtx()
	defer cancel()

	filter := bson.D{
		{"height", height},
		{"timestamp", commitSig.Timestamp},
		{"validator_address", commitSig.ValidatorAddress.String()},
		{"validator_voting_power", votingPower},
		{"proposer_priority", proposerPriority},
	}
	update := ConvertPrecommitToBSONSetDocument(commitSig, votingPower, proposerPriority)

	collection := db.Mongo.Collection("pre_commits")
	if _, err := collection.UpdateOne(ctx, filter, update, options.Update().SetUpsert(true)); err != nil {
		return err
	}

	return nil
}
