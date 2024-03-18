package cosmos

import (
	"fmt"

	"github.com/forbole/juno/v5/interfaces"
)

var _ interfaces.AdditionalOperationsModule = &Module{}

// RunAdditionalOperations implements modules.AdditionalOperationsModule
func (m *Module) RunAdditionalOperations() error {
	return m.runAdditionalOperations()
}

// runAdditionalOperations runs the module additional operations
func (m *Module) runAdditionalOperations() error {
	if m.parseGenesis {
		// Handle genesis parsing
		genesisDoc, genesisState, err := GetGenesisDocAndState(m.genesisFilePath, m.source)
		if err != nil {
			return fmt.Errorf("failed to get genesis: %s", err)
		}

		for _, module := range m.modules {
			if genesisModule, ok := module.(GenesisModule); ok {
				if err := genesisModule.HandleGenesis(genesisDoc, genesisState); err != nil {
					return err
				}
			}
		}
	}

	return nil
}
