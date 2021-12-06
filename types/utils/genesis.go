package utils

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/forbole/juno/v2/node"

	tmjson "github.com/tendermint/tendermint/libs/json"
	tmos "github.com/tendermint/tendermint/libs/os"
	tmtypes "github.com/tendermint/tendermint/types"
)

// GetGenesisDocAndState reads the genesis from node or file and returns genesis doc and state
func GetGenesisDocAndState(genesisPath string, node node.Node) (*tmtypes.GenesisDoc, map[string]json.RawMessage, error) {
	var genesisDoc *tmtypes.GenesisDoc
	if strings.TrimSpace(genesisPath) != "" {
		bz, err := tmos.ReadFile(genesisPath)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to read genesis file: %s", err)
		}

		err = tmjson.Unmarshal(bz, &genesisDoc)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to unmarshal genesis doc: %s", err)
		}

	} else {
		response, err := node.Genesis()
		if err != nil {
			return nil, nil, fmt.Errorf("failed to get genesis: %s", err)
		}
		genesisDoc = response.Genesis
	}

	var genesisState map[string]json.RawMessage
	err := json.Unmarshal(genesisDoc.AppState, &genesisState)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to unmarshal genesis state: %s", err)
	}

	return genesisDoc, genesisState, nil
}
