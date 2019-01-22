package version

import (
	"fmt"
	"runtime"

	"github.com/spf13/cobra"
)

var (
	// VersionCmd prints out the current sdk version
	VersionCmd = &cobra.Command{
		Use:   "version",
		Short: "Print the app version",
		Run:   printVersion,
	}
)

// return version of CLI/node and commit hash
func GetVersion() string {
	return Version
}

// CMD
func printVersion(cmd *cobra.Command, args []string) {
	fmt.Println("cosmos-sdk", GetVersion())
	fmt.Println("Built with Go", runtime.Version())
	fmt.Println("vendor/ directory's hash:", VendorDirHash)
}
