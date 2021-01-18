package builder

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/simapp/params"

	"github.com/desmos-labs/juno/db"

	"github.com/desmos-labs/juno/config"
	"github.com/desmos-labs/juno/db/postgresql"
)

// Builder represents a generic Builder implementation that build the proper database
// instance based on the configuration the user has specified
func Builder(cfg *config.Config, encodingConfig *params.EncodingConfig) (db.Database, error) {
	switch cfg := cfg.DatabaseConfig.Config.(type) {
	case *config.PostgreSQLConfig:
		return postgresql.Builder(cfg, encodingConfig)
	}

	return nil, fmt.Errorf("invalid config")
}
