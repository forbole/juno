package remote

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	constypes "github.com/tendermint/tendermint/consensus/types"
	tmjson "github.com/tendermint/tendermint/libs/json"

	"github.com/forbole/juno/v2/node"

	"github.com/cosmos/cosmos-sdk/types/tx"

	"github.com/forbole/juno/v2/types"

	httpclient "github.com/tendermint/tendermint/rpc/client/http"
	tmctypes "github.com/tendermint/tendermint/rpc/core/types"
	jsonrpcclient "github.com/tendermint/tendermint/rpc/jsonrpc/client"
)

var (
	_ node.Node = &Node{}
)

// Node implements a wrapper around both a Tendermint RPCConfig client and a
// chain SDK REST client that allows for essential data queries.
type Node struct {
	ctx             context.Context
	codec           codec.Marshaler
	client          *httpclient.HTTP
	txServiceClient tx.ServiceClient
}

// NewNode allows to build a new Node instance
func NewNode(cfg *Details, codec codec.Marshaler) (*Node, error) {
	httpClient, err := jsonrpcclient.DefaultHTTPClient(cfg.RPC.Address)
	if err != nil {
		return nil, err
	}

	// Tweak the transport
	httpTransport, ok := (httpClient.Transport).(*http.Transport)
	if !ok {
		return nil, fmt.Errorf("invalid HTTP Transport: %T", httpTransport)
	}
	httpTransport.MaxConnsPerHost = cfg.RPC.MaxConnections

	rpcClient, err := httpclient.NewWithClient(cfg.RPC.Address, "/websocket", httpClient)
	if err != nil {
		return nil, err
	}

	err = rpcClient.Start()
	if err != nil {
		return nil, err
	}

	grpcConnection, err := CreateGrpcConnection(cfg.GRPC)
	if err != nil {
		return nil, err
	}

	return &Node{
		ctx:   context.Background(),
		codec: codec,

		client:          rpcClient,
		txServiceClient: tx.NewServiceClient(grpcConnection),
	}, nil
}

// Genesis implements node.Node
func (cp *Node) Genesis() (*tmctypes.ResultGenesis, error) {
	return cp.client.Genesis(cp.ctx)
}

// ConsensusState implements node.Node
func (cp *Node) ConsensusState() (*constypes.RoundStateSimple, error) {
	state, err := cp.client.ConsensusState(context.Background())
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

// LatestHeight implements node.Node
func (cp *Node) LatestHeight() (int64, error) {
	status, err := cp.client.Status(cp.ctx)
	if err != nil {
		return -1, err
	}

	height := status.SyncInfo.LatestBlockHeight
	return height, nil
}

// Validators implements node.Node
func (cp *Node) Validators(height int64) (*tmctypes.ResultValidators, error) {
	vals := &tmctypes.ResultValidators{
		BlockHeight: height,
	}

	page := 1
	perPage := 100
	stop := false
	for !stop {
		result, err := cp.client.Validators(cp.ctx, &height, &page, &perPage)
		if err != nil {
			return nil, err
		}
		vals.Validators = append(vals.Validators, result.Validators...)
		vals.Count += result.Count
		vals.Total = result.Total

		page += 1
		stop = vals.Count == len(vals.Validators)
	}

	return vals, nil
}

// Block implements node.Node
func (cp *Node) Block(height int64) (*tmctypes.ResultBlock, error) {
	return cp.client.Block(cp.ctx, &height)
}

// BlockResults implements node.Node
func (cp *Node) BlockResults(height int64) (*tmctypes.ResultBlockResults, error) {
	return cp.client.BlockResults(cp.ctx, &height)
}

// Tx implements node.Node
func (cp *Node) Tx(hash string) (*types.Tx, error) {
	res, err := cp.txServiceClient.GetTx(context.Background(), &tx.GetTxRequest{Hash: hash})
	if err != nil {
		return nil, err
	}

	// Decode messages
	for _, msg := range res.Tx.Body.Messages {
		var stdMsg sdk.Msg
		err = cp.codec.UnpackAny(msg, &stdMsg)
		if err != nil {
			return nil, fmt.Errorf("error while unpacking message: %s", err)
		}
	}

	convTx, err := types.NewTx(res.TxResponse, res.Tx)
	if err != nil {
		return nil, fmt.Errorf("error converting transaction: %s", err.Error())
	}

	return convTx, nil
}

// Txs implements node.Node
func (cp *Node) Txs(block *tmctypes.ResultBlock) ([]*types.Tx, error) {
	txResponses := make([]*types.Tx, len(block.Block.Txs))
	for i, tmTx := range block.Block.Txs {
		txResponse, err := cp.Tx(fmt.Sprintf("%X", tmTx.Hash()))
		if err != nil {
			return nil, err
		}

		txResponses[i] = txResponse
	}

	return txResponses, nil
}

// TxSearch implements node.Node
func (cp *Node) TxSearch(query string, page *int, perPage *int, orderBy string) (*tmctypes.ResultTxSearch, error) {
	return cp.client.TxSearch(cp.ctx, query, false, page, perPage, orderBy)
}

// SubscribeEvents implements node.Node
func (cp *Node) SubscribeEvents(subscriber, query string) (<-chan tmctypes.ResultEvent, context.CancelFunc, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	eventCh, err := cp.client.Subscribe(ctx, subscriber, query)
	return eventCh, cancel, err
}

// SubscribeNewBlocks implements node.Node
func (cp *Node) SubscribeNewBlocks(subscriber string) (<-chan tmctypes.ResultEvent, context.CancelFunc, error) {
	return cp.SubscribeEvents(subscriber, "tm.event = 'NewBlock'")
}

// Stop implements node.Node
func (cp *Node) Stop() {
	err := cp.client.Stop()
	if err != nil {
		panic(fmt.Errorf("error while stopping proxy: %s", err))
	}
}
