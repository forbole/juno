package config

type Config struct {
	Name               string `yaml:"name"`
	Host               string `yaml:"host"`
	Port               int64  `yaml:"port"`
	User               string `yaml:"user"`
	Password           string `yaml:"password"`
	SSLMode            string `yaml:"ssl_mode,omitempty"`
	Schema             string `yaml:"schema,omitempty"`
	MaxOpenConnections int    `yaml:"max_open_connections"`
	MaxIdleConnections int    `yaml:"max_idle_connections"`
	PartitionSize      int64  `yaml:"partition_size"`
	PartitionBatchSize int64  `yaml:"partition_batch"`
}

func NewDatabaseConfig(
	name, host string, port int64, user string, password string,
	sslMode string, schema string,
	maxOpenConnections int, maxIdleConnections int,
	partitionSize int64, batchSize int64,
) Config {
	return Config{
		Name:               name,
		Host:               host,
		Port:               port,
		User:               user,
		Password:           password,
		SSLMode:            sslMode,
		Schema:             schema,
		MaxOpenConnections: maxOpenConnections,
		MaxIdleConnections: maxIdleConnections,
		PartitionSize:      partitionSize,
		PartitionBatchSize: batchSize,
	}
}

// DefaultDatabaseConfig returns the default instance of Config
func DefaultDatabaseConfig() Config {
	return NewDatabaseConfig(
		"database-name",
		"localhost",
		5432,
		"user",
		"password",
		"",
		"public",
		1,
		1,
		100000,
		1000,
	)
}
