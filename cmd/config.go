package cmd

import (
	initcmd "github.com/forbole/juno/v4/cmd/init"
	parsecmd "github.com/forbole/juno/v4/cmd/parse/types"
)

// Config represents the general configuration for the commands
type Config struct {
	name        string
	initConfig  *initcmd.Config
	parseConfig *parsecmd.Config
}

// NewConfig allows to build a new Config instance
func NewConfig(name string) *Config {
	return &Config{
		name: name,
	}
}

// GetName returns the name of the root command
func (c *Config) GetName() string {
	return c.name
}

// WithInitConfig sets cfg as the parse command configuration
func (c *Config) WithInitConfig(cfg *initcmd.Config) *Config {
	c.initConfig = cfg
	return c
}

// GetInitConfig returns the currently set parse configuration
func (c *Config) GetInitConfig() *initcmd.Config {
	if c.initConfig == nil {
		return initcmd.NewConfig()
	}
	return c.initConfig
}

// WithParseConfig sets cfg as the parse command configuration
func (c *Config) WithParseConfig(cfg *parsecmd.Config) *Config {
	c.parseConfig = cfg
	return c
}

// GetParseConfig returns the currently set parse configuration
func (c *Config) GetParseConfig() *parsecmd.Config {
	if c.parseConfig == nil {
		return parsecmd.NewConfig()
	}
	return c.parseConfig
}
