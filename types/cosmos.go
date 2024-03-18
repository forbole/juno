package types

import (
	"fmt"
	"time"

	tmctypes "github.com/cometbft/cometbft/rpc/core/types"
	"github.com/forbole/juno/v5/interfaces"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx"
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

var _ interfaces.Block = &Block{}

// Block contains the data of a single chain block
type Block struct {
	height    int64
	hash      string
	TxNum     int
	TotalGas  uint64
	proposer  string
	timestamp time.Time
	value     interface{}
}

// Height returns the height of the block
func (b Block) Height() int64 {
	return b.height
}

// Hash returns the hash of the block
func (b Block) Hash() string {
	return b.hash
}

// Timestamp returns the timestamp of the block
func (b Block) Timestamp() time.Time {
	return b.timestamp
}

// Proposer returns the address of the block proposer
func (b Block) Proposer() string {
	return b.proposer
}

func (b Block) Value() interface{} {
	return b.value
}

// NewBlock allows to build a new Block instance
func NewBlock(
	height int64, hash string, txNum int, totalGas uint64, proposerAddress string, timestamp time.Time, value interface{},
) *Block {
	return &Block{
		height:    height,
		hash:      hash,
		TxNum:     txNum,
		TotalGas:  totalGas,
		proposer:  proposerAddress,
		timestamp: timestamp,
		value:     value,
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
		blk,
	)
}

// -------------------------------------------------------------------------------------------------------------------

// Tx represents an already existing blockchain transaction
type Tx struct {
	*tx.Tx
	*sdk.TxResponse
}

// NewTx allows to create a new Tx instance from the given txResponse
func NewTx(txResponse *sdk.TxResponse, tx *tx.Tx) (*Tx, error) {
	return &Tx{
		Tx:         tx,
		TxResponse: txResponse,
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
	return tx.TxResponse.Code == 0
}

// -------------------------------------------------------------------------------------------------------------------

// Message represents the data of a single message
type Message struct {
	TxHash    string
	Index     int
	Type      string
	Value     string
	Addresses []string
	Height    int64
}

// NewMessage allows to build a new Message instance
func NewMessage(txHash string, index int, msgType string, value string, addresses []string, height int64) *Message {
	return &Message{
		TxHash:    txHash,
		Index:     index,
		Type:      msgType,
		Value:     value,
		Addresses: addresses,
		Height:    height,
	}
}
