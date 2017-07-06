package commands

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/tendermint/basecoin/version"
)

// VersionCmd - command to show the application version
var VersionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show version info",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(version.Version)
	},
}

// QuickVersionCmd - returns a version command based on version input
func QuickVersionCmd(version string) *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Show version info",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println(version)
		},
	}
}
