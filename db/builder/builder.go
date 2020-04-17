package builder

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/desmos-labs/juno/config"
	"github.com/desmos-labs/juno/db"
	"github.com/desmos-labs/juno/db/mongo"
	"github.com/desmos-labs/juno/db/postgresql"
)

// DatabaseBuilder represents a generic Builder implementation that build the proper database
// instance based on the configuration the user has specified
func DatabaseBuilder(cfg config.Config, codec *codec.Codec) (*db.Database, error) {
	switch cfg := cfg.DatabaseConfig.Config.(type) {
	case *config.MongoDBConfig:
		return mongo.Builder(*cfg, codec)
	case *config.PostgreSQLConfig:
		return postgresql.Builder(*cfg, codec)
	}

	return nil, fmt.Errorf("invalid config")
}
