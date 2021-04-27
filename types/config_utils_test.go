package types_test

import (
	"github.com/desmos-labs/juno/types"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseConfigString_PostgreSQL(t *testing.T) {
	tomlString := `
rpc_node = "https://rpc.morpheus.desmos.network:26657"
client_node = "https://lcd.morpheus.desmos.network:1317"

[database]
name = "desmos"
host = "localhost"
port = 5432
user = "user"
password = "password"
`

	cfg, err := types.DefaultConfigParser([]byte(tomlString))
	require.NoError(t, err)

	require.Equal(t, "desmos", cfg.Database.Name)
	require.Equal(t, "localhost", cfg.Database.Host)
	require.Equal(t, uint64(5432), cfg.Database.Port)
	require.Equal(t, "user", cfg.Database.User)
	require.Equal(t, "password", cfg.Database.Password)
}
