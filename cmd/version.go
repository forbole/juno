package cmd

import (
	"encoding/json"
	"fmt"
	"runtime"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
)

var (
	// Version defines the application version (defined at compile time)
	Version = ""

	// Commit defines the application commit hash (defined at compile time)
	Commit = ""
)

const (
	flagFormat = "format"
)

// VersionCmd returns the command that allows to show the version information
func VersionCmd() *cobra.Command {
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

			versionFormat := viper.GetString(flagFormat)
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

	versionCmd.Flags().String(flagFormat, "text", "Print the version in the given format (text | json)")

	return versionCmd
}
