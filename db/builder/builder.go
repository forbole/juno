package builder

import (
	"github.com/desmos-labs/juno/db"

	"github.com/desmos-labs/juno/db/postgresql"
)

// Builder represents a generic Builder implementation that build the proper database
// instance based on the configuration the user has specified
func Builder(ctx *db.Context) (db.Database, error) {
	return postgresql.Builder(ctx)
}
