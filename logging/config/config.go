package config

import "github.com/rs/zerolog"

type Config struct {
	LogLevel  string `yaml:"level"`
	LogFormat string `yaml:"format"`
}

// NewLoggingConfig returns a new Config instance
func NewLoggingConfig(level, format string) Config {
	return Config{
		LogLevel:  level,
		LogFormat: format,
	}
}

// DefaultLoggingConfig returns the default Config instance
func DefaultLoggingConfig() Config {
	return NewLoggingConfig(zerolog.DebugLevel.String(), "text")
}
