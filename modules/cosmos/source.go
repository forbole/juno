package cosmos

import (
	tmctypes "github.com/cometbft/cometbft/rpc/core/types"
	"github.com/forbole/juno/v5/types"
)

type Source interface {
	// Genesis returns the genesis state
	Genesis() (*tmctypes.ResultGenesis, error)

	// ChainID returns the network ID
	ChainID() (string, error)

	// Validators returns all the known Tendermint validators for a given block
	// height. An error is returned if the query fails.
	Validators(height int64) (*tmctypes.ResultValidators, error)

	// ResultBlock queries for a result block by height. An error is returned if the query fails.
	ResultBlock(height int64) (*tmctypes.ResultBlock, error)

	// BlockResults queries the results of a block by height. An error is returnes if the query fails
	BlockResults(height int64) (*tmctypes.ResultBlockResults, error)

	// Txs queries for all the transactions in a block. Transactions are returned
	// in the sdk.TxResponse format which internally contains an sdk.Tx. An error is
	// returned if any query fails.
	Txs(block *tmctypes.ResultBlock) ([]*types.Tx, error)
}
