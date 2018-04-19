package version

import (
	"fmt"
	"net/http"

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

func getVersion() string {
	v := Version
	if GitCommit != "" {
		v = v + "-" + GitCommit
	}
	return v
}

// CMD
func printVersion(cmd *cobra.Command, args []string) {
	v := getVersion()
	fmt.Println(v)
}

// version REST handler endpoint
func RequestHandler(w http.ResponseWriter, r *http.Request) {
	v := getVersion()
	w.Write([]byte(v))
}
