package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/cosmos/cosmos-sdk/cosmovisor"
	cverrors "github.com/cosmos/cosmos-sdk/cosmovisor/errors"
	"github.com/cosmos/cosmos-sdk/cosmovisor/logging"
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
func (cv *Cosmovisor) PrintVersion(args []string) error {
	for _, arg := range args {
		if strings.Contains(arg, FlagJSON) {
			return cv.printVersionJSON(args)
		}
	}

	return cv.printVersion(args)
}

func (cv *Cosmovisor) printVersion(args []string) error {
	fmt.Println("Cosmovisor Version: ", Version)

	if err := cv.Run(append([]string{"version"}, args...)); err != nil {
		cv.handleRunVersionFailure(err)
	}

	return nil
}

func (cv *Cosmovisor) printVersionJSON(args []string) error {
	buf := new(strings.Builder)

	if err := cv.Run(
		[]string{"version", "--long", "--output", "json"},
		StdOutRunOption(buf),
		DisableLogging(cv.logger),
	); err != nil {
		cv.handleRunVersionFailure(err)
	}

	out, err := json.Marshal(struct {
		Version    string          `json:"cosmovisor_version"`
		AppVersion json.RawMessage `json:"app_version"`
	}{
		Version:    Version,
		AppVersion: json.RawMessage(buf.String()),
	})
	if err != nil {
		cv.logger.EnableLogger()
		return fmt.Errorf("Can't print version output, expected valid json from APP, got: %s - %w", buf.String(), err)
	}

	fmt.Println(string(out))
	return nil
}

func (cv *Cosmovisor) handleRunVersionFailure(err error) {
	// Check the config and output details or any errors.
	// Not using the cosmovisor.Logger in order to ignore any level it might have set,
	// and also to not have any of the extra parameters in the output.
	cv.logger.EnableLogger()
	output := zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.Kitchen}
	logger := logging.Logger{Logger: zerolog.New(output).With().Timestamp().Logger()}
	cverrors.LogErrors(logger, fmt.Sprintf("Can't run %s version", strings.ToUpper(os.Getenv(cosmovisor.EnvName))), err)
}
