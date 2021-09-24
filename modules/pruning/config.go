package pruning

import (
	"gopkg.in/yaml.v3"
)

type Config struct {
	KeepRecent int64 `yaml:"keep_recent"`
	KeepEvery  int64 `yaml:"keep_every"`
	Interval   int64 `yaml:"interval"`
}

func ParseConfig(bz []byte) (*Config, error) {
	type T struct {
		Config *Config `yaml:"pruning"`
	}
	var cfg T
	err := yaml.Unmarshal(bz, &cfg)
	return cfg.Config, err
}
