package types

import "github.com/rs/zerolog/log"

// Logger defines a function that takes an error and logs it.
type Logger func(err error)

// ------------------------------------------------------------------------

var logGenesisError Logger = func(err error) {
	log.Error().Err(err).Msg("error while handling genesis")
}

// SetGenesisLogger sets the method that should be called each time a genesisHandler returns an error.
func SetGenesisLogger(logger Logger) {
	logGenesisError = logger
}

// LogGenesisError allows to log the given error as a genesis handling error.
func LogGenesisError(err error) {
	logGenesisError(err)
}

// ------------------------------------------------------------------------

var logBlockError Logger = func(err error) {
	log.Error().Err(err).Msg("error while handling block")
}

// SetBlockLogger sets the method that should be called when a blockHandler returns an error.
func SetBlockLogger(logger Logger) {
	logBlockError = logger
}

// LogBlockError allows to log the given error as a genesis handling error.
func LogBlockError(err error) {
	logBlockError(err)
}

// ------------------------------------------------------------------------

var logTxError Logger = func(err error) {
	log.Error().Err(err).Msg("error while handling transaction")
}

// SetTxLogger sets the method that should be called each time a transaction handler returns an error.
func SetTxLogger(logger Logger) {
	logTxError = logger
}

// LogTxError allows to log the given error as a genesis handling error.
func LogTxError(err error) {
	logTxError(err)
}

// ------------------------------------------------------------------------

var logMsgError Logger = func(err error) {
	log.Error().Err(err).Msg("error while handling message")
}

// SetMsgLogger sets the method that should be called when a message handler returns an error.
func SetMsgLogger(logger Logger) {
	logMsgError = logger
}

// LogMsgError allows to log the given error as a genesis handling error.
func LogMsgError(err error) {
	logMsgError(err)
}
