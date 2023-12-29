package local

import (
	"fmt"
	"os"
	"path"
	"reflect"
	"unsafe"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/simapp/params"

	"github.com/cosmos/cosmos-sdk/store"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/spf13/viper"
	cfg "github.com/tendermint/tendermint/config"
	"github.com/tendermint/tendermint/libs/log"
	tmnode "github.com/tendermint/tendermint/node"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	tmstore "github.com/tendermint/tendermint/store"
	db "github.com/tendermint/tm-db"

	"github.com/forbole/juno/v4/node"
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

	return &Source{
		StoreDB: levelDB,

		Codec:       encodingConfig.Codec,
		LegacyAmino: encodingConfig.Amino,

		BlockStore: tmstore.NewBlockStore(blockStoreDB),
		Logger:     log.NewTMLogger(log.NewSyncWriter(os.Stdout)).With("module", "explorer"),
		Cms:        store.NewCommitMultiStore(levelDB),
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

func getFieldUsingReflection(app interface{}, fieldName string) interface{} {
	fv := reflect.ValueOf(app).Elem().FieldByName(fieldName)
	return reflect.NewAt(fv.Type(), unsafe.Pointer(fv.UnsafeAddr())).Elem().Interface()
}

// MountKVStores allows to register the KV stores using the same KVStoreKey instances
// that are used inside the given app. To do so, this method uses the reflection to access
// the field with the specified name inside the given app. Such field must be of type
// map[string]*sdk.KVStoreKey and is commonly named something similar to "keys"
func (k Source) MountKVStores(app interface{}, fieldName string) error {
	keys, ok := getFieldUsingReflection(app, fieldName).(map[string]*storetypes.KVStoreKey)
	if !ok {
		return fmt.Errorf("error while getting keys")
	}

	for _, key := range keys {
		k.Cms.MountStoreWithDB(key, storetypes.StoreTypeIAVL, nil)
	}

	return nil
}

// MountTransientStores allows to register the Transient stores using the same TransientStoreKey instances
// that are used inside the given app. To do so, this method uses the reflection to access
// the field with the specified name inside the given app. Such field must be of type
// map[string]*sdk.TransientStoreKey and is commonly named something similar to "tkeys"
func (k Source) MountTransientStores(app interface{}, fieldName string) error {
	tkeys, ok := getFieldUsingReflection(app, fieldName).(map[string]*storetypes.TransientStoreKey)
	if !ok {
		return fmt.Errorf("error while getting transient keys")
	}

	for _, key := range tkeys {
		k.Cms.MountStoreWithDB(key, storetypes.StoreTypeTransient, nil)
	}

	return nil
}

// MountMemoryStores allows to register the Memory stores using the same MemoryStoreKey instances
// that are used inside the given app. To do so, this method uses the reflection to access
// the field with the specified name inside the given app. Such field must be of type
// map[string]*sdk.MemoryStoreKey and is commonly named something similar to "memkeys"
func (k Source) MountMemoryStores(app interface{}, fieldName string) error {
	memKeys, ok := getFieldUsingReflection(app, fieldName).(map[string]*storetypes.MemoryStoreKey)
	if !ok {
		return fmt.Errorf("error while getting memory keys")
	}

	for _, key := range memKeys {
		k.Cms.MountStoreWithDB(key, storetypes.StoreTypeMemory, nil)
	}

	return nil
}

// InitStores initializes the stores by mounting the various keys that have been specified.
// This method MUST be called before using any method that relies on the local storage somehow.
func (k Source) InitStores() error {
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
