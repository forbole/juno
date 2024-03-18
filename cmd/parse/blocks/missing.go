package blocks

import (
	"context"
	"fmt"
	"strconv"

	parsecmdtypes "github.com/forbole/juno/v5/cmd/parse/types"

	"github.com/spf13/cobra"

	"github.com/forbole/juno/v5/parser"
	"github.com/forbole/juno/v5/types/config"
)

// newMissingCmd returns a Cobra command that allows to fix missing blocks in database
func newMissingCmd(parseConfig *parsecmdtypes.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "missing [start height]",
		Short: "Refetch all the missing heights in the database starting from the given start height",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			startHeight, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("make sure the given start height is a positive integer")
			}

			infrastructures, err := parsecmdtypes.GetInfrastructures(config.Cfg, parseConfig)
			if err != nil {
				return err
			}

			dbLastHeight, err := infrastructures.Database.GetLastBlockHeight()
			if err != nil {
				return fmt.Errorf("error while getting DB last block height: %s", err)
			}

			// Setup the context and the worker
			ctx := parser.NewContext(context.Background(), infrastructures.Node, infrastructures.Database, infrastructures.Logger, infrastructures.Modules)
			worker := parser.NewWorker(0)

			for _, height := range infrastructures.Database.GetMissingHeights(startHeight, dbLastHeight) {
				block, err := infrastructures.Node.Block(height)
				if err != nil {
					return fmt.Errorf("error while fetching block %d: %s", height, err)
				}

				err = worker.Process(ctx, block)
				if err != nil {
					return fmt.Errorf("error while processing block %d: %s", height, err)
				}
			}

			return nil
		},
	}

	return cmd
}
