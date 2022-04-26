package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	cverrors "github.com/cosmos/cosmos-sdk/cosmovisor/errors"
	"github.com/rs/zerolog"
)

var (
	// FlagJSON formats the output in json
	FlagJSON = "--json"
	// Version represents Cosmovisor version value. Overwritten during build
	Version = "1.1.0"
	// VersionArgs is the strings that indicate a cosmovisor version command.
	VersionArgs = []string{"version", "--version"}
)

// IsVersionCommand checks if the given args indicate that the version is being requested.
func IsVersionCommand(arg string) bool {
	return isOneOf(arg, VersionArgs)
}

// PrintVersion prints the cosmovisor version.
func PrintVersion(logger *zerolog.Logger, args []string) error {
	for _, arg := range args {
		if strings.Contains(arg, FlagJSON) {
			return printVersionJSON(logger, args)
		}
	}

	return printVersion(logger, args)
}

func printVersion(logger *zerolog.Logger, args []string) error {
	fmt.Println("Cosmovisor Version: ", Version)

	if err := Run(logger, append([]string{"version"}, args...)); err != nil {
		handleRunVersionFailure(err)
	}

	return nil
}

func printVersionJSON(logger *zerolog.Logger, args []string) error {
	buf := new(strings.Builder)

	// disable logger
	l := logger.Level(zerolog.Disabled)
	logger = &l

	if err := Run(
		logger,
		[]string{"version", "--long", "--output", "json"},
		StdOutRunOption(buf),
	); err != nil {
		handleRunVersionFailure(err)
	}

	out, err := json.Marshal(struct {
		Version    string          `json:"cosmovisor_version"`
		AppVersion json.RawMessage `json:"app_version"`
	}{
		Version:    Version,
		AppVersion: json.RawMessage(buf.String()),
	})
	if err != nil {
		l := logger.Level(zerolog.TraceLevel)
		logger = &l
		return fmt.Errorf("Can't print version output, expected valid json from APP, got: %s - %w", buf.String(), err)
	}

	fmt.Println(string(out))
	return nil
}

func handleRunVersionFailure(err error) {
	// Check the config and output details or any errors.
	// Not using the cosmovisor.Logger in order to ignore any level it might have set,
	// and also to not have any of the extra parameters in the output.
	output := zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.Kitchen}
	logger := zerolog.New(output).With().Timestamp().Logger()
	cverrors.LogErrors(&logger, "Can't run APP version", err)
}
