package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
)

// Tx represents an already existing blockchain transaction
type Tx struct {
	sdk.TxResponse
	Messages   []sdk.Msg   `json:"messages"`
	Fee        auth.StdFee `json:"fee"`
	Signatures []Signature `json:"signatures"`
	Memo       string      `json:"memo"`
}

// NewTx allows to create a new Tx instance from the given txResponse
func NewTx(txResponse sdk.TxResponse) (*Tx, error) {
	stdTx, ok := txResponse.Tx.(auth.StdTx)
	if !ok {
		return nil, fmt.Errorf("unsupported tx type: %T", txResponse.Tx)
	}

	// Convert Tendermint signatures into a more human-readable format
	sigs := make([]Signature, len(stdTx.Signatures), len(stdTx.Signatures))
	for i, sig := range stdTx.Signatures {
		sigs[i] = Signature{
			StdSignature: sig,
			Address:      sdk.AccAddress(sig.Address()).String(),
		}
	}

	return &Tx{
		TxResponse: txResponse,
		Fee:        stdTx.Fee,
		Messages:   stdTx.GetMsgs(),
		Signatures: sigs,
		Memo:       stdTx.Memo,
	}, nil
}

// Signature wraps auth.StdSignature adding the address of the signer
type Signature struct {
	auth.StdSignature
	Address string `json:"address,omitempty"`
}
