package migrate

import (
	"fmt"

	"github.com/spf13/cobra"

	v2 "github.com/forbole/juno/v2/cmd/migrate/v2"
	v3 "github.com/forbole/juno/v2/cmd/migrate/v3"
	"github.com/forbole/juno/v2/cmd/parse"
)

type Migrator func(parseCfg *parse.Config) error

var (
	migrations = map[string]Migrator{
		"v2": v2.RunMigration,
		"v3": v3.RunMigration,
	}
)

// MigrateCmd returns the Cobra command allowing to migrate config and tables to v3 version
func MigrateCmd(appName string, parseConfig *parse.Config) *cobra.Command {
	return &cobra.Command{
		Use:   "migrate [to-version]",
		Short: "Perform the migrations from the current version to the specified one",
		Long: `Migrates all the necessary things (config file, database, etc) from the current version to the new one.
If you are upgrading from a very old version to the latest one, migrations must be performed in order 
(eg. to migrate from v1 to v3 you need to do v1 -> v2 and then v2 -> v3). 
`,
		Example: fmt.Sprintf("%s migrate v3", appName),
		RunE: func(cmd *cobra.Command, args []string) error {
			version := args[0]

			migrator, ok := migrations[version]
			if !ok {
				return fmt.Errorf("migration for version %s not found", version)
			}

			return migrator(parseConfig)
		},
	}
}
