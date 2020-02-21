package version

import (
	"encoding/json"
	"fmt"
	"runtime"

	"github.com/angelorc/desmos-parser/config"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v2"
)

var (
	// Version defines the application version (defined at compile time)
	Version = ""

	// GitCommit defines the application commit hash (defined at compile time)
	Commit = ""
)

// GetVersionCmd returns the command that allows to show the version information
func GetVersionCmd() *cobra.Command {
	versionCmd := &cobra.Command{
		Use:   "version",
		Short: "Print the version information",
		RunE: func(cmd *cobra.Command, args []string) error {

			verInfo := struct {
				Version string `json:"version" yaml:"version"`
				Commit  string `json:"commit" yaml:"commit"`
				Go      string `json:"go" yaml:"go"`
			}{
				Version: Version,
				Commit:  Commit,
				Go:      fmt.Sprintf("%s %s/%s", runtime.Version(), runtime.GOOS, runtime.GOARCH),
			}

			var (
				bz  []byte
				err error
			)

			versionFormat := viper.GetString(config.FlagFormat)
			switch versionFormat {
			case "json":
				bz, err = json.Marshal(verInfo)

			default:
				bz, err = yaml.Marshal(&verInfo)
			}

			if err != nil {
				return err
			}

			_, err = fmt.Println(string(bz))
			return err
		},
	}

	versionCmd.Flags().String(config.FlagFormat, "text", "Print the version in the given format (text | json)")

	return versionCmd
}
