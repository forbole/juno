package transactions

import (
	"fmt"

	parsecmdtypes "github.com/saifullah619/juno/v3/cmd/parse/types"

	"github.com/rs/zerolog/log"

	"github.com/spf13/cobra"

	"github.com/saifullah619/juno/v3/parser"
	"github.com/saifullah619/juno/v3/types/config"
)

const (
	flagStart = "start"
	flagEnd   = "end"
)

// newTransactionsCmd returns a Cobra command that allows to fix missing or incomplete transactions in database
func newTransactionsCmd(parseConfig *parsecmdtypes.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "all",
		Short: "Parse missing or incomplete transactions",
		Long: fmt.Sprintf(`Refetch missing or incomplete transactions and store them inside the database. 
You can specify a custom height range by using the %s and %s flags. 
`, flagStart, flagEnd),
		RunE: func(cmd *cobra.Command, args []string) error {
			parseCtx, err := parsecmdtypes.GetParserContext(config.Cfg, parseConfig)
			if err != nil {
				return err
			}

			workerCtx := parser.NewContext(parseCtx.EncodingConfig, parseCtx.Node, parseCtx.Database, parseCtx.Logger, parseCtx.Modules)
			worker := parser.NewWorker(workerCtx, nil, 0)

			// Get the flag values
			start, _ := cmd.Flags().GetInt64(flagStart)
			end, _ := cmd.Flags().GetInt64(flagEnd)

			// Get the start height, default to the config's height; use flagStart if set
			startHeight := config.Cfg.Parser.StartHeight
			if start > 0 {
				startHeight = start
			}

			// Get the end height, default to the node latest height; use flagEnd if set
			endHeight, err := parseCtx.Node.LatestHeight()
			if err != nil {
				return fmt.Errorf("error while getting chain latest block height: %s", err)
			}
			if end > 0 {
				endHeight = end
			}

			log.Info().Int64("start height", startHeight).Int64("end height", endHeight).
				Msg("getting transactions...")
			for k := startHeight; k <= endHeight; k++ {
				log.Info().Int64("height", k).Msg("processing transactions...")
				err = worker.ProcessTransactions(k)
				if err != nil {
					return fmt.Errorf("error while re-fetching transactions of height %d: %s", k, err)
				}
			}

			return nil
		},
	}

	cmd.Flags().Int64(flagStart, 0, "Height from which to start fetching missing transactions. If 0, the start height inside the config file will be used instead")
	cmd.Flags().Int64(flagEnd, 0, "Height at which to finish fetching missing transactions. If 0, the latest height available inside the node will be used instead")

	return cmd
}
