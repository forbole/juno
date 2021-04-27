package builder

import (
	"github.com/cosmos/cosmos-sdk/simapp/params"
	"github.com/desmos-labs/juno/types"

	"github.com/desmos-labs/juno/db"

	"github.com/desmos-labs/juno/db/postgresql"
)

// Builder represents a generic Builder implementation that build the proper database
// instance based on the configuration the user has specified
func Builder(cfg *types.Config, encodingConfig *params.EncodingConfig) (db.Database, error) {
	return postgresql.Builder(cfg.Database, encodingConfig)
}
