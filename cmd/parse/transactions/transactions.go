package transactions

import (
	"fmt"

	parsecmdtypes "github.com/forbole/juno/v5/cmd/parse/types"
	"github.com/forbole/juno/v5/modules/cosmos"

	"github.com/rs/zerolog/log"

	"github.com/spf13/cobra"

	"github.com/forbole/juno/v5/types/config"
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
			infrastructures, err := parsecmdtypes.GetInfrastructures(config.Cfg, parseConfig)
			if err != nil {
				return err
			}

			module, found := infrastructures.Modules.FindByName("cosmos")
			if !found {
				return fmt.Errorf("cosmos module not found")
			}

			cosmosModule, ok := module.(*cosmos.Module)
			if !ok {
				return fmt.Errorf("unexpected module type")
			}

			// Get the flag values
			start, _ := cmd.Flags().GetInt64(flagStart)
			end, _ := cmd.Flags().GetInt64(flagEnd)

			// Get the start height, default to the config's height; use flagStart if set
			startHeight := config.Cfg.Parser.StartHeight
			if start > 0 {
				startHeight = start
			}

			// Get the end height, default to the node latest height; use flagEnd if set
			endBlock, err := infrastructures.Node.LatestBlock()
			if err != nil {
				return fmt.Errorf("error while getting chain latest block height: %s", err)
			}

			endHeight := endBlock.Height()
			if end > 0 {
				endHeight = end
			}

			log.Info().Int64("start height", startHeight).Int64("end height", endHeight).
				Msg("getting transactions...")
			for k := startHeight; k <= endHeight; k++ {
				log.Info().Int64("height", k).Msg("processing transactions...")
				err = cosmosModule.ProcessTransactions(k)
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
