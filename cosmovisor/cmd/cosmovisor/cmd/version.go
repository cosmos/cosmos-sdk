package cmd

import (
	"fmt"
	"os"
	"time"

	cverrors "github.com/cosmos/cosmos-sdk/cosmovisor/errors"
	"github.com/rs/zerolog"
)

// Version represents Cosmovisor version value. Set during build
var Version string

// VersionArgs is the strings that indicate a cosmovisor version command.
var VersionArgs = []string{"version", "--version"}

// IsVersionCommand checks if the given args indicate that the version is being requested.
func IsVersionCommand(arg string) bool {
	return isOneOf(arg, VersionArgs)
}

// PrintVersion prints the cosmovisor version.
func PrintVersion() error {
	fmt.Println("Cosmovisor Version: ", Version)

	if err := Run([]string{"version"}); err != nil {
		// Check the config and output details or any errors.
		// Not using the cosmovisor.Logger in order to ignore any level it might have set,
		// and also to not have any of the extra parameters in the output.
		output := zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.Kitchen}
		logger := zerolog.New(output).With().Timestamp().Logger()
		cverrors.LogErrors(logger, "Can't run APP version", err)
	}
	return nil
}
