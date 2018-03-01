package version

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	// VersionCmd prints out the current sdk version
	VersionCmd = &cobra.Command{
		Use:   "version",
		Short: "Print the app version",
		Run:   doVersionCmd,
	}
)

func doVersionCmd(cmd *cobra.Command, args []string) {
	v := Version
	if GitCommit != "" {
		v = v + " " + GitCommit
	}
	fmt.Println(v)
}
