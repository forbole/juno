package transactions

import (
	"github.com/spf13/cobra"

	parsecmdtypes "github.com/forbole/juno/v4/cmd/parse/types"
)

// NewTransactionsCmd returns the Cobra command that allows to fix missing or incomplete transactions
func NewTransactionsCmd(parseConfig *parsecmdtypes.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "transactions",
		Short: "Parse things related to transactions",
	}

	cmd.AddCommand(
		newTransactionsCmd(parseConfig),
	)

	return cmd
}
