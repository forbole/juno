package db

import (
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/fissionlabsio/juno/codec"
	tmctypes "github.com/tendermint/tendermint/rpc/core/types"
	tmtypes "github.com/tendermint/tendermint/types"
)

// parseTxs parses a set of transactions returned by Tendermint into StdTx
// transactions. It returns an error if any unmarshalling fails.
func parseTxs(txs []*tmctypes.ResultTx) ([]auth.StdTx, error) {
	stdTxs := make([]auth.StdTx, len(txs))

	for i, tx := range txs {
		var stdTx auth.StdTx

		err := codec.Codec.UnmarshalBinaryLengthPrefixed(tx.Tx, &stdTx)
		if err != nil {
			return nil, err
		}

		stdTxs[i] = stdTx
	}

	return stdTxs, nil
}

// findValidatorByAddr finds a validator by a HEX address given a set of
// Tendermint validators for a particular block. If no validator is found, nil
// is returned.
func findValidatorByAddr(addrHex string, vals *tmctypes.ResultValidators) *tmtypes.Validator {
	for _, val := range vals.Validators {
		if strings.ToLower(addrHex) == strings.ToLower(val.Address.String()) {
			return val
		}
	}

	return nil
}

// sumGasTxs returns the total gas consumed by a set of transactions.
func sumGasTxs(txs []sdk.TxResponse) uint64 {
	var totalGas uint64

	for _, tx := range txs {
		totalGas += uint64(tx.GasUsed)
	}

	return totalGas
}
