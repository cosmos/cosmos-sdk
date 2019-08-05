package version

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v2"

	"github.com/tendermint/tendermint/libs/cli"
)

const flagLong = "long"

func init() {
	Cmd.Flags().Bool(flagLong, false, "Print long version information")
}

// Cmd prints out the application's version information passed via build flags.
var Cmd = &cobra.Command{
	Use:   "version",
	Short: "Print the app version",
	RunE: func(_ *cobra.Command, _ []string) error {
		verInfo := NewInfo()

		if !viper.GetBool(flagLong) {
			fmt.Println(verInfo.Version)
			return nil
		}

		var bz []byte
		var err error

		switch viper.GetString(cli.OutputFlag) {
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
