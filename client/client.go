package client

import (
	"context"
	"encoding/hex"
	"time"

	rpcclient "github.com/tendermint/tendermint/rpc/client"
	tmctypes "github.com/tendermint/tendermint/rpc/core/types"
)

// RPCClient defines a Tendermint RPC client wrapper and implements essential
// data queriers.
type RPCClient struct {
	nodeURI string
	node    rpcclient.Client
}

func NewRPCClient(nodeURI string) (RPCClient, error) {
	rpc := rpcclient.NewHTTP(nodeURI, "/websocket")

	if err := rpc.Start(); err != nil {
		return RPCClient{}, err
	}

	return RPCClient{nodeURI, rpc}, nil
}

// LatestHeight returns the latest block height on the active chain. An error
// is returned if the query fails.
func (c RPCClient) LatestHeight() (int64, error) {
	status, err := c.node.Status()
	if err != nil {
		return -1, err
	}

	height := status.SyncInfo.LatestBlockHeight
	return height, nil
}

// Block queries for a block by height. An error is returned if the query fails.
func (c RPCClient) Block(height int64) (*tmctypes.ResultBlock, error) {
	return c.node.Block(&height)
}

// Tx queries for a transaction by hash. An error is returned if the query fails.
func (c RPCClient) Tx(hash string) (*tmctypes.ResultTx, error) {
	hashRaw, err := hex.DecodeString(hash)
	if err != nil {
		return nil, err
	}

	return c.node.Tx(hashRaw, false)
}

// TxsFromBlock returns all the transactions for a given block height. An error
// is returned if the query fails.
func (c RPCClient) TxsFromBlock(block *tmctypes.ResultBlock) ([]*tmctypes.ResultTx, error) {
	dataTxs := block.Block.Data.Txs
	txs := make([]*tmctypes.ResultTx, len(dataTxs))

	for i, d := range dataTxs {
		tx, err := c.Tx(hex.EncodeToString(d.Hash()))
		if err != nil {
			return nil, err
		}

		txs[i] = tx
	}

	return txs, nil
}

// Validators returns all the known Tendermint validators for a given block
// height. An error is returned if the query fails.
func (c RPCClient) Validators(height int64) (*tmctypes.ResultValidators, error) {
	return c.node.Validators(&height)
}

// Stop defers the node stop execution to the RPC client.
func (c RPCClient) Stop() error {
	return c.node.Stop()
}

// SubscribeNewBlocks subscribes to the new block event handler through the RPC
// client with the given subscriber name. An receiving only channel, context
// cancel function and an error is returned. It is up to the caller to cancel
// the context and handle any errors appropriately.
func (c RPCClient) SubscribeNewBlocks(subscriber string) (<-chan tmctypes.ResultEvent, context.CancelFunc, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	eventCh, err := c.node.Subscribe(ctx, subscriber, "tm.event = 'NewBlock'")
	return eventCh, cancel, err
}
