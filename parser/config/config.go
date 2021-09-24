package config

type Config struct {
	Workers         int64  `yaml:"workers"`
	ParseNewBlocks  bool   `yaml:"listen_new_blocks"`
	ParseOldBlocks  bool   `yaml:"parse_old_blocks"`
	GenesisFilePath string `yaml:"genesis_file_path"`
	ParseGenesis    bool   `yaml:"parse_genesis"`
	StartHeight     int64  `yaml:"start_height"`
	FastSync        bool   `yaml:"fast_sync"`
}

// NewParsingConfig allows to build a new Config instance
func NewParsingConfig(
	workers int64,
	parseNewBlocks, parseOldBlocks bool,
	parseGenesis bool, genesisFilePath string,
	startHeight int64, fastSync bool,
) Config {
	return Config{
		Workers:         workers,
		ParseOldBlocks:  parseOldBlocks,
		ParseNewBlocks:  parseNewBlocks,
		ParseGenesis:    parseGenesis,
		GenesisFilePath: genesisFilePath,
		StartHeight:     startHeight,
		FastSync:        fastSync,
	}
}

// DefaultParsingConfig returns the default instance of Config
func DefaultParsingConfig() Config {
	return NewParsingConfig(
		1,
		true,
		true,
		true,
		"",
		1,
		false,
	)
}
