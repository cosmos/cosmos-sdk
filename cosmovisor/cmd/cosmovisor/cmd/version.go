package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/cosmos/cosmos-sdk/cosmovisor"
	cverrors "github.com/cosmos/cosmos-sdk/cosmovisor/errors"
	"github.com/rs/zerolog"
)

var (
	// Version represents Cosmovisor version value. Overwritten during build
	Version = "1.1.0"
	// VersionArgs is the strings that indicate a cosmovisor version command.
	VersionArgs = []string{"version", "--version"}
	// VersionJSONOutput is a format of the version output.
	VersionJSONOutput = "json"
)

// IsVersionCommand checks if the given args indicate that the version is being requested.
func IsVersionCommand(arg string) bool {
	return isOneOf(arg, VersionArgs)
}

// PrintVersion prints the cosmovisor version.
func PrintVersion(args []string) error {
	for _, arg := range args {
		if strings.Contains(arg, VersionJSONOutput) {
			return printVersionJSON(args)
		}
	}

	return printVersion(args)
}

func printVersion(args []string) error {
	fmt.Println("Cosmovisor Version: ", Version)

	if err := Run(append([]string{"version"}, args...)); err != nil {
		handleFailureRunVersion(err)
	}

	return nil
}

func printVersionJSON(args []string) error {
	buf := new(strings.Builder)

	if err := Run(
		append([]string{"version"}, args...),
		cosmovisor.StdOut(buf),
		cosmovisor.DisableLogging(),
	); err != nil {
		handleFailureRunVersion(err)
	}

	out, err := json.Marshal(struct {
		Version    string          `json:"cosmovisor_version"`
		AppVersion json.RawMessage `json:"app_version"`
	}{
		Version:    Version,
		AppVersion: json.RawMessage(buf.String()),
	})
	if err != nil {
		return fmt.Errorf("Can't marshal version output: %v", err)
	}

	fmt.Println(string(out))
	return nil
}

func handleFailureRunVersion(err error) {
	// Check the config and output details or any errors.
	// Not using the cosmovisor.Logger in order to ignore any level it might have set,
	// and also to not have any of the extra parameters in the output.
	output := zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.Kitchen}
	logger := zerolog.New(output).With().Timestamp().Logger()
	cverrors.LogErrors(logger, fmt.Sprintf("Can't run %s version", strings.ToUpper(os.Getenv(cosmovisor.EnvName))), err)
}
