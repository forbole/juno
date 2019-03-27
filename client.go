package main

import (
	"encoding/hex"

	rpcclient "github.com/tendermint/tendermint/rpc/client"
	tmctypes "github.com/tendermint/tendermint/rpc/core/types"
)

// rpcClient defines a Tendermint RPC client wrapper and implements essential
// data queriers.
type rpcClient struct {
	nodeURI string
	node    rpcclient.Client
}

func newRPCClient(nodeURI string) (rpcClient, error) {
	rpc := rpcclient.NewHTTP(nodeURI, "/websocket")

	if err := rpc.Start(); err != nil {
		return rpcClient{}, err
	}

	return rpcClient{nodeURI, rpc}, nil
}

// latestHeight returns the latest block height on the active chain. An error
// is returned if the query fails.
func (c rpcClient) latestHeight() (int64, error) {
	status, err := c.node.Status()
	if err != nil {
		return -1, err
	}

	height := status.SyncInfo.LatestBlockHeight
	return height, nil
}

// block queries for a block by height. An error is returned if the query fails.
func (c rpcClient) block(height int64) (*tmctypes.ResultBlock, error) {
	return c.node.Block(&height)
}

// tx queries for a transaction by hash. An error is returned if the query fails.
func (c rpcClient) tx(hash string) (*tmctypes.ResultTx, error) {
	hashRaw, err := hex.DecodeString(hash)
	if err != nil {
		return nil, err
	}

	return c.node.Tx(hashRaw, false)
}

// txsFromBlock returns all the transactions for a given block height. An error
// is returned if the query fails.
func (c rpcClient) txsFromBlock(block *tmctypes.ResultBlock) ([]*tmctypes.ResultTx, error) {
	dataTxs := block.Block.Data.Txs
	txs := make([]*tmctypes.ResultTx, len(dataTxs))

	for i, d := range dataTxs {
		tx, err := c.tx(hex.EncodeToString(d.Hash()))
		if err != nil {
			return nil, err
		}

		txs[i] = tx
	}

	return txs, nil
}

// validators returns all the known Tendermint validators for a given block
// height. An error is returned if the query fails.
func (c rpcClient) validators(height int64) (*tmctypes.ResultValidators, error) {
	return c.node.Validators(&height)
}
