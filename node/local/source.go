package local

import (
	"fmt"
	"os"
	"path"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/simapp/params"

	"github.com/cosmos/cosmos-sdk/store"
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramskeeper "github.com/cosmos/cosmos-sdk/x/params/keeper"
	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"github.com/spf13/viper"
	cfg "github.com/tendermint/tendermint/config"
	"github.com/tendermint/tendermint/libs/log"
	tmnode "github.com/tendermint/tendermint/node"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	tmstore "github.com/tendermint/tendermint/store"
	db "github.com/tendermint/tm-db"

	"github.com/forbole/juno/v2/node"
)

var (
	_ node.Source = &Source{}
)

// Source represents the Source interface implementation that reads the data from a local node
type Source struct {
	Initialized bool

	StoreDB db.DB

	Codec       codec.Codec
	LegacyAmino *codec.LegacyAmino

	BlockStore *tmstore.BlockStore
	Logger     log.Logger
	Cms        sdk.CommitMultiStore

	Keys  map[string]*sdk.KVStoreKey
	TKeys map[string]*sdk.TransientStoreKey

	ParamsKeeper paramskeeper.Keeper
}

// NewSource returns a new Source instance
func NewSource(home string, encodingConfig *params.EncodingConfig) (*Source, error) {
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

	cdc := encodingConfig.Marshaler
	legacyAmino := encodingConfig.Amino

	keys := sdk.NewKVStoreKeys(paramstypes.StoreKey)
	tKeys := sdk.NewTransientStoreKeys(paramstypes.TStoreKey)

	return &Source{
		StoreDB: levelDB,

		Codec:       cdc,
		LegacyAmino: legacyAmino,

		BlockStore: tmstore.NewBlockStore(blockStoreDB),
		Logger:     log.NewTMLogger(log.NewSyncWriter(os.Stdout)).With("module", "explorer"),
		Cms:        store.NewCommitMultiStore(levelDB),
		Keys:       keys,
		TKeys:      tKeys,

		ParamsKeeper: paramskeeper.NewKeeper(cdc, legacyAmino, keys[paramstypes.StoreKey], tKeys[paramstypes.TStoreKey]),
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

// Type implements keeper.Source
func (k Source) Type() string {
	return node.LocalKeeper
}

func (k Source) RegisterKey(key string) *sdk.KVStoreKey {
	k.Keys[key] = sdk.NewKVStoreKey(key)
	return k.Keys[key]
}

func (k Source) RegisterTKey(key string) *sdk.TransientStoreKey {
	k.TKeys[key] = sdk.NewTransientStoreKey(key)
	return k.TKeys[key]
}

func (k Source) RegisterSubspace(moduleName string) paramstypes.Subspace {
	subspace, ok := k.ParamsKeeper.GetSubspace(moduleName)
	if !ok {
		subspace = k.ParamsKeeper.Subspace(moduleName)
	}

	return subspace
}

// InitStores initializes the stores by mounting the various keys that have been specified
func (k Source) InitStores() error {
	for _, key := range k.Keys {
		k.Cms.MountStoreWithDB(key, sdk.StoreTypeIAVL, nil)
	}

	for _, tKey := range k.TKeys {
		k.Cms.MountStoreWithDB(tKey, sdk.StoreTypeTransient, nil)
	}

	// Load the latest version to properly init all the stores
	return k.Cms.LoadLatestVersion()
}

// LoadHeight loads the given height from the store.
// It returns a new Context that can be used to query the data, or an error if something wrong happens.
func (k Source) LoadHeight(height int64) (sdk.Context, error) {
	var err error
	var cms sdk.CacheMultiStore
	if height > 0 {
		cms, err = k.Cms.CacheMultiStoreWithVersion(height)
		if err != nil {
			return sdk.Context{}, err
		}
	} else {
		cms, err = k.Cms.CacheMultiStoreWithVersion(k.BlockStore.Height())
		if err != nil {
			return sdk.Context{}, err
		}
	}

	return sdk.NewContext(cms, tmproto.Header{}, false, k.Logger), nil
}
