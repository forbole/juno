package migrate

import (
	"github.com/spf13/cobra"
	"github.com/forbole/juno/v2/cmd/parse"
)

// NewFixCmd returns the Cobra command allowing to fix some BDJuno bugs without having to re-sync the whole database
func MigrateCmd(parseConfig *parse.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:               "migrate",
		Short:             "Migrate to latest version",
		PersistentPreRunE: runPersistentPreRuns(parse.ReadConfig(parseConfig)),
	}

	cmd.AddCommand(
		MigrateConfigCmd(),
		MigrateTablesCmd(parseConfig),
		PrepareTablesCmd(parseConfig),
	)

	return cmd
}

func runPersistentPreRuns(preRun func(_ *cobra.Command, _ []string) error) func(_ *cobra.Command, _ []string) error {
	return func(cmd *cobra.Command, args []string) error {
		if root := cmd.Root(); root != nil {
			if root.PersistentPreRunE != nil {
				err := root.PersistentPreRunE(root, args)
				if err != nil {
					return err
				}
			}
		}

		return preRun(cmd, args)
	}
}
