package telemetry

import "gopkg.in/yaml.v3"

// Config represents the configuration for the telemetry module
type Config struct {
	Port uint `yaml:"port"`
}

// ParseConfig allows to parse a byte array as a Config instance
func ParseConfig(bytes []byte) (*Config, error) {
	type T struct {
		Telemetry *Config `yaml:"telemetry"`
	}
	var cfg T
	err := yaml.Unmarshal(bytes, &cfg)
	return cfg.Telemetry, err
}
