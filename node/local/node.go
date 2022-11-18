package local

import (
	"context"
	"encoding/hex"
	"fmt"
	"os"
	"sort"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/types/tx"
	"github.com/spf13/viper"
	cfg "github.com/tendermint/tendermint/config"
	cs "github.com/tendermint/tendermint/consensus"
	"github.com/tendermint/tendermint/evidence"
	"github.com/tendermint/tendermint/libs/log"
	tmmath "github.com/tendermint/tendermint/libs/math"
	tmquery "github.com/tendermint/tendermint/libs/pubsub/query"
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

	"github.com/saifullah619/juno/v3/node"
	"github.com/saifullah619/juno/v3/types"

	"path"
	"time"

	constypes "github.com/tendermint/tendermint/consensus/types"
	tmjson "github.com/tendermint/tendermint/libs/json"
	tmnode "github.com/tendermint/tendermint/node"
	tmctypes "github.com/tendermint/tendermint/rpc/core/types"
	dbm "github.com/tendermint/tm-db"
)

const (
	// see README
	defaultPerPage = 30
	maxPerPage     = 100
)

var (
	_ node.Node = &Node{}
)

// Node represents the node implementation that uses a local node
type Node struct {
	ctx      context.Context
	codec    codec.Codec
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
func NewNode(config *Details, txConfig client.TxConfig, codec codec.Codec) (*Node, error) {
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
		codec:    codec,
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
	blockStoreDB, err = dbProvider(&tmnode.DBContext{ID: "blockstore", Config: config})
	if err != nil {
		return
	}
	blockStore = store.NewBlockStore(blockStoreDB)

	stateDB, err = dbProvider(&tmnode.DBContext{ID: "state", Config: config})
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
		store, err := dbProvider(&tmnode.DBContext{ID: "tx_index", Config: config})
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

// latestHeight can be either latest committed or uncommitted (+1) height.
func (cp *Node) getHeight(latestHeight int64, heightPtr *int64) (int64, error) {
	if heightPtr != nil {
		height := *heightPtr
		if height <= 0 {
			return 0, fmt.Errorf("height must be greater than 0, but got %d", height)
		}
		if height > latestHeight {
			return 0, fmt.Errorf("height %d must be less than or equal to the current blockchain height %d",
				height, latestHeight)
		}
		base := cp.blockStore.Base()
		if height < base {
			return 0, fmt.Errorf("height %d is not available, lowest height is %d",
				height, base)
		}
		return height, nil
	}
	return latestHeight, nil
}

func validatePerPage(perPagePtr *int) int {
	if perPagePtr == nil { // no per_page parameter
		return defaultPerPage
	}

	perPage := *perPagePtr
	if perPage < 1 {
		return defaultPerPage
	} else if perPage > maxPerPage {
		return maxPerPage
	}
	return perPage
}

func validatePage(pagePtr *int, perPage, totalCount int) (int, error) {
	if perPage < 1 {
		panic(fmt.Sprintf("zero or negative perPage: %d", perPage))
	}

	if pagePtr == nil { // no page parameter
		return 1, nil
	}

	pages := ((totalCount - 1) / perPage) + 1
	if pages == 0 {
		pages = 1 // one page (even if it's empty)
	}
	page := *pagePtr
	if page <= 0 || page > pages {
		return 1, fmt.Errorf("page should be within [1, %d] range, given %d", pages, page)
	}

	return page, nil
}

func validateSkipCount(page, perPage int) int {
	skipCount := (page - 1) * perPage
	if skipCount < 0 {
		return 0
	}

	return skipCount
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

// ChainID implements node.Node
func (cp *Node) ChainID() (string, error) {
	return cp.genesisDoc.ChainID, nil
}

// Validators implements node.Node
func (cp *Node) Validators(height int64) (*tmctypes.ResultValidators, error) {
	height, err := cp.getHeight(cp.blockStore.Height(), &height)
	if err != nil {
		return nil, err
	}

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
	height, err := cp.getHeight(cp.blockStore.Height(), &height)
	if err != nil {
		return nil, err
	}

	block := cp.blockStore.LoadBlock(height)
	blockMeta := cp.blockStore.LoadBlockMeta(height)
	if blockMeta == nil {
		return &tmctypes.ResultBlock{BlockID: tmtypes.BlockID{}, Block: block}, nil
	}
	return &tmctypes.ResultBlock{BlockID: blockMeta.BlockID, Block: block}, nil
}

// BlockResults implements node.Node
func (cp *Node) BlockResults(height int64) (*tmctypes.ResultBlockResults, error) {
	height, err := cp.getHeight(cp.blockStore.Height(), &height)
	if err != nil {
		return nil, err
	}

	results, err := cp.stateStore.LoadABCIResponses(height)
	if err != nil {
		return nil, err
	}

	return &tmctypes.ResultBlockResults{
		Height:                height,
		TxsResults:            results.DeliverTxs,
		BeginBlockEvents:      results.BeginBlock.Events,
		EndBlockEvents:        results.EndBlock.Events,
		ValidatorUpdates:      results.EndBlock.ValidatorUpdates,
		ConsensusParamUpdates: results.EndBlock.ConsensusParamUpdates,
	}, nil
}

// Tx implements node.Node
func (cp *Node) Tx(hash string) (*types.Tx, error) {
	// if index is disabled, return error
	if _, ok := cp.txIndexer.(*null.TxIndex); ok {
		return nil, fmt.Errorf("transaction indexing is disabled")
	}

	hashBz, err := hex.DecodeString(hash)
	if err != nil {
		return nil, err
	}

	r, err := cp.txIndexer.Get(hashBz)
	if err != nil {
		return nil, err
	}

	if r == nil {
		return nil, fmt.Errorf("tx %s not found", hash)
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
		return nil, err
	}

	txResponse, err := makeTxResult(cp.txConfig, resTx, resBlock)
	if err != nil {
		return nil, err
	}

	protoTx, ok := txResponse.Tx.GetCachedValue().(*tx.Tx)
	if !ok {
		return nil, fmt.Errorf("expected %T, got %T", tx.Tx{}, txResponse.Tx.GetCachedValue())
	}

	// Decode messages
	for _, msg := range protoTx.Body.Messages {
		var stdMsg sdk.Msg
		err = cp.codec.UnpackAny(msg, &stdMsg)
		if err != nil {
			return nil, fmt.Errorf("error while unpacking message: %s", err)
		}
	}

	convTx, err := types.NewTx(txResponse, protoTx)
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
func (cp *Node) TxSearch(query string, pagePtr *int, perPagePtr *int, orderBy string) (*tmctypes.ResultTxSearch, error) {
	q, err := tmquery.New(query)
	if err != nil {
		return nil, err
	}

	results, err := cp.txIndexer.Search(cp.ctx, q)
	if err != nil {
		return nil, err
	}

	// sort results (must be done before pagination)
	switch orderBy {
	case "desc":
		sort.Slice(results, func(i, j int) bool {
			if results[i].Height == results[j].Height {
				return results[i].Index > results[j].Index
			}
			return results[i].Height > results[j].Height
		})
	case "asc", "":
		sort.Slice(results, func(i, j int) bool {
			if results[i].Height == results[j].Height {
				return results[i].Index < results[j].Index
			}
			return results[i].Height < results[j].Height
		})
	default:
		return nil, fmt.Errorf("expected order_by to be either `asc` or `desc` or empty")
	}

	// paginate results
	totalCount := len(results)
	perPage := validatePerPage(perPagePtr)

	page, err := validatePage(pagePtr, perPage, totalCount)
	if err != nil {
		return nil, err
	}

	skipCount := validateSkipCount(page, perPage)
	pageSize := tmmath.MinInt(perPage, totalCount-skipCount)

	apiResults := make([]*tmctypes.ResultTx, 0, pageSize)
	for i := skipCount; i < skipCount+pageSize; i++ {
		r := results[i]

		var proof tmtypes.TxProof
		apiResults = append(apiResults, &tmctypes.ResultTx{
			Hash:     tmtypes.Tx(r.Tx).Hash(),
			Height:   r.Height,
			Index:    r.Index,
			TxResult: r.Result,
			Tx:       r.Tx,
			Proof:    proof,
		})
	}

	return &tmctypes.ResultTxSearch{Txs: apiResults, TotalCount: totalCount}, nil
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
