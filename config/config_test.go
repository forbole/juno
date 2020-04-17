package config_test

import (
	"testing"

	"github.com/desmos-labs/juno/config"
	"github.com/stretchr/testify/require"
)

func TestParseConfigString_PostgreSQL(t *testing.T) {
	tomlString := `
rpc_node = "http://rpc.morpheus.desmos.network:26657"
client_node = "http://lcd.morpheus.desmos.network:1317"

[database]
type = "postgresql"

[database.config]
name = "desmos"
host = "localhost"
port = 5432
user = "user"
password = "password"
`

	cfg, err := config.ParseConfigString([]byte(tomlString))
	require.NoError(t, err)

	postgreConfig, ok := cfg.DatabaseConfig.Config.(*config.PostgreSQLConfig)
	require.True(t, ok)
	require.Equal(t, "desmos", postgreConfig.Name)
	require.Equal(t, "localhost", postgreConfig.Host)
	require.Equal(t, uint64(5432), postgreConfig.Port)
	require.Equal(t, "user", postgreConfig.User)
	require.Equal(t, "password", postgreConfig.Password)
}

func TestParseConfigString_MongoDB(t *testing.T) {
	tomlString := `
rpc_node = "http://rpc.morpheus.desmos.network:26657"
client_node = "http://lcd.morpheus.desmos.network:1317"

[database]
type = "mongodb"

[database.config]
name = "desmos"
uri = "mongodb://example.com"
`

	cfg, err := config.ParseConfigString([]byte(tomlString))
	require.NoError(t, err)

	mongoConfig, ok := cfg.DatabaseConfig.Config.(*config.MongoDBConfig)
	require.True(t, ok)
	require.Equal(t, "desmos", mongoConfig.Name)
	require.Equal(t, "mongodb://example.com", mongoConfig.Uri)
}
