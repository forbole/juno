package blocks

import (
	"fmt"
	"strconv"

	parsecmdtypes "github.com/forbole/juno/v4/cmd/parse/types"

	"github.com/spf13/cobra"

	"github.com/forbole/juno/v4/parser"
	"github.com/forbole/juno/v4/types/config"
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

			parseCtx, err := parsecmdtypes.GetParserContext(config.Cfg, parseConfig)
			if err != nil {
				return err
			}

			workerCtx := parser.NewContext(parseCtx.EncodingConfig, parseCtx.Node, parseCtx.Database, parseCtx.Logger, parseCtx.Modules)
			worker := parser.NewWorker(workerCtx, nil, 0)

			dbLastHeight, err := parseCtx.Database.GetLastBlockHeight()
			if err != nil {
				return fmt.Errorf("error while getting DB last block height: %s", err)
			}

			for _, k := range parseCtx.Database.GetMissingHeights(startHeight, dbLastHeight) {
				err = worker.Process(k)
				if err != nil {
					return fmt.Errorf("error while re-fetching block %d: %s", k, err)
				}
			}

			return nil
		},
	}

	return cmd
}
