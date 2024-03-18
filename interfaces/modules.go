package interfaces

import "strings"

type Module interface {
	// Name returns the name of the module.
	Name() string
}

type BlockModule interface {
	// HandleBlock handles a block parsed by workers.
	HandleBlock(block Block) error
}

type AdditionalOperationsModule interface {
	// RunAdditionalOperations runs all the additional operations required by the module.
	// This is the perfect place where to initialize all the operations that subscribe to websockets or other
	// external sources.
	// NOTE. This method will only be run ONCE before starting the parsing of the blocks.
	RunAdditionalOperations() error
}

// --------------------------------------------------------------------------------------------------------------------

// Modules represents a slice of Module objects
type Modules []Module

// FindByName returns the module having the given name inside the m slice.
// If no modules are found, returns nil and false.
func (m Modules) FindByName(name string) (module Module, found bool) {
	for _, m := range m {
		if strings.EqualFold(m.Name(), name) {
			return m, true
		}
	}
	return nil, false
}
