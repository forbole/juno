package genesis

import (
	"github.com/spf13/cobra"

	parsecmdtypes "github.com/forbole/juno/v5/cmd/parse/types"

	"github.com/forbole/juno/v5/modules/cosmos"
	nodeconfig "github.com/forbole/juno/v5/node/config"
)

const (
	flagPath = "genesis-file-path"
)

// NewGenesisCmd returns the Cobra command allowing to parse the genesis file
func NewGenesisCmd(parseConfig *parsecmdtypes.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "genesis-file",
		Short: "Parse the genesis file",
		Long: `
Parse the genesis file only. 
Note that the modules built will NOT have access to the node as they are only supposed to deal with the genesis
file itself and not the on-chain data.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Read the configuration
			cfg, err := parsecmdtypes.ReadConfig(parseConfig)
			if err != nil {
				return err
			}

			// Set the node to be of type None so that the node won't be built
			cfg.Node.Type = nodeconfig.TypeNone

			// Build the parsing infrastructures
			infrastructures, err := parsecmdtypes.GetInfrastructures(cfg, parseConfig)
			if err != nil {
				return err
			}

			// Get the file path
			genesisFilePath := cfg.Parser.GenesisFilePath
			customPath, _ := cmd.Flags().GetString(flagPath)
			if customPath != "" {
				genesisFilePath = customPath
			}

			// Read the genesis file
			genDoc, err := cosmos.ReadGenesisFileGenesisDoc(genesisFilePath)
			if err != nil {
				return err
			}

			genState, err := cosmos.GetGenesisState(genDoc)
			if err != nil {
				return err
			}

			for _, module := range infrastructures.Modules {
				if module, ok := module.(cosmos.GenesisModule); ok {
					err = module.HandleGenesis(genDoc, genState)
					if err != nil {
						return err
					}
				}
			}

			return nil
		},
	}

	cmd.Flags().String(flagPath, "", "Path to the genesis file to be used. If empty, the path will be taken from the config file")

	return cmd
}
