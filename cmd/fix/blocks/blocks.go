package blocks

import (
	"fmt"

	"github.com/forbole/juno/v2/cmd/parse"
	"github.com/forbole/juno/v2/types/config"

	"github.com/spf13/cobra"

	"github.com/forbole/juno/v2/parser"
)

const (
	flagForce       = "force"
	flagStartHeight = "start"
	flagEndHeight   = "end"
)

// blocksCmd returns a Cobra command that allows to fix missing blocks in database
func blocksCmd(parseConfig *parse.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "all",
		Short: "Fix missing blocks and transactions in database from the start height",
		RunE: func(cmd *cobra.Command, args []string) error {

			parseCtx, err := parse.GetParsingContext(parseConfig)
			if err != nil {
				return err
			}

			workerCtx := parser.NewContext(parseCtx.EncodingConfig.Marshaler, nil, parseCtx.Node, parseCtx.Database, parseCtx.Logger, parseCtx.Modules)
			worker := parser.NewWorker(0, workerCtx)

			// Get latest height
			height, err := parseCtx.Node.LatestHeight()
			if err != nil {
				return fmt.Errorf("error while getting chain latest block height: %s", err)
			}

			force, _ := cmd.Flags().GetBool(flagForce)

			startHeight := config.Cfg.Parser.StartHeight
			startFlag, _ := cmd.Flags().GetInt64(flagStartHeight)
			if startFlag > 0 {
				// Get start flag height if set
				startHeight = startFlag
			}

			endFlag, _ := cmd.Flags().GetInt64(flagEndHeight)
			if endFlag > 0 {
				// Get end flag height if set
				height = endFlag
			}

			k := startHeight

			fmt.Printf("Refetching missing blocks and transactions from height %d to %d \n", k, height)
			for ; k <= height; k++ {
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

	cmd.Flags().Bool(flagForce, false, "If set, forces the fetch of blocks by overwriting any existing one")
	cmd.Flags().Int64(flagStartHeight, 0, "If set, parses from the configured height")
	cmd.Flags().Int64(flagEndHeight, 0, "If set, parses up to the configured height")

	return cmd
}
