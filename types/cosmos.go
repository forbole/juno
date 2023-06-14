package types

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	tmctypes "github.com/cometbft/cometbft/rpc/core/types"
	"github.com/cosmos/cosmos-sdk/types/tx"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Validator contains the data of a single validator
type Validator struct {
	ConsAddr   string
	ConsPubKey string
}

// NewValidator allows to build a new Validator instance
func NewValidator(consAddr string, consPubKey string) *Validator {
	return &Validator{
		ConsAddr:   consAddr,
		ConsPubKey: consPubKey,
	}
}

// -------------------------------------------------------------------------------------------------------------------

// CommitSig contains the data of a single validator commit signature
type CommitSig struct {
	Height           int64
	ValidatorAddress string
	VotingPower      int64
	ProposerPriority int64
	Timestamp        time.Time
}

// NewCommitSig allows to build a new CommitSign object
func NewCommitSig(validatorAddress string, votingPower, proposerPriority, height int64, timestamp time.Time) *CommitSig {
	return &CommitSig{
		Height:           height,
		ValidatorAddress: validatorAddress,
		VotingPower:      votingPower,
		ProposerPriority: proposerPriority,
		Timestamp:        timestamp,
	}
}

// -------------------------------------------------------------------------------------------------------------------

// Block contains the data of a single chain block
type Block struct {
	Height          int64
	Hash            string
	TxNum           int
	TotalGas        uint64
	ProposerAddress string
	Timestamp       time.Time
}

// NewBlock allows to build a new Block instance
func NewBlock(
	height int64, hash string, txNum int, totalGas uint64, proposerAddress string, timestamp time.Time,
) *Block {
	return &Block{
		Height:          height,
		Hash:            hash,
		TxNum:           txNum,
		TotalGas:        totalGas,
		ProposerAddress: proposerAddress,
		Timestamp:       timestamp,
	}
}

// NewBlockFromTmBlock builds a new Block instance from a given ResultBlock object
func NewBlockFromTmBlock(blk *tmctypes.ResultBlock, totalGas uint64) *Block {
	return NewBlock(
		blk.Block.Height,
		blk.Block.Hash().String(),
		len(blk.Block.Txs),
		totalGas,
		ConvertValidatorAddressToBech32String(blk.Block.ProposerAddress),
		blk.Block.Time,
	)
}

// -------------------------------------------------------------------------------------------------------------------

// Transaction represents an already existing blockchain transaction.
type Transaction struct {
	*sdk.TxResponse

	// Override these fields to apply the proper type since the Cosmos SDK encodes uint64 as strings
	Height    uint64 `json:"height,string,omitempty"`
	GasWanted uint64 `json:"gas_wanted,string,omitempty"`
	GasUsed   uint64 `json:"gas_used,string,omitempty"`

	// Override the Tx field to apply the custom type
	Tx *Tx `json:"tx,omitempty"`
}

// FindEventByType searches inside the given tx events for the message having the specified index, in order
// to find the event having the given type, and returns it.
// If no such event is found, returns an error instead.
func (tx Transaction) FindEventByType(index int, eventType string) (sdk.StringEvent, error) {
	for _, ev := range tx.Logs[index].Events {
		if ev.Type == eventType {
			return ev, nil
		}
	}

	return sdk.StringEvent{}, fmt.Errorf("no %s event found inside tx with hash %s", eventType, tx.TxHash)
}

// FindAttributeByKey searches inside the specified event of the given tx to find the attribute having the given key.
// If the specified event does not contain a such attribute, returns an error instead.
func (tx Transaction) FindAttributeByKey(event sdk.StringEvent, attrKey string) (string, error) {
	for _, attr := range event.Attributes {
		if attr.Key == attrKey {
			return attr.Value, nil
		}
	}

	return "", fmt.Errorf("no event with attribute %s found inside tx with hash %s", attrKey, tx.TxHash)
}

// Successful tells whether this tx is successful or not
func (tx Transaction) Successful() bool {
	return tx.TxResponse.Code == 0
}

// -------------------------------------------------------------------------------------------------------------------

// Tx represents the data of a single transaction.
// It embeds the Cosmos Tx type, but it overrides the Body field with a custom type.
type Tx struct {
	*tx.Tx
	Body     *TxBody   `json:"body,omitempty"`
	AuthInfo *AuthInfo `json:"auth_info,omitempty"`
}

// -------------------------------------------------------------------------------------------------------------------

// TxBody represents the data of a single transaction body.
// It embeds the Cosmos TxBody type, but it overrides the Messages field with a custom type.
type TxBody struct {
	*tx.TxBody
	TimeoutHeight uint64    `json:"timeout_height,string,omitempty"`
	Messages      []Message `json:"messages,omitempty"`
}

// UnmarshalJSON implements the json.Unmarshaler interface.
// This is done to properly unmarshal the Messages field by setting the Index field to the proper value.
func (tb *TxBody) UnmarshalJSON(data []byte) error {
	// Define a temporary type
	type TempTxBody struct {
		*tx.TxBody
		TimeoutHeight uint64            `json:"timeout_height,string,omitempty"`
		RawMessages   []json.RawMessage `json:"messages,omitempty"`
	}

	var temp TempTxBody
	err := json.Unmarshal(data, &temp)
	if err != nil {
		return fmt.Errorf("failed to unmarshal TxBody: %v", err)
	}

	tb.TxBody = temp.TxBody
	tb.TimeoutHeight = temp.TimeoutHeight

	// Initialize the Messages slice
	tb.Messages = make([]Message, len(temp.RawMessages))

	// Iterate over the RawMessages and populate the Messages slice
	for i, rawMsg := range temp.RawMessages {
		msg, err := UnmarshalMessage(i, rawMsg)
		if err != nil {
			return fmt.Errorf("failed to create message: %v", err)
		}

		tb.Messages[i] = msg
	}

	return nil
}

// AuthInfo represents the data of a single transaction auth info.
// It embeds the Cosmos AuthInfo type, but it overrides the SignerInfos and Fee fields with a custom type.
type AuthInfo struct {
	*tx.AuthInfo
	SignerInfos []*SignerInfo `json:"signer_infos,omitempty"`
	Fee         *Fee          `json:"fee"`
}

// SignerInfo represents the data of a single transaction signer info.
// It embeds the Cosmos SignerInfo type, but it overrides the PublicKey and Sequence fields with a custom type.
type SignerInfo struct {
	*tx.SignerInfo
	PublicKey json.RawMessage `json:"public_key,omitempty"`
	Sequence  uint64          `json:"sequence,string,omitempty"`
}

// Fee represents the data of a single transaction fee.
// It embeds the Cosmos Fee type, but it overrides the GasLimit field with a custom type.
type Fee struct {
	*tx.Fee
	GasLimit uint64 `json:"gas_limit,string,omitempty"`
}

// -------------------------------------------------------------------------------------------------------------------

type Message interface {
	GetType() string
	GetIndex() int
	GetBytes() json.RawMessage
}

// UnmarshalMessage can be used to unmarshal a Message instance from a JSON representation.
func UnmarshalMessage(i int, rawMsg json.RawMessage) (Message, error) {
	var temp struct {
		Type string `json:"@type"`
	}

	err := json.Unmarshal(rawMsg, &temp)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal Message: %v", err)
	}

	msg, err := unmarshalMessageWithType(i, temp.Type, rawMsg)
	if err != nil {
		return nil, fmt.Errorf("failed to create message: %v", err)
	}

	return msg, nil
}

func unmarshalMessageWithType(index int, msgType string, rawMsg json.RawMessage) (Message, error) {
	if strings.Contains(msgType, "MsgExec") {
		var msg MessageExec
		err := json.Unmarshal(rawMsg, &msg)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal Type1Message: %v", err)
		}
		msg.Index = index
		msg.Bytes = rawMsg
		return &msg, nil
	}

	var msg StandardMessage
	err := json.Unmarshal(rawMsg, &msg)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal StandardMessage: %v", err)
	}
	msg.Index = index
	msg.Bytes = rawMsg
	return &msg, nil
}

// -------------------------------------------------------------------------------------------------------------------

// StandardMessage represents the data of a single transaction message.
// It contains the raw bytes of the message, plus the type of the message. This is done in order to be able to
// support any kind of message agnosticly while still being able to decode the message bytes into a concrete type.
// It also contains
type StandardMessage struct {
	Index int
	Type  string `json:"@type"`
	Bytes json.RawMessage
}

func (msg *StandardMessage) GetType() string {
	return msg.Type
}

func (msg *StandardMessage) GetIndex() int {
	return msg.Index
}

func (msg *StandardMessage) GetBytes() json.RawMessage {
	return msg.Bytes
}

// UnmarshalJSON allows to unmarshal a Message from a JSON representation.
func (msg *StandardMessage) UnmarshalJSON(data []byte) error {
	// Define a temporary type
	type TempMessage struct {
		Type string `json:"@type"`
	}

	var temp TempMessage
	err := json.Unmarshal(data, &temp)
	if err != nil {
		return fmt.Errorf("failed to unmarshal Message: %v", err)
	}

	msg.Type = temp.Type
	msg.Bytes = data

	return nil
}

// MarshalJSON allows to marshal a Message into a JSON representation.
func (msg *StandardMessage) MarshalJSON() ([]byte, error) {
	return msg.Bytes, nil
}

// -------------------------------------------------------------------------------------------------------------------

// MessageExec represents the Cosmos SDK MsgExec type.
// It embeds the StandardMessage type, and it adds the Messages field to represent all
// the messages that are executed.
type MessageExec struct {
	*StandardMessage `json:"-"`
	Messages         []Message `json:"msgs"`
}

// UnmarshalJSON allows to unmarshal a MessageExec from a JSON representation.
func (msg *MessageExec) UnmarshalJSON(data []byte) error {
	// Define a temporary type
	type TempMessageExec struct {
		Type     string            `json:"@type"`
		Messages []json.RawMessage `json:"msgs"`
	}

	var temp TempMessageExec
	err := json.Unmarshal(data, &temp)
	if err != nil {
		return fmt.Errorf("failed to unmarshal MessageExec: %v", err)
	}

	msg.StandardMessage = &StandardMessage{Type: temp.Type, Bytes: data}
	msg.Messages = make([]Message, len(temp.Messages))

	for i, rawMsg := range temp.Messages {
		message, err := UnmarshalMessage(i, rawMsg)
		if err != nil {
			return fmt.Errorf("failed to create message: %v", err)
		}
		msg.Messages[i] = message
	}

	return nil
}
