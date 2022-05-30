package main

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/cosmos/cosmos-sdk/cosmovisor"
	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
	tmcli "github.com/tendermint/tendermint/libs/cli"
)

func init() {
	versionCmd.Flags().StringP(OutputFlag, "o", "text", "Output format (text|json)")
	rootCmd.AddCommand(versionCmd)
}

var (
	// Version represents Cosmovisor version value. Overwritten during build
	Version = "1.1.0"
	// OutputFlag defines the output format flag
	OutputFlag = tmcli.OutputFlag
)

var versionCmd = &cobra.Command{
	Use:          "version",
	Short:        "Prints the version of Cosmovisor.",
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		logger := cmd.Context().Value(cosmovisor.LoggerKey).(*zerolog.Logger)

		if val, err := cmd.Flags().GetString(OutputFlag); val == "json" && err == nil {
			return printVersionJSON(logger, args)
		}

		return printVersion(logger, args)
	},
}

func printVersion(logger *zerolog.Logger, args []string) error {
	fmt.Println("cosmovisor version: ", Version)

	if err := Run(logger, append([]string{"version"}, args...)); err != nil {
		return fmt.Errorf("failed to run version command: %w", err)
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
		return fmt.Errorf("failed to run version command: %w", err)
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
		return fmt.Errorf("can't print version output, expected valid json from APP, got: %s - %w", buf.String(), err)
	}

	fmt.Println(string(out))
	return nil
}
