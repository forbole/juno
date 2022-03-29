package config

import "time"

type Config struct {
	Workers         int64         `yaml:"workers"`
	ParseNewBlocks  bool          `yaml:"listen_new_blocks"`
	ParseOldBlocks  bool          `yaml:"parse_old_blocks"`
	GenesisFilePath string        `yaml:"genesis_file_path,omitempty"`
	ParseGenesis    bool          `yaml:"parse_genesis"`
	StartHeight     int64         `yaml:"start_height"`
	FastSync        bool          `yaml:"fast_sync,omitempty"`
	AvgBlockTime    time.Duration `yaml:"average_block_time"`
}

// NewParsingConfig allows to build a new Config instance
func NewParsingConfig(
	workers int64,
	parseNewBlocks, parseOldBlocks bool,
	parseGenesis bool, genesisFilePath string,
	startHeight int64, fastSync bool,
	avgBlockTime time.Duration,
) Config {
	return Config{
		Workers:         workers,
		ParseOldBlocks:  parseOldBlocks,
		ParseNewBlocks:  parseNewBlocks,
		ParseGenesis:    parseGenesis,
		GenesisFilePath: genesisFilePath,
		StartHeight:     startHeight,
		FastSync:        fastSync,
		AvgBlockTime:    avgBlockTime,
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
		5*time.Second,
	)
}
