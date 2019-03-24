package main

type config struct {
	Node string         `toml:"node"`
	DB   databaseConfig `toml:"database"`
}

type databaseConfig struct {
	Host     string `toml:"host"`
	Port     uint64 `toml:"port"`
	Name     string `toml:"name"`
	User     string `toml:"user"`
	Password string `toml:"password"`
}
