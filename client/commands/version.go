package commands

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/version"
)

// CommitHash should be filled by linker flags
var CommitHash = ""

// VersionCmd - command to show the application version
var VersionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show version info",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("%s-%s\n", version.Version, CommitHash)
	},
}
