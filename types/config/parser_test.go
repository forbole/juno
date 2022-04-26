package config

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDefaultConfigParser(t *testing.T) {
	data := `
chain:
  bech32_prefix: cosmos
  modules:
    - pruning

node: 
  type: remote
  rpc:
    client_name: juno
    address: http://localhost:26657

  grpc:
    address: localhost:9090
    insecure: true

logging:
  format: text
  level: debug

parser:
  workers: 5
  listen_new_blocks: true
  parse_old_blocks: true
  parse_genesis: true
  start_height: 1
  fast_sync: false

database:
  host: localhost
  name: juno
  password: password
  port: 5432
  schema: public
  ssl_mode: 
  user: user
`

	cfg, err := DefaultConfigParser([]byte(data))
	require.NoError(t, err)
	bytes, _ := cfg.GetBytes()
	require.NotEmpty(t, bytes)
	require.Equal(t, "cosmos", cfg.Chain.Bech32Prefix)
	require.Equal(t, []string{"pruning"}, cfg.Chain.Modules)
}
