package types

import (
	"fmt"
	"reflect"

	"github.com/forbole/juno/v5/database/postgresql"
	"github.com/forbole/juno/v5/interfaces"
	"github.com/forbole/juno/v5/node"

	sdk "github.com/cosmos/cosmos-sdk/types"
	nodebuilder "github.com/forbole/juno/v5/node/builder"
	"github.com/forbole/juno/v5/types/config"
	"github.com/forbole/juno/v5/types/params"

	modsregistrar "github.com/forbole/juno/v5/modules/registrar"
)

type Infrastructures struct {
	Database       *postgresql.Database
	EncodingConfig params.EncodingConfig
	Node           node.Node
	Logger         interfaces.Logger
	Modules        interfaces.Modules
}

// GetInfrastructures setups all the things that can be used to later parse the chain state
func GetInfrastructures(cfg config.Config, parseConfig *Config) (*Infrastructures, error) {
	// Build the codec
	encodingConfig := parseConfig.GetEncodingConfigBuilder()()

	// Setup the SDK configuration
	sdkConfig, sealed := getConfig()
	if !sealed {
		parseConfig.GetSetupConfig()(cfg, sdkConfig)
		sdkConfig.Seal()
	}

	// Get the db
	databaseCtx := postgresql.NewContext(cfg.Database, encodingConfig, parseConfig.GetLogger())
	db, err := postgresql.Builder(databaseCtx)
	if err != nil {
		return nil, err
	}

	// Init the client
	cp, err := nodebuilder.BuildNode(cfg.Node, encodingConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to start client: %s", err)
	}

	// Setup the logging
	err = parseConfig.GetLogger().SetLogFormat(cfg.Logging.LogFormat)
	if err != nil {
		return nil, fmt.Errorf("error while setting logging format: %s", err)
	}

	err = parseConfig.GetLogger().SetLogLevel(cfg.Logging.LogLevel)
	if err != nil {
		return nil, fmt.Errorf("error while setting logging level: %s", err)
	}

	// Get the modules
	modContext := modsregistrar.NewContext(cfg, sdkConfig, encodingConfig, db, cp, parseConfig.GetLogger())
	mods := parseConfig.GetRegistrar().BuildModules(modContext)
	registeredModules := modsregistrar.GetModules(mods, cfg.Chain.Modules, parseConfig.GetLogger())

	return &Infrastructures{
		EncodingConfig: encodingConfig,
		Node:           cp,
		Database:       db,
		Logger:         parseConfig.GetLogger(),
		Modules:        registeredModules,
	}, nil
}

// getConfig returns the SDK Config instance as well as if it's sealed or not
func getConfig() (config *sdk.Config, sealed bool) {
	sdkConfig := sdk.GetConfig()
	fv := reflect.ValueOf(sdkConfig).Elem().FieldByName("sealed")
	return sdkConfig, fv.Bool()
}
