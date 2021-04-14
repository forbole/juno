package client

import (
	"context"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/rs/zerolog/log"

	"github.com/cosmos/cosmos-sdk/simapp/params"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx"

	"github.com/desmos-labs/juno/types"

	"google.golang.org/grpc"

	tmjson "github.com/tendermint/tendermint/libs/json"
	rpcclient "github.com/tendermint/tendermint/rpc/client"
	httpclient "github.com/tendermint/tendermint/rpc/client/http"
	tmctypes "github.com/tendermint/tendermint/rpc/core/types"

	constypes "github.com/tendermint/tendermint/consensus/types"

	"github.com/desmos-labs/juno/config"
)

// Proxy implements a wrapper around both a Tendermint RPC client and a
// Cosmos Sdk REST client that allows for essential data queries.
type Proxy struct {
	ctx            context.Context
	encodingConfig *params.EncodingConfig

	rpcClient rpcclient.Client

	grpConnection   *grpc.ClientConn
	txServiceClient tx.ServiceClient
}

// NewClientProxy allows to build a new Proxy instance
func NewClientProxy(cfg *config.Config, encodingConfig *params.EncodingConfig) (*Proxy, error) {
	rpcClient, err := httpclient.New(cfg.RPC.Address, "/websocket")
	if err != nil {
		return nil, err
	}

	if err := rpcClient.Start(); err != nil {
		return nil, err
	}

	grpcConnection, err := CreateGrpcConnection(cfg)
	if err != nil {
		return nil, err
	}

	return &Proxy{
		encodingConfig:  encodingConfig,
		ctx:             context.Background(),
		rpcClient:       rpcClient,
		grpConnection:   grpcConnection,
		txServiceClient: tx.NewServiceClient(grpcConnection),
	}, nil
}

// LatestHeight returns the latest block height on the active chain. An error
// is returned if the query fails.
func (cp *Proxy) LatestHeight() (int64, error) {
	status, err := cp.rpcClient.Status(cp.ctx)
	if err != nil {
		return -1, err
	}

	height := status.SyncInfo.LatestBlockHeight
	return height, nil
}

// Block queries for a block by height. An error is returned if the query fails.
func (cp *Proxy) Block(height int64) (*tmctypes.ResultBlock, error) {
	return cp.rpcClient.Block(cp.ctx, &height)
}

// TendermintTx queries for a transaction by hash. An error is returned if the
// query fails.
func (cp *Proxy) TendermintTx(hash string) (*tmctypes.ResultTx, error) {
	hashRaw, err := hex.DecodeString(hash)
	if err != nil {
		return nil, err
	}

	return cp.rpcClient.Tx(cp.ctx, hashRaw, false)
}

// Validators returns all the known Tendermint validators for a given block
// height. An error is returned if the query fails.
func (cp *Proxy) Validators(height int64) (*tmctypes.ResultValidators, error) {
	vals := &tmctypes.ResultValidators{
		BlockHeight: height,
	}

	page := 1
	stop := false
	for !stop {
		result, err := cp.rpcClient.Validators(cp.ctx, &height, &page, nil)
		if err != nil {
			return nil, err
		}
		vals.Validators = append(vals.Validators, result.Validators...)
		vals.Count += result.Count
		vals.Total = result.Total

		page += 1
		stop = vals.Count == vals.Total
	}

	return vals, nil
}

// Genesis returns the genesis state
func (cp *Proxy) Genesis() (*tmctypes.ResultGenesis, error) {
	return cp.rpcClient.Genesis(cp.ctx)
}

// ConsensusState returns the consensus state of the chain
func (cp *Proxy) ConsensusState() (*constypes.RoundStateSimple, error) {
	state, err := cp.rpcClient.ConsensusState(context.Background())
	if err != nil {
		return nil, err
	}

	var data constypes.RoundStateSimple
	err = tmjson.Unmarshal(state.RoundState, &data)
	if err != nil {
		return nil, err
	}
	return &data, nil
}

// SubscribeEvents subscribes to new events with the given query through the RPC
// client with the given subscriber name. A receiving only channel, context
// cancel function and an error is returned. It is up to the caller to cancel
// the context and handle any errors appropriately.
func (cp *Proxy) SubscribeEvents(subscriber, query string) (<-chan tmctypes.ResultEvent, context.CancelFunc, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	eventCh, err := cp.rpcClient.Subscribe(ctx, subscriber, query)
	return eventCh, cancel, err
}

// SubscribeNewBlocks subscribes to the new block event handler through the RPC
// client with the given subscriber name. An receiving only channel, context
// cancel function and an error is returned. It is up to the caller to cancel
// the context and handle any errors appropriately.
func (cp *Proxy) SubscribeNewBlocks(subscriber string) (<-chan tmctypes.ResultEvent, context.CancelFunc, error) {
	return cp.SubscribeEvents(subscriber, "tm.event = 'NewBlock'")
}

// Tx queries for a transaction from the REST client and decodes it into a sdk.Tx
// if the transaction exists. An error is returned if the tx doesn't exist or
// decoding fails.
func (cp *Proxy) Tx(hash string) (*sdk.TxResponse, *tx.Tx, error) {
	res, err := cp.txServiceClient.GetTx(context.Background(), &tx.GetTxRequest{Hash: hash})
	if err != nil {
		return nil, nil, err
	}
	return res.TxResponse, res.Tx, nil
}

// Txs queries for all the transactions in a block. Transactions are returned
// in the sdk.TxResponse format which internally contains an sdk.Tx. An error is
// returned if any query fails.
func (cp *Proxy) Txs(block *tmctypes.ResultBlock) ([]*types.Tx, error) {
	txResponses := make([]*types.Tx, len(block.Block.Txs))
	for i, tmTx := range block.Block.Txs {
		txResponse, txObj, err := cp.Tx(fmt.Sprintf("%X", tmTx.Hash()))
		if err != nil {
			return nil, err
		}

		convTx, err := types.NewTx(txResponse, txObj)
		if err != nil {
			return nil, fmt.Errorf("error converting transaction: %s", err.Error())
		}

		txResponses[i] = convTx
	}

	return txResponses, nil
}

// Stop defers the node stop execution to the RPC client.
func (cp *Proxy) Stop() {
	err := cp.rpcClient.Stop()
	if err != nil {
		log.Fatal().Str("module", "client proxy").Err(err).Msg("error while stopping proxy")
	}
}
