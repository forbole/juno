package types

import (
	"fmt"
	"github.com/cosmos/cosmos-sdk/x/auth/legacy/legacytx"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Tx represents an already existing blockchain transaction
type Tx struct {
	legacytx.StdTx
	sdk.TxResponse
	Signatures []Signature
	Logs       sdk.ABCIMessageLogs
}

// NewTx allows to create a new Tx instance from the given txResponse
func NewTx(txResponse sdk.TxResponse) (*Tx, error) {
	stdTx, ok := txResponse.Tx.GetCachedValue().(legacytx.StdTx)
	if !ok {
		return nil, fmt.Errorf("unsupported tx type: %T", txResponse.Tx)
	}

	// Convert Tendermint signatures into a more human-readable format
	sigs := make([]Signature, len(stdTx.Signatures))
	for i, sig := range stdTx.Signatures {
		sigs[i] = Signature{
			StdSignature: sig,
			Address:      sdk.AccAddress(sig.Address()).String(),
		}
	}

	return &Tx{
		TxResponse: txResponse,
		StdTx:      stdTx,
		Signatures: sigs,
		Logs:       txResponse.Logs,
	}, nil
}

// FindEventByType searches inside the given tx events for the message having the specified index, in order
// to find the event having the given type, and returns it.
// If no such event is found, returns an error instead.
func (tx Tx) FindEventByType(index int, eventType string) (sdk.StringEvent, error) {
	for _, ev := range tx.Logs[index].Events {
		if ev.Type == eventType {
			return ev, nil
		}
	}

	return sdk.StringEvent{}, fmt.Errorf("no %s event found inside tx with hash %s", eventType, tx.TxHash)
}

// FindAttributeByKey searches inside the specified event of the given tx to find the attribute having the given key.
// If the specified event does not contain a such attribute, returns an error instead.
func (tx Tx) FindAttributeByKey(event sdk.StringEvent, attrKey string) (string, error) {
	for _, attr := range event.Attributes {
		if attr.Key == attrKey {
			return attr.Value, nil
		}
	}

	return "", fmt.Errorf("no event with attribute %s found inside tx with hash %s", attrKey, tx.TxHash)
}

// Successful tells whether this tx is successful or not
func (tx Tx) Successful() bool {
	return tx.Code == 0
}

// Signature wraps auth.StdSignature adding the address of the signer
type Signature struct {
	legacytx.StdSignature
	Address string `json:"address,omitempty"`
}
