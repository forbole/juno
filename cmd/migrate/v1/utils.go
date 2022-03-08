package v1

import (
	"github.com/pelletier/go-toml"

	"github.com/forbole/juno/v3/cmd/migrate/utils"
)

// ParseConfig attempts to read and parse a Juno Config from the given string bytes.
// An error reading or parsing the Config results in a panic.
func ParseConfig(configData []byte) (Config, error) {
	var cfg Config
	err := toml.Unmarshal(configData, &cfg)
	return cfg, err
}

func GetConfig() (Config, error) {
	bz, err := utils.ReadConfig()
	if err != nil {
		return Config{}, nil
	}

	return ParseConfig(bz)
}
