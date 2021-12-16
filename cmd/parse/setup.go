package parse

import (
	"fmt"
	"reflect"

	nodebuilder "github.com/forbole/juno/v2/node/builder"
	"github.com/forbole/juno/v2/types/config"

	"github.com/forbole/juno/v2/database"

	sdk "github.com/cosmos/cosmos-sdk/types"

	modsregistrar "github.com/forbole/juno/v2/modules/registrar"
)

// GetParsingContext setups all the things that should be later passed to StartParsing in order
// to parse the chain data properly.
func GetParsingContext(parseConfig *Config) (*Context, error) {
	// Get the global config
	cfg := config.Cfg

	// Build the codec
	encodingConfig := parseConfig.GetEncodingConfigBuilder()()

	// Setup the SDK configuration
	sdkConfig, sealed := getConfig()
	if !sealed {
		parseConfig.GetSetupConfig()(cfg, sdkConfig)
		sdkConfig.Seal()
	}

	// Get the db
	databaseCtx := database.NewContext(cfg.Database, &encodingConfig, parseConfig.GetLogger())
	db, err := parseConfig.GetDBBuilder()(databaseCtx)
	if err != nil {
		return nil, err
	}

	// Init the client
	cp, err := nodebuilder.BuildNode(cfg.Node, &encodingConfig)
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
	context := modsregistrar.NewContext(cfg, sdkConfig, &encodingConfig, db, cp, parseConfig.GetLogger())
	mods := parseConfig.GetRegistrar().BuildModules(context)
	registeredModules := modsregistrar.GetModules(mods, cfg.Chain.Modules, parseConfig.GetLogger())

	return NewContext(&encodingConfig, cp, db, parseConfig.GetLogger(), registeredModules), nil
}

// getConfig returns the SDK Config instance as well as if it's sealed or not
func getConfig() (config *sdk.Config, sealed bool) {
	sdkConfig := sdk.GetConfig()
	fv := reflect.ValueOf(sdkConfig).Elem().FieldByName("sealed")
	return sdkConfig, fv.Bool()
}
