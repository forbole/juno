package config_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"

	nodeconfig "github.com/saifullah619/juno/v3/node/config"
	"github.com/saifullah619/juno/v3/node/local"
	"github.com/saifullah619/juno/v3/node/remote"
)

func TestConfig_UnmarshalYAML(t *testing.T) {
	var remoteData = `
type: "remote"
config:
  rpc:
    client_name: "juno"
    max_connections: 1
    address: "http://localhost:26657"

  grpc:
    insecure: true
    address: "http://localhost:9090"
`

	var config nodeconfig.Config
	err := yaml.Unmarshal([]byte(remoteData), &config)
	require.NoError(t, err)
	require.IsType(t, &remote.Details{}, config.Details)

	var localData = `
type: "local"
config: 
  home: /home/user/.cosmos
`

	err = yaml.Unmarshal([]byte(localData), &config)
	require.NoError(t, err)
	require.IsType(t, &local.Details{}, config.Details)
}

func TestConfig_MarshalYAML(t *testing.T) {
	config := nodeconfig.Config{
		Type: nodeconfig.TypeLocal,
		Details: &local.Details{
			Home: "/home/user/.cosmos",
		},
	}

	bz, err := yaml.Marshal(&config)
	require.NoError(t, err)

	var expected = `
type: local
config:
    home: /home/user/.cosmos
`
	require.Equal(t, strings.TrimLeft(expected, "\n"), string(bz))

	config = nodeconfig.Config{
		Type: nodeconfig.TypeRemote,
		Details: &remote.Details{
			RPC: &remote.RPCConfig{
				ClientName:     "juno",
				Address:        "http://localhost:26657",
				MaxConnections: 10,
			},
			GRPC: &remote.GRPCConfig{
				Address:  "http://localhost:9090",
				Insecure: true,
			},
		},
	}
	bz, err = yaml.Marshal(&config)
	require.NoError(t, err)

	expected = `
type: remote
config:
    rpc:
        client_name: juno
        address: http://localhost:26657
        max_connections: 10
    grpc:
        address: http://localhost:9090
        insecure: true
`
	require.Equal(t, strings.TrimLeft(expected, "\n"), string(bz))
}
