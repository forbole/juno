package blocks

import (
	"context"
	"fmt"

	parsecmdtypes "github.com/forbole/juno/v5/cmd/parse/types"
	"github.com/forbole/juno/v5/types/utils"

	"github.com/rs/zerolog/log"

	"github.com/spf13/cobra"

	"github.com/forbole/juno/v5/parser"
	"github.com/forbole/juno/v5/types/config"
)

const (
	flagStart = "start"
	flagEnd   = "end"
)

// newAllCmd returns a Cobra command that allows to fix missing blocks in database
func newAllCmd(parseConfig *parsecmdtypes.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "all",
		Short: "Reparse blocks and transactions ranged from the given start height to the given end height",
		Long: fmt.Sprintf(`Refetch all the blocks in the specified range and stores them inside the database. 
You can specify a custom blocks range by using the %s and %s flags. 
By default, all the blocks fetched from the node will not be stored inside the database if they are already present. 
`, flagStart, flagEnd),
		RunE: func(cmd *cobra.Command, args []string) error {
			infrastructures, err := parsecmdtypes.GetInfrastructures(config.Cfg, parseConfig)
			if err != nil {
				return err
			}

			// Get the flag values
			start, _ := cmd.Flags().GetInt64(flagStart)
			end, _ := cmd.Flags().GetInt64(flagEnd)

			lastDbBlockHeight, err := infrastructures.Database.GetLastBlockHeight()
			if err != nil {
				return err
			}

			// Compare start height from config file and last block height in database
			// and set higher block as start height
			startHeight := utils.MaxInt64(config.Cfg.Parser.StartHeight, lastDbBlockHeight)
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
				Msg("getting blocks and transactions")

			// Setup Worker and its context
			ctx := parser.NewContext(context.Background(), infrastructures.Node, infrastructures.Database, infrastructures.Logger, infrastructures.Modules)
			worker := parser.NewWorker(0)

			for k := startHeight; k <= endHeight; k++ {
				block, err := ctx.BlockNode().Block(k)
				if err != nil {
					return fmt.Errorf("error while fetching block %d: %s", k, err)
				}

				err = worker.Process(ctx, block)
				if err != nil {
					return fmt.Errorf("error while processing block %d: %s", k, err)
				}
			}

			return nil
		},
	}

	cmd.Flags().Int64(flagStart, 0, "Height from which to start getting missing blocks. If 0, the start height inside the config will be used instead")
	cmd.Flags().Int64(flagEnd, 0, "Height at which to finish getting missing. If 0, the latest height available inside the node will be used instead")

	return cmd
}
