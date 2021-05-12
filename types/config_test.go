package types

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDefaultConfigParser(t *testing.T) {
	data := `
[cosmos]
  prefix = "cosmos"
  modules = [
    "pruning"
  ]

[rpc]
  client_name = "juno"
  address = "http://localhost:26657"

[grpc]
  address = "localhost:9090"
  insecure = true

[logging]
  format = "text"
  level = "debug"

[parsing]
  workers = 5
  listen_new_blocks = true
  parse_old_blocks = true
  parse_genesis = true
  start_height = 1
  fast_sync = false

[database]
  host = "localhost"
  name = "juno"
  password = "password"
  port = 5432
  schema = "public"
  ssl_mode = ""
  user = "user"

[pruning]
  keep_recent = 100
  keep_every = 5
  interval = 10
`

	cfg, err := DefaultConfigParser([]byte(data))
	require.NoError(t, err)

	require.Equal(t, "cosmos", cfg.GetCosmosConfig().GetPrefix())
	require.Equal(t, []string{"pruning"}, cfg.GetCosmosConfig().GetModules())

	require.Equal(t, "juno", cfg.GetRPCConfig().GetClientName())
	require.Equal(t, "http://localhost:26657", cfg.GetRPCConfig().GetAddress())

	require.Equal(t, "localhost:9090", cfg.GetGrpcConfig().GetAddress())
	require.Equal(t, true, cfg.GetGrpcConfig().IsInsecure())
}
