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
	cfg := types.Cfg.GetLoggingConfig()

	// Init logging level
	logLvl, err := zerolog.ParseLevel(cfg.LogLevel)
	if err != nil {
		return err
	}
	zerolog.SetGlobalLevel(logLvl)

	// Init logging format
	switch cfg.LogFormat {
	case "json":
		// JSON is the default logging format
		break

	case "text":
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
		break

	default:
		return fmt.Errorf("invalid logging format: %s", cfg.LogFormat)
	}
	return err
}
