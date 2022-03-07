package migrate

import (
	"github.com/spf13/cobra"
	"github.com/forbole/juno/v2/cmd/parse"
)

// MigrateCmd returns the Cobra command allowing to migrate config and tables to v3 version
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
