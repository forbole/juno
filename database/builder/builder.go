package builder

import (
	"github.com/forbole/juno/v2/database"

	"github.com/forbole/juno/v2/database/postgresql"
)

// Builder represents a generic Builder implementation that build the proper database
// instance based on the configuration the user has specified
func Builder(ctx *database.Context) (database.Database, error) {
	return postgresql.Builder(ctx)
}
