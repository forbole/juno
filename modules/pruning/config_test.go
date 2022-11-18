package pruning_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/saifullah619/juno/v3/modules/pruning"
)

func TestParseConfig(t *testing.T) {
	data := []byte(`
pruning:
  keep_recent: 100
  keep_every: 10
  interval: 1
`)

	cfg, err := pruning.ParseConfig(data)
	require.NoError(t, err)

	require.NotNil(t, cfg)
	require.Equal(t, int64(100), cfg.KeepRecent)
	require.Equal(t, int64(10), cfg.KeepEvery)
	require.Equal(t, int64(1), cfg.Interval)

	data = []byte(`invalid_field: yes`)
	cfg, err = pruning.ParseConfig(data)
	require.NoError(t, err)
	require.Nil(t, cfg)
}
