package version

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

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
		verInfo := newVersionInfo()

		if !viper.GetBool(flagLong) {
			fmt.Println(verInfo.Version)
			return nil
		}

		if viper.GetString(cli.OutputFlag) != "json" {
			fmt.Println(verInfo)
			return nil
		}

		bz, err := json.Marshal(verInfo)
		if err != nil {
			return err
		}
		fmt.Println(string(bz))
		return nil
	},
}
