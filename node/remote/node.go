package remote

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	tmtypes "github.com/cometbft/cometbft/types"

	constypes "github.com/cometbft/cometbft/consensus/types"
	tmjson "github.com/cometbft/cometbft/libs/json"

	"github.com/forbole/juno/v5/node"

	"github.com/forbole/juno/v5/types"

	httpclient "github.com/cometbft/cometbft/rpc/client/http"
	tmctypes "github.com/cometbft/cometbft/rpc/core/types"
	jsonrpcclient "github.com/cometbft/cometbft/rpc/jsonrpc/client"
)

var (
	_ node.Node = &Node{}
)

// Node implements a wrapper around both a Tendermint RPCConfig client and a
// chain SDK REST client that allows for essential data queries.
type Node struct {
	ctx          context.Context
	client       *httpclient.HTTP
	txServiceAPI string
}

// NewNode allows to build a new Node instance
func NewNode(cfg *Details) (*Node, error) {
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

	return &Node{
		ctx: context.Background(),

		client:       rpcClient,
		txServiceAPI: cfg.API.Address,
	}, nil
}

// Genesis implements node.Node
func (cp *Node) Genesis() (*tmctypes.ResultGenesis, error) {
	res, err := cp.client.Genesis(cp.ctx)
	if err != nil && strings.Contains(err.Error(), "use the genesis_chunked API instead") {
		return cp.getGenesisChunked()
	}
	return res, err
}

// getGenesisChunked gets the genesis data using the chinked API instead
func (cp *Node) getGenesisChunked() (*tmctypes.ResultGenesis, error) {
	bz, err := cp.getGenesisChunksStartingFrom(0)
	if err != nil {
		return nil, err
	}

	var genDoc *tmtypes.GenesisDoc
	err = tmjson.Unmarshal(bz, &genDoc)
	if err != nil {
		return nil, err
	}

	return &tmctypes.ResultGenesis{Genesis: genDoc}, nil
}

// getGenesisChunksStartingFrom returns all the genesis chunks data starting from the chunk with the given id
func (cp *Node) getGenesisChunksStartingFrom(id uint) ([]byte, error) {
	res, err := cp.client.GenesisChunked(cp.ctx, id)
	if err != nil {
		return nil, fmt.Errorf("error while getting genesis chunk %d out of %d", id, res.TotalChunks)
	}

	bz, err := base64.StdEncoding.DecodeString(res.Data)
	if err != nil {
		return nil, fmt.Errorf("error while decoding genesis chunk %d out of %d", id, res.TotalChunks)
	}

	if id == uint(res.TotalChunks-1) {
		return bz, nil
	}

	nextChunk, err := cp.getGenesisChunksStartingFrom(id + 1)
	if err != nil {
		return nil, err
	}

	return append(bz, nextChunk...), nil
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

// ChainID implements node.Node
func (cp *Node) ChainID() (string, error) {
	status, err := cp.client.Status(cp.ctx)
	if err != nil {
		return "", err
	}

	chainID := status.NodeInfo.Network
	return chainID, err
}

// Validators implements node.Node
func (cp *Node) Validators(height int64) (*tmctypes.ResultValidators, error) {
	vals := &tmctypes.ResultValidators{
		BlockHeight: height,
	}

	page := 1
	perPage := 100 // maximum 100 entries per page
	stop := false
	for !stop {
		result, err := cp.client.Validators(cp.ctx, &height, &page, &perPage)
		if err != nil {
			return nil, err
		}
		vals.Validators = append(vals.Validators, result.Validators...)
		vals.Count += result.Count
		vals.Total = result.Total
		page++
		stop = vals.Count == vals.Total
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
func (cp *Node) Tx(hash string) (*types.Transaction, error) {
	resp, err := http.Get(fmt.Sprintf("%s/cosmos/tx/v1beta1/txs/%s", cp.txServiceAPI, hash))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("request failed with status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body")
	}

	var convTx *types.Transaction
	err = json.Unmarshal(body, &convTx)
	if err != nil {
		return nil, fmt.Errorf("error converting transaction: %s", err.Error())
	}

	return convTx, nil
}

// Txs implements node.Node
func (cp *Node) Txs(block *tmctypes.ResultBlock) ([]*types.Transaction, error) {
	txResponses := make([]*types.Transaction, len(block.Block.Txs))
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
