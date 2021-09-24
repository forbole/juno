package local

import (
	"context"
	"fmt"
	"os"

	"github.com/cosmos/cosmos-sdk/client"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx"
	"github.com/spf13/viper"
	"github.com/tendermint/tendermint/libs/log"

	"github.com/desmos-labs/juno/node"
	"github.com/desmos-labs/juno/types"

	"path"
	"time"

	constypes "github.com/tendermint/tendermint/consensus/types"
	tmjson "github.com/tendermint/tendermint/libs/json"
	tmnode "github.com/tendermint/tendermint/node"
	localclient "github.com/tendermint/tendermint/rpc/client/local"
	tmctypes "github.com/tendermint/tendermint/rpc/core/types"
)

var (
	_ node.Node = &Node{}
)

// Node represents the node implementation that uses a local node
type Node struct {
	ctx      context.Context
	txConfig client.TxConfig
	client   *localclient.Local
}

// NewNode returns a new Node instance
func NewNode(config *Details, txConfig client.TxConfig) (*Node, error) {
	// Load the config
	viper.SetConfigFile(path.Join(config.Home, "config", "config.toml"))
	tmCfg, err := ParseConfig()
	if err != nil {
		return nil, err
	}
	tmCfg.SetRoot(config.Home)

	// Build the local node
	logger := log.NewTMLogger(log.NewSyncWriter(os.Stdout)).With("module", "explorer")
	tmNode, err := tmnode.DefaultNewNode(tmCfg, logger)
	if err != nil {
		return nil, err
	}

	return &Node{
		ctx:      context.Background(),
		txConfig: txConfig,
		client:   localclient.New(tmNode),
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
	stop := false
	for !stop {
		result, err := cp.client.Validators(cp.ctx, &height, &page, nil)
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

// Block implements node.Node
func (cp *Node) Block(height int64) (*tmctypes.ResultBlock, error) {
	return cp.client.Block(cp.ctx, &height)
}

// Tx implements node.Node
func (cp *Node) Tx(hash string) (*sdk.TxResponse, *tx.Tx, error) {
	resTx, err := cp.client.Tx(cp.ctx, []byte(hash), false)
	if err != nil {
		return nil, nil, err
	}

	resBlock, err := cp.Block(resTx.Height)
	if err != nil {
		return nil, nil, err
	}

	txResponse, err := makeTxResult(cp.txConfig, resTx, resBlock)
	if err != nil {
		return nil, nil, err
	}

	protoTx, ok := txResponse.Tx.GetCachedValue().(*tx.Tx)
	if !ok {
		return nil, nil, fmt.Errorf("expected %T, got %T", tx.Tx{}, txResponse.Tx.GetCachedValue())
	}

	return txResponse, protoTx, nil
}

// Txs implements node.Node
func (cp *Node) Txs(block *tmctypes.ResultBlock) ([]*types.Tx, error) {
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
