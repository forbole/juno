package blocks

import (
	"fmt"

	"github.com/forbole/juno/v2/cmd/parse"
	"github.com/forbole/juno/v2/types/config"

	"github.com/spf13/cobra"

	"github.com/forbole/juno/v2/parser"
)

const (
	flagForce = "force"
	flagStart = "start"
	flagEnd   = "end"
)

// blocksCmd returns a Cobra command that allows to fix missing blocks in database
func blocksCmd(parseConfig *parse.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "all",
		Short: "Fix missing blocks and transactions in database",
		RunE: func(cmd *cobra.Command, args []string) error {

			parseCtx, err := parse.GetParsingContext(parseConfig)
			if err != nil {
				return err
			}

			workerCtx := parser.NewContext(parseCtx.EncodingConfig.Marshaler, nil, parseCtx.Node, parseCtx.Database, parseCtx.Logger, parseCtx.Modules)
			worker := parser.NewWorker(0, workerCtx)

			// Get the flag values
			start, _ := cmd.Flags().GetInt64(flagStart)
			end, _ := cmd.Flags().GetInt64(flagEnd)
			force, _ := cmd.Flags().GetBool(flagForce)

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

			fmt.Printf("Refetching missing blocks and transactions from height %d to %d \n", startHeight, endHeight)
			for k := startHeight; k <= endHeight; k++ {
				if force {
					err = worker.Process(k)
				} else {
					err = worker.ProcessIfNotExists(k)
				}

				if err != nil {
					return fmt.Errorf("error while re-fetching block %d: %s", k, err)
				}
			}

			return nil
		},
	}

	cmd.Flags().Bool(flagForce, false, "If set, forces the fetch of blocks by overwriting any existing ones")
	cmd.Flags().Int64(flagStart, 0, "Set the height from which the refetching starts")
	cmd.Flags().Int64(flagEnd, 0, "Set the height to which the refetching finishes")

	return cmd
}
