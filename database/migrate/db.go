package migrate

import (
	"fmt"

	"database/sql"
	"github.com/forbole/juno/v2/logging"

	"github.com/cosmos/cosmos-sdk/simapp/params"

	"github.com/forbole/juno/v2/database"
	"github.com/jmoiron/sqlx"

)

// MigrateDbBuilder creates a database connection with the given database connection info
// from config. It returns a database connection handle or an error if the
// connection fails.
func MigrateDbBuilder(ctx *database.Context) (database.MigrateDb, error) {
	sslMode := "disable"
	if ctx.Cfg.SSLMode != "" {
		sslMode = ctx.Cfg.SSLMode
	}

	schema := "public"
	if ctx.Cfg.Schema != "" {
		schema = ctx.Cfg.Schema
	}

	connStr := fmt.Sprintf(
		"host=%s port=%d dbname=%s user=%s sslmode=%s search_path=%s",
		ctx.Cfg.Host, ctx.Cfg.Port, ctx.Cfg.Name, ctx.Cfg.User, sslMode, schema,
	)

	if ctx.Cfg.Password != "" {
		connStr += fmt.Sprintf(" password=%s", ctx.Cfg.Password)
	}

	postgresDb, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}

	// Set max open connections
	postgresDb.SetMaxOpenConns(ctx.Cfg.MaxOpenConnections)
	postgresDb.SetMaxIdleConns(ctx.Cfg.MaxIdleConnections)

	return &MigrateDb{
		Sql:            postgresDb,
		Sqlx: 			sqlx.NewDb(postgresDb, "postgresql"),
		EncodingConfig: ctx.EncodingConfig,
		Logger:         ctx.Logger,
	}, nil
}

// type check to ensure interface is properly implemented
var _ database.MigrateDb = &MigrateDb{}

// MigrateDb defines a wrapper around a SQL database and implements functionality
// for data aggregation and exporting.
type MigrateDb struct {
	Sql            *sql.DB
	Sqlx   		   *sqlx.DB
	EncodingConfig *params.EncodingConfig
	Logger         logging.Logger
}