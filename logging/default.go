package logging

import (
	"fmt"
	"os"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/forbole/juno/v5/interfaces"
)

var (
	_ interfaces.Logger = &defaultLogger{}
)

// defaultLogger represents the default logger for any kind of error
type defaultLogger struct {
	Logger zerolog.Logger
}

// DefaultLogger allows to build a new defaultLogger instance
func DefaultLogger() interfaces.Logger {
	return &defaultLogger{
		Logger: log.Logger,
	}
}

// SetLogLevel implements Logger
func (d *defaultLogger) SetLogLevel(level string) error {
	logLvl, err := zerolog.ParseLevel(level)
	if err != nil {
		return err
	}

	zerolog.SetGlobalLevel(logLvl)
	return nil
}

// SetLogFormat implements Logger
func (d *defaultLogger) SetLogFormat(format string) error {
	switch format {
	case "json":
		// JSON is the default logging format
		break

	case "text":
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	default:
		return fmt.Errorf("invalid logging format: %s", format)
	}

	return nil
}

// Info implements Logger
func (d *defaultLogger) Info(msg string, keyVals ...interface{}) {
	d.Logger.Info().Fields(getLogFields(keyVals...)).Msg(msg)
}

// Debug implements Logger
func (d *defaultLogger) Debug(msg string, keyVals ...interface{}) {
	d.Logger.Debug().Fields(getLogFields(keyVals...)).Msg(msg)
}

// Error implements Logger
func (d *defaultLogger) Error(msg string, keyVals ...interface{}) {
	ErrorCount.Inc()
	d.Logger.Error().Fields(getLogFields(keyVals...)).Msg(msg)
}

func getLogFields(keyVals ...interface{}) map[string]interface{} {
	if len(keyVals)%2 != 0 {
		return nil
	}

	fields := make(map[string]interface{})
	for i := 0; i < len(keyVals); i += 2 {
		fields[keyVals[i].(string)] = keyVals[i+1]
	}

	return fields
}
