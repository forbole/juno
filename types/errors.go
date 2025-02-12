package types

import "errors"

var (
	ErrModuleSynced = errors.New("module has caught up with main chain")
)
