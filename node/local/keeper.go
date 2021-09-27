package local

import (
	"fmt"
	"os"
	"path"

	"github.com/cosmos/cosmos-sdk/store"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/spf13/viper"
	cfg "github.com/tendermint/tendermint/config"
	"github.com/tendermint/tendermint/libs/log"
	tmnode "github.com/tendermint/tendermint/node"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	tmstore "github.com/tendermint/tendermint/store"
	tmtypes "github.com/tendermint/tendermint/types"
	db "github.com/tendermint/tm-db"

	"github.com/desmos-labs/juno/node"
)

var (
	_ node.Keeper = &Keeper{}
)

// Keeper represents the Keeper interface implementation that reads the data from a local node
type Keeper struct {
	Initialized bool

	StoreDB db.DB

	BlockStore *tmstore.BlockStore
	Logger     log.Logger
	Cms        sdk.CommitMultiStore

	Keys  map[string]*sdk.KVStoreKey
	TKeys map[string]*sdk.TransientStoreKey
}

// NewKeeper returns a new Keeper instance
func NewKeeper(home string) (*Keeper, error) {
	levelDB, err := sdk.NewLevelDB("application", path.Join(home, "data"))
	if err != nil {
		return nil, err
	}

	tmCfg, err := parseConfig(home)
	if err != nil {
		return nil, err
	}

	blockStoreDB, err := tmnode.DefaultDBProvider(&tmnode.DBContext{ID: "blockstore", Config: tmCfg})
	if err != nil {
		return nil, err
	}

	return &Keeper{
		StoreDB: levelDB,

		BlockStore: tmstore.NewBlockStore(blockStoreDB),
		Logger:     log.NewTMLogger(log.NewSyncWriter(os.Stdout)).With("module", "explorer"),
		Cms:        store.NewCommitMultiStore(levelDB),
		Keys:       make(map[string]*sdk.KVStoreKey),
		TKeys:      make(map[string]*sdk.TransientStoreKey),
	}, nil
}

func parseConfig(home string) (*cfg.Config, error) {
	viper.SetConfigFile(path.Join(home, "config", "config.toml"))

	conf := cfg.DefaultConfig()
	err := viper.Unmarshal(conf)
	if err != nil {
		return nil, err
	}
	conf.SetRoot(conf.RootDir)

	err = conf.ValidateBasic()
	if err != nil {
		return nil, fmt.Errorf("error in config file: %v", err)
	}

	return conf, nil
}

// Type implements keeper.Keeper
func (k Keeper) Type() string {
	return node.LocalKeeper
}

// InitStores initializes the stores by mounting the various keys that have been specified
func (k Keeper) InitStores() error {
	for _, key := range k.Keys {
		k.Cms.MountStoreWithDB(key, sdk.StoreTypeIAVL, nil)
	}

	for _, tKey := range k.Keys {
		k.Cms.MountStoreWithDB(tKey, sdk.StoreTypeTransient, nil)
	}

	// Load the latest version to properly init all the stores
	return k.Cms.LoadLatestVersion()
}

// LoadHeight loads the given height from the store.
// It returns a new Context that can be used to query the data, or an error if something wrong happens.
func (k Keeper) LoadHeight(height int64) (sdk.Context, error) {
	// Init the stores if not done already
	if !k.Initialized {
		err := k.InitStores()
		if err != nil {
			return sdk.Context{}, fmt.Errorf("error while initializing stores: %s", err)
		}
	}

	var err error
	var cms sdk.CacheMultiStore
	var block *tmtypes.Block
	if height > 0 {
		cms, err = k.Cms.CacheMultiStoreWithVersion(height)
		if err != nil {
			return sdk.Context{}, err
		}

		block = k.BlockStore.LoadBlock(height)
	} else {
		commit := k.Cms.LastCommitID()
		cms, err = k.Cms.CacheMultiStoreWithVersion(commit.Version)
		if err != nil {
			return sdk.Context{}, nil
		}

		block = k.BlockStore.LoadBlock(k.BlockStore.Height())
	}

	return sdk.NewContext(cms, tmproto.Header{ChainID: block.ChainID, Height: block.Height}, false, k.Logger), nil
}
