package utils

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/forbole/juno/v4/node"

	tmjson "github.com/tendermint/tendermint/libs/json"
	tmos "github.com/tendermint/tendermint/libs/os"
	tmtypes "github.com/tendermint/tendermint/types"
)

// ReadGenesisFileGenesisDoc reads the genesis file located at the given path
func ReadGenesisFileGenesisDoc(genesisPath string) (*tmtypes.GenesisDoc, error) {
	var genesisDoc *tmtypes.GenesisDoc
	bz, err := tmos.ReadFile(genesisPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read genesis file: %s", err)
	}

	err = tmjson.Unmarshal(bz, &genesisDoc)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal genesis doc: %s", err)
	}

	return genesisDoc, nil
}

// GetGenesisState returns the genesis state by getting it from the given genesis doc
func GetGenesisState(doc *tmtypes.GenesisDoc) (map[string]json.RawMessage, error) {
	var genesisState map[string]json.RawMessage
	err := json.Unmarshal(doc.AppState, &genesisState)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal genesis state: %s", err)
	}
	return genesisState, nil
}

// GetGenesisDocAndState reads the genesis from node or file and returns genesis doc and state
func GetGenesisDocAndState(genesisPath string, node node.Node) (*tmtypes.GenesisDoc, map[string]json.RawMessage, error) {
	var genesisDoc *tmtypes.GenesisDoc
	if strings.TrimSpace(genesisPath) != "" {
		genDoc, err := ReadGenesisFileGenesisDoc(genesisPath)
		if err != nil {
			return nil, nil, fmt.Errorf("error while reading genesis file: %s", err)
		}
		genesisDoc = genDoc

	} else {
		response, err := node.Genesis()
		if err != nil {
			return nil, nil, fmt.Errorf("failed to get genesis: %s", err)
		}
		genesisDoc = response.Genesis
	}

	genesisState, err := GetGenesisState(genesisDoc)
	if err != nil {
		return nil, nil, err
	}

	return genesisDoc, genesisState, nil
}
