package builder

import (
	"fmt"

	"github.com/desmos-labs/juno/db"

	"github.com/cosmos/cosmos-sdk/codec"

	"github.com/desmos-labs/juno/config"
	"github.com/desmos-labs/juno/db/mongo"
	"github.com/desmos-labs/juno/db/postgresql"
)

// Builder represents a generic Builder implementation that build the proper database
// instance based on the configuration the user has specified
func Builder(cfg *config.Config, codec *codec.LegacyAmino) (db.Database, error) {
	switch cfg := cfg.DatabaseConfig.Config.(type) {
	case *config.MongoDBConfig:
		return mongo.Builder(cfg, codec)
	case *config.PostgreSQLConfig:
		return postgresql.Builder(cfg, codec)
	}

	return nil, fmt.Errorf("invalid config")
}
