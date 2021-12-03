package parser

import (
	"encoding/json"
	"fmt"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	tmjson "github.com/tendermint/tendermint/libs/json"
	tmos "github.com/tendermint/tendermint/libs/os"
	tmctypes "github.com/tendermint/tendermint/rpc/core/types"
	tmtypes "github.com/tendermint/tendermint/types"

	"github.com/forbole/juno/v2/node"
	"github.com/forbole/juno/v2/types"
)

// findValidatorByAddr finds a validator by a consensus address given a set of
// Tendermint validators for a particular block. If no validator is found, nil
// is returned.
func findValidatorByAddr(consAddr string, vals *tmctypes.ResultValidators) *tmtypes.Validator {
	for _, val := range vals.Validators {
		if consAddr == sdk.ConsAddress(val.Address).String() {
			return val
		}
	}

	return nil
}

// sumGasTxs returns the total gas consumed by a set of transactions.
func sumGasTxs(txs []*types.Tx) uint64 {
	var totalGas uint64

	for _, tx := range txs {
		totalGas += uint64(tx.GasUsed)
	}

	return totalGas
}

func GetGenesisDocAndState(genesisPath string, node node.Node) (*tmtypes.GenesisDoc, map[string]json.RawMessage, error) {
	var genesisDoc *tmtypes.GenesisDoc
	var err error
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
	err = json.Unmarshal(genesisDoc.AppState, &genesisState)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to unmarshal genesis state: %s", err)
	}

	return genesisDoc, genesisState, nil
}
