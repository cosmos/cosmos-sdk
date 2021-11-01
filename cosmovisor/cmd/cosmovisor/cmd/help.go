package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/rs/zerolog"

	"github.com/cosmos/cosmos-sdk/cosmovisor"
)

// HelpArgs are the strings that indicate a cosmovisor help command.
var HelpArgs = []string{"help", "--help", "-h"}

// ShouldGiveHelp checks the env and provided args to see if help is needed or being requested.
// Help is needed if either cosmovisor.EnvName and/or cosmovisor.EnvHome env vars aren't set.
// Help is requested if the first arg is "help", "--help", or "-h".
func ShouldGiveHelp(arg string) bool {
	return isOneOf(arg, HelpArgs) || len(os.Getenv(cosmovisor.EnvName)) == 0 || len(os.Getenv(cosmovisor.EnvHome)) == 0
}

// DoHelp outputs help text
func DoHelp() {
	// Not using the logger for this output because the header and footer look weird for help text.
	fmt.Println(GetHelpText())
	// Check the config and output details or any errors.
	// Not using the cosmovisor.Logger in order to ignore any level it might have set,
	// and also to not have any of the extra parameters in the output.
	output := zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.Kitchen}
	logger := zerolog.New(output).With().Timestamp().Logger()
	cfg, err := cosmovisor.GetConfigFromEnv()
	cosmovisor.LogConfigOrError(logger, cfg, err)
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

To get help for the configured binary:
  cosmovisor run help
`, cosmovisor.EnvName, cosmovisor.EnvHome)
}
