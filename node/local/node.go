package local

import (
	"context"
	"encoding/hex"
	"fmt"
	"os"

	cfg "github.com/tendermint/tendermint/config"
	cs "github.com/tendermint/tendermint/consensus"
	"github.com/tendermint/tendermint/evidence"
	"github.com/tendermint/tendermint/privval"
	"github.com/tendermint/tendermint/proxy"
	sm "github.com/tendermint/tendermint/state"
	"github.com/tendermint/tendermint/state/indexer"
	blockidxkv "github.com/tendermint/tendermint/state/indexer/block/kv"
	blockidxnull "github.com/tendermint/tendermint/state/indexer/block/null"
	"github.com/tendermint/tendermint/state/txindex"
	"github.com/tendermint/tendermint/state/txindex/kv"
	"github.com/tendermint/tendermint/state/txindex/null"
	"github.com/tendermint/tendermint/store"
	tmtypes "github.com/tendermint/tendermint/types"

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
	tmctypes "github.com/tendermint/tendermint/rpc/core/types"
	dbm "github.com/tendermint/tm-db"
)

var (
	_ node.Node = &Node{}
)

// Node represents the node implementation that uses a local node
type Node struct {
	ctx      context.Context
	txConfig client.TxConfig

	// config
	tmCfg      *cfg.Config
	genesisDoc *tmtypes.GenesisDoc

	// services
	eventBus       *tmtypes.EventBus
	stateStore     sm.Store
	blockStore     *store.BlockStore
	consensusState *cs.State
	txIndexer      txindex.TxIndexer
	blockIndexer   indexer.BlockIndexer
}

// NewNode returns a new Node instance
func NewNode(config *Details, txConfig client.TxConfig) (*Node, error) {
	// Load the config
	viper.SetConfigFile(path.Join(config.Home, "config", "config.yaml"))
	tmCfg, err := ParseConfig()
	if err != nil {
		return nil, err
	}
	tmCfg.SetRoot(config.Home)

	// Build the local node
	dbProvider := tmnode.DefaultDBProvider
	genesisDocProvider := tmnode.DefaultGenesisDocProviderFunc(tmCfg)
	logger := log.NewTMLogger(log.NewSyncWriter(os.Stdout)).With("module", "explorer")
	clientCreator := proxy.DefaultClientCreator(tmCfg.ProxyApp, tmCfg.ABCI, tmCfg.DBDir())
	metricsProvider := tmnode.DefaultMetricsProvider(tmCfg.Instrumentation)

	privval.LoadOrGenFilePV(tmCfg.PrivValidatorKeyFile(), tmCfg.PrivValidatorStateFile())
	proxy.DefaultClientCreator(tmCfg.ProxyApp, tmCfg.ABCI, tmCfg.DBDir())

	blockStore, stateDB, err := initDBs(tmCfg, dbProvider)
	if err != nil {
		return nil, err
	}

	stateStore := sm.NewStore(stateDB)

	_, genDoc, err := tmnode.LoadStateFromDBOrGenesisDocProvider(stateDB, genesisDocProvider)
	if err != nil {
		return nil, err
	}

	eventBus, err := createAndStartEventBus(logger)
	if err != nil {
		return nil, err
	}

	_, txIndexer, blockIndexer, err := createAndStartIndexerService(tmCfg, dbProvider, eventBus, logger)
	if err != nil {
		return nil, err
	}

	state, err := stateStore.Load()
	if err != nil {
		return nil, err
	}

	proxyApp := proxy.NewAppConns(clientCreator)

	csMetrics, _, _, smMetrics := metricsProvider(genDoc.ChainID)

	evidenceDB, err := dbProvider(&tmnode.DBContext{ID: "evidence", Config: tmCfg})
	if err != nil {
		return nil, err
	}

	evidencePool, err := evidence.NewPool(evidenceDB, sm.NewStore(stateDB), blockStore)
	if err != nil {
		return nil, err
	}

	blockExec := sm.NewBlockExecutor(
		stateStore,
		logger.With("module", "state"),
		proxyApp.Consensus(),
		nil,
		evidencePool,
		sm.BlockExecutorWithMetrics(smMetrics),
	)

	consensusState := cs.NewState(
		tmCfg.Consensus,
		state.Copy(),
		blockExec,
		blockStore,
		nil,
		evidencePool,
		cs.StateMetrics(csMetrics),
	)

	return &Node{
		ctx:      context.Background(),
		txConfig: txConfig,

		tmCfg:      tmCfg,
		genesisDoc: genDoc,

		eventBus:       eventBus,
		stateStore:     stateStore,
		consensusState: consensusState,
		blockStore:     blockStore,
		txIndexer:      txIndexer,
		blockIndexer:   blockIndexer,
	}, nil
}

func initDBs(config *cfg.Config, dbProvider tmnode.DBProvider) (blockStore *store.BlockStore, stateDB dbm.DB, err error) {
	var blockStoreDB dbm.DB
	blockStoreDB, err = dbProvider(&tmnode.DBContext{"blockstore", config})
	if err != nil {
		return
	}
	blockStore = store.NewBlockStore(blockStoreDB)

	stateDB, err = dbProvider(&tmnode.DBContext{"state", config})
	if err != nil {
		return
	}

	return
}

func createAndStartEventBus(logger log.Logger) (*tmtypes.EventBus, error) {
	eventBus := tmtypes.NewEventBus()
	eventBus.SetLogger(logger.With("module", "events"))
	if err := eventBus.Start(); err != nil {
		return nil, err
	}
	return eventBus, nil
}

func createAndStartIndexerService(
	config *cfg.Config,
	dbProvider tmnode.DBProvider,
	eventBus *tmtypes.EventBus,
	logger log.Logger,
) (*txindex.IndexerService, txindex.TxIndexer, indexer.BlockIndexer, error) {

	var (
		txIndexer    txindex.TxIndexer
		blockIndexer indexer.BlockIndexer
	)

	switch config.TxIndex.Indexer {
	case "kv":
		store, err := dbProvider(&tmnode.DBContext{"tx_index", config})
		if err != nil {
			return nil, nil, nil, err
		}

		txIndexer = kv.NewTxIndex(store)
		blockIndexer = blockidxkv.New(dbm.NewPrefixDB(store, []byte("block_events")))
	default:
		txIndexer = &null.TxIndex{}
		blockIndexer = &blockidxnull.BlockerIndexer{}
	}

	indexerService := txindex.NewIndexerService(txIndexer, blockIndexer, eventBus)
	indexerService.SetLogger(logger.With("module", "txindex"))

	if err := indexerService.Start(); err != nil {
		return nil, nil, nil, err
	}

	return indexerService, txIndexer, blockIndexer, nil
}

// Genesis implements node.Node
func (cp *Node) Genesis() (*tmctypes.ResultGenesis, error) {
	return &tmctypes.ResultGenesis{Genesis: cp.genesisDoc}, nil
}

// ConsensusState implements node.Node
func (cp *Node) ConsensusState() (*constypes.RoundStateSimple, error) {
	bz, err := cp.consensusState.GetRoundStateSimpleJSON()
	if err != nil {
		return nil, err
	}

	var data constypes.RoundStateSimple
	err = tmjson.Unmarshal(bz, &data)
	if err != nil {
		return nil, err
	}
	return &data, nil
}

// LatestHeight implements node.Node
func (cp *Node) LatestHeight() (int64, error) {
	return cp.blockStore.Height(), nil
}

// Validators implements node.Node
func (cp *Node) Validators(height int64) (*tmctypes.ResultValidators, error) {
	valSet, err := cp.stateStore.LoadValidators(height)
	if err != nil {
		return nil, err
	}

	return &tmctypes.ResultValidators{
		BlockHeight: height,
		Validators:  valSet.Validators,
		Count:       len(valSet.Validators),
		Total:       len(valSet.Validators),
	}, nil
}

// Block implements node.Node
func (cp *Node) Block(height int64) (*tmctypes.ResultBlock, error) {
	block := cp.blockStore.LoadBlock(height)
	blockMeta := cp.blockStore.LoadBlockMeta(height)
	if blockMeta == nil {
		return &tmctypes.ResultBlock{BlockID: tmtypes.BlockID{}, Block: block}, nil
	}
	return &tmctypes.ResultBlock{BlockID: blockMeta.BlockID, Block: block}, nil
}

// Tx implements node.Node
func (cp *Node) Tx(hash string) (*sdk.TxResponse, *tx.Tx, error) {
	// if index is disabled, return error
	if _, ok := cp.txIndexer.(*null.TxIndex); ok {
		return nil, nil, fmt.Errorf("transaction indexing is disabled")
	}

	hashBz, err := hex.DecodeString(hash)
	if err != nil {
		return nil, nil, err
	}

	r, err := cp.txIndexer.Get(hashBz)
	if err != nil {
		return nil, nil, err
	}

	if r == nil {
		return nil, nil, fmt.Errorf("tx %s not found", hash)
	}

	height := r.Height
	index := r.Index

	resTx := &tmctypes.ResultTx{
		Hash:     []byte(hash),
		Height:   height,
		Index:    index,
		TxResult: r.Result,
		Tx:       r.Tx,
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
	_, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	eventCh := make(<-chan tmctypes.ResultEvent)
	return eventCh, cancel, nil
}

// SubscribeNewBlocks implements node.Node
func (cp *Node) SubscribeNewBlocks(subscriber string) (<-chan tmctypes.ResultEvent, context.CancelFunc, error) {
	return cp.SubscribeEvents(subscriber, "tm.event = 'NewBlock'")
}

// Stop implements node.Node
func (cp *Node) Stop() {
}
