package parse

import (
	"fmt"
	"os"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"

	"github.com/desmos-labs/juno/types"
)

// setupLogging setups the logging for the entire project
func setupLogging(_ *cobra.Command, _ []string) error {
	// Init logging level
	logLvl, err := zerolog.ParseLevel(types.Cfg.Logging.LogLevel)
	if err != nil {
		return err
	}
	zerolog.SetGlobalLevel(logLvl)

	// Init logging format
	switch types.Cfg.Logging.LogFormat {
	case "json":
		// JSON is the default logging format
		break

	case "text":
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
		break

	default:
		return fmt.Errorf("invalid logging format: %s", types.Cfg.Logging.LogFormat)
	}
	return err
}
