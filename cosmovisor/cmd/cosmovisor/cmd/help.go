package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/cosmos/cosmos-sdk/cosmovisor"
)

// ShouldGiveHelp checks the env and provided args to see if help is needed or being requested.
// Help is needed if either cosmovisor.EnvName and/or cosmovisor.EnvHome env vars aren't set.
// Help is requested if any args are "help", "--help", or "-h".
func ShouldGiveHelp(args []string) bool {
	if len(os.Getenv(cosmovisor.EnvName)) == 0 || len(os.Getenv(cosmovisor.EnvHome)) == 0 {
		return true
	}
	if len(args) == 0 {
		return false
	}
	for _, arg := range args {
		if strings.EqualFold(arg, "help") || strings.EqualFold(arg, "--help") || strings.EqualFold(arg, "-h") {
			return true
		}
	}
	return false
}

// DoHelp outputs help text
func DoHelp() {
	// Not using the logger for this output because the header and footer look weird for help text.
	fmt.Println(GetHelpText())
}

// GetHelpText creates the help text multi-line string.
func GetHelpText() string {
	return fmt.Sprintf(`Cosmosvisor - A process manager for Cosmos SDK application binaries.

Cosmovisor is a wrapper for a Cosmos SDK based App (set using the required %s env variable).
It starts the App by passing all provided arguments and monitors the %s/data/upgrade-info.json
file to perform an update. The upgrade-info.json file is created by the App x/upgrade module
when the blockchain height reaches an approved upgrade proposal. The file includes data from
the proposal. Cosmovisor interprets that data to perform an update: switch a current binary
and restart the App.

Configuration of Cosmovisor is done through environment variables, which are
documented in: https://github.com/cosmos/cosmos-sdk/tree/master/cosmovisor/README.md
`, cosmovisor.EnvName, cosmovisor.EnvHome)
}
