package client

import (
	"context"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"google.golang.org/grpc"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/rest"
	rpcclient "github.com/tendermint/tendermint/rpc/client"
	httpclient "github.com/tendermint/tendermint/rpc/client/http"
	tmctypes "github.com/tendermint/tendermint/rpc/core/types"

	"github.com/desmos-labs/juno/config"
)

// Proxy implements a wrapper around both a Tendermint RPC client and a
// CosmosConfig Sdk REST client that allows for essential data queries.
type Proxy struct {
	ctx         context.Context
	legacyAmino *codec.LegacyAmino

	rpcClient     rpcclient.Client
	apiEndpoint   string
	grpConnection *grpc.ClientConn
}

func New(cfg *config.Config, legacyAmino *codec.LegacyAmino) (*Proxy, error) {
	rpcClient, err := httpclient.New(cfg.RPCConfig.Address, "/websocket")
	if err != nil {
		return nil, err
	}

	if err := rpcClient.Start(); err != nil {
		return nil, err
	}

	var grpcOpts []grpc.DialOption
	if cfg.GrpcConfig.Insecure {
		grpcOpts = append(grpcOpts, grpc.WithInsecure())
	}

	grpcConnection, err := grpc.Dial(cfg.GrpcConfig.Address, grpcOpts...)
	if err != nil {
		return nil, err
	}

	return &Proxy{
		legacyAmino:   legacyAmino,
		ctx:           context.Background(),
		rpcClient:     rpcClient,
		apiEndpoint:   cfg.APIConfig.Address,
		grpConnection: grpcConnection,
	}, nil
}

// LatestHeight returns the latest block height on the active chain. An error
// is returned if the query fails.
func (cp Proxy) LatestHeight() (int64, error) {
	status, err := cp.rpcClient.Status(cp.ctx)
	if err != nil {
		return -1, err
	}

	height := status.SyncInfo.LatestBlockHeight
	return height, nil
}

// Block queries for a block by height. An error is returned if the query fails.
func (cp Proxy) Block(height int64) (*tmctypes.ResultBlock, error) {
	return cp.rpcClient.Block(cp.ctx, &height)
}

func (cp Proxy) BlockResults(height int64) (*tmctypes.ResultBlockResults, error) {
	return cp.rpcClient.BlockResults(cp.ctx, &height)
}

// TendermintTx queries for a transaction by hash. An error is returned if the
// query fails.
func (cp Proxy) TendermintTx(hash string) (*tmctypes.ResultTx, error) {
	hashRaw, err := hex.DecodeString(hash)
	if err != nil {
		return nil, err
	}

	return cp.rpcClient.Tx(cp.ctx, hashRaw, false)
}

// Validators returns all the known Tendermint validators for a given block
// height. An error is returned if the query fails.
func (cp Proxy) Validators(height int64) (*tmctypes.ResultValidators, error) {
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
func (cp Proxy) Genesis() (*tmctypes.ResultGenesis, error) {
	return cp.rpcClient.Genesis(cp.ctx)
}

// Stop defers the node stop execution to the RPC client.
func (cp Proxy) Stop() error {
	return cp.rpcClient.Stop()
}

// SubscribeEvents subscribes to new events with the given query through the RPC
// client with the given subscriber name. A receiving only channel, context
// cancel function and an error is returned. It is up to the caller to cancel
// the context and handle any errors appropriately.
func (cp Proxy) SubscribeEvents(subscriber, query string) (<-chan tmctypes.ResultEvent, context.CancelFunc, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	eventCh, err := cp.rpcClient.Subscribe(ctx, subscriber, query)
	return eventCh, cancel, err
}

// SubscribeNewBlocks subscribes to the new block event handler through the RPC
// client with the given subscriber name. An receiving only channel, context
// cancel function and an error is returned. It is up to the caller to cancel
// the context and handle any errors appropriately.
func (cp Proxy) SubscribeNewBlocks(subscriber string) (<-chan tmctypes.ResultEvent, context.CancelFunc, error) {
	return cp.SubscribeEvents(subscriber, "tm.event = 'NewBlock'")
}

// QueryLCD queries the LCD at the given endpoint, and deserializes the result into the given pointer.
// If an error is raised, returns the error.
func (cp Proxy) QueryLCD(endpoint string, ptr interface{}) error {
	resp, err := http.Get(fmt.Sprintf("%s/%s", cp.apiEndpoint, endpoint))
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	bz, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if err := cp.legacyAmino.UnmarshalJSON(bz, ptr); err != nil {
		return err
	}

	return nil
}

// QueryLCDWithHeight should be used when the endpoint of the LCD returns the height of the
// request inside the result. It queries such endpoint, deserializes the result and further the
// result data into the given pointer. It returns the retrieved height as well as any error that
// might have been raised.
func (cp Proxy) QueryLCDWithHeight(endpoint string, ptr interface{}) (int64, error) {
	var result rest.ResponseWithHeight
	err := cp.QueryLCD(endpoint, &result)
	if err != nil {
		return -1, err
	}

	return result.Height, cp.legacyAmino.UnmarshalJSON(result.Result, ptr)
}

// Tx queries for a transaction from the REST client and decodes it into a sdk.Tx
// if the transaction exists. An error is returned if the tx doesn't exist or
// decoding fails.
func (cp Proxy) Tx(hash string) (sdk.TxResponse, error) {
	var response sdk.TxResponse
	if err := cp.QueryLCD(fmt.Sprintf("txs/%s", hash), &response); err != nil {
		return sdk.TxResponse{}, err
	}

	return response, nil
}

// Txs queries for all the transactions in a block. Transactions are returned
// in the sdk.TxResponse format which internally contains an sdk.Tx. An error is
// returned if any query fails.
func (cp Proxy) Txs(block *tmctypes.ResultBlock) ([]sdk.TxResponse, error) {
	txResponses := make([]sdk.TxResponse, len(block.Block.Txs))
	for i, tmTx := range block.Block.Txs {
		txResponse, err := cp.Tx(fmt.Sprintf("%X", tmTx.Hash()))
		if err != nil {
			return nil, err
		}

		txResponses[i] = txResponse
	}

	return txResponses, nil
}
