package cmd

import (
	"fmt"
	"os"
	"regexp"

	"github.com/spf13/cobra"
)

// Cosmovisor version value. Set during build
var Version string

const (
	versionRegex = `v(?P<major>0|[1-9]\d*)\.(?P<minor>0|[1-9]\d*)\.(?P<patch>0|[1-9]\d*)`
	// v                       - Version string starts with `v`
	// (?P<major>0|[1-9]\d*)   - Match group `major` - Matches 0 once or digits 1-9 zero or many times
	// \.                      - Dot between `major` and `minor`
	// (?P<minor>0|[1-9]\d*)   - Match group `minor`- Matches 0 once or digits 1-9 zero or many times
	// \.                      - Dot between `minor` and `patch`
	// (?P<patch>0|[1-9]\d*)   - Match group `patch`- Matches 0 once or digits 1-9 zero or many times
)

func init() {
	rootCmd.AddCommand(versionCmd)
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Prints the version number of Cosmovisor and the application",
	Long:  `Prints the version number of Comovisor and the underlying application`,
	Run: func(cmd *cobra.Command, args []string) {

		versionFound, version := GetVersion(Version)

		if versionFound {
			fmt.Println("Cosmovisor Version: ", version)
		} else {
			fmt.Fprintf(os.Stderr, "Could not parse the Cosmovisor version from the binary. \nVersion value found in binary: %s \n", Version)
		}

	},
}

// GetVersion does regex matching against the version value set in the cosmovisor binary
// Returns if valid version was found, and if valid  version is found, returns the version string
// Returns false and empty string when valid version is not found
func GetVersion(versionStr string) (bool, string) {
	re := regexp.MustCompile(versionRegex) // Requires major, minor and patch versions in the form v1.0.0

	for _, version_match := range re.FindAllString(versionStr, -1) {
		return true, version_match
	}

	return false, ""
}
