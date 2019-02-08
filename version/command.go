package version

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/tendermint/tendermint/libs/cli"
)

const (
	flagLong = "long"
)

var (

	// VersionCmd prints out the current sdk version
	VersionCmd = &cobra.Command{
		Use:   "version",
		Short: "Print the app version",
		RunE: func(_ *cobra.Command, _ []string) error {
			verInfo := newVersionInfo()

			if !viper.GetBool(flagLong) {
				fmt.Println(verInfo.CosmosSDK)
				return nil
			}

			if viper.GetString(cli.OutputFlag) != "json" {
				fmt.Print(verInfo)
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
)

func init() {
	VersionCmd.Flags().Bool(flagLong, false, "Print long version information")
}
