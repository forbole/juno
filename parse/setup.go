package parse

import (
	"fmt"
	"os"

	"github.com/desmos-labs/juno/config"
	"github.com/desmos-labs/juno/db"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
)

// SetupConfig takes the path to a configuration file and returns the properly parsed configuration
func SetupConfig(configPath string) (*config.Config, error) {
	log.Debug().Msg("Reading config file")
	cfg, err := config.ParseConfig(configPath)
	if err != nil {
		return nil, err
	}
	return cfg, err
}

// SetupLogging setups the logging for the entire project
func SetupLogging() error {
	// Init logging level
	logLvl, err := zerolog.ParseLevel(viper.GetString(config.FlagLogLevel))
	if err != nil {
		return err
	}
	zerolog.SetGlobalLevel(logLvl)

	// Init logging format
	logFormat := viper.GetString(config.FlagLogFormat)
	switch logFormat {
	case logLevelJSON:
		// JSON is the default logging format
		break

	case logLevelText:
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
		break

	default:
		return fmt.Errorf("invalid logging format: %s", logFormat)
	}
	return err
}

// AdditionalOperation represents a single additional operation that should be done when
// the setup is completed. It receives the same configuration and database instances
// that are going to be used later during the parsing.
type AdditionalOperation = func(cfg config.Config, db db.Database) error
