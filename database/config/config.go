package config

import "net/url"

type Config struct {
	URL                string `yaml:"url"`
	MaxOpenConnections int    `yaml:"max_open_connections"`
	MaxIdleConnections int    `yaml:"max_idle_connections"`
	PartitionSize      int64  `yaml:"partition_size"`
	PartitionBatchSize int64  `yaml:"partition_batch"`
}

func (c *Config) getURL() *url.URL {
	parsedURL, err := url.Parse(c.URL)
	if err != nil {
		panic(err)
	}
	return parsedURL
}

func (c *Config) GetUser() string {
	return c.getURL().User.Username()
}

func (c *Config) GetPassword() string {
	password, _ := c.getURL().User.Password()
	return password
}

func (c *Config) GetHost() string {
	return c.getURL().Host
}

func (c *Config) GetPort() string {
	return c.getURL().Port()
}

func (c *Config) GetSchema() string {
	return c.getURL().Query().Get("search_path")
}

func (c *Config) GetSSLMode() string {
	return c.getURL().Query().Get("sslmode")
}

func NewDatabaseConfig(
	url string,
	maxOpenConnections int, maxIdleConnections int,
	partitionSize int64, batchSize int64,
) Config {
	return Config{
		URL:                url,
		MaxOpenConnections: maxOpenConnections,
		MaxIdleConnections: maxIdleConnections,
		PartitionSize:      partitionSize,
		PartitionBatchSize: batchSize,
	}
}

// DefaultDatabaseConfig returns the default instance of Config
func DefaultDatabaseConfig() Config {
	return NewDatabaseConfig(
		"postgresql://user:password@localhost:5432/database-name?sslmode=disable&search_path=public",
		1,
		1,
		100000,
		1000,
	)
}
