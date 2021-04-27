package worker

import (
	"github.com/cosmos/cosmos-sdk/simapp/params"

	"github.com/desmos-labs/juno/client"
	"github.com/desmos-labs/juno/db"
	"github.com/desmos-labs/juno/modules"
	"github.com/desmos-labs/juno/types"
)

type Config struct {
	EncodingConfig *params.EncodingConfig
	Queue          types.HeightQueue
	ClientProxy    *client.Proxy
	Database       db.Database
	Modules        []modules.Module
}

func NewConfig(
	queue types.HeightQueue,
	encodingConfig *params.EncodingConfig,
	clientProxy *client.Proxy,
	db db.Database,
	modules []modules.Module,
) *Config {
	return &Config{
		EncodingConfig: encodingConfig,
		Queue:          queue,
		ClientProxy:    clientProxy,
		Database:       db,
		Modules:        modules,
	}
}
