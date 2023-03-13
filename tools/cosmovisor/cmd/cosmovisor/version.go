package main

import (
	"encoding/json"
	"fmt"
	"runtime/debug"
	"strings"

	"github.com/spf13/cobra"
)

// OutputFlag defines the output format flag
var OutputFlag = "output"

func NewVersionCmd() *cobra.Command {
	versionCmd := &cobra.Command{
		Use:          "version",
		Short:        "Display cosmovisor and APP version.",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if val, err := cmd.Flags().GetString(OutputFlag); val == "json" && err == nil {
				return printVersionJSON(cmd, args)
			}

			return printVersion(cmd, args)
		},
	}

	versionCmd.Flags().StringP(OutputFlag, "o", "text", "Output format (text|json)")

	return versionCmd
}

func getVersion() string {
	version, ok := debug.ReadBuildInfo()
	if !ok {
		panic("failed to get cosmovisor version")
	}

	return strings.TrimSpace(version.Main.Version)
}

func printVersion(cmd *cobra.Command, args []string) error {
	fmt.Printf("cosmovisor version: %s\n", getVersion())

	if err := Run(cmd, append([]string{"version"}, args...)); err != nil {
		return fmt.Errorf("failed to run version command: %w", err)
	}

	return nil
}

func printVersionJSON(cmd *cobra.Command, args []string) error {
	buf := new(strings.Builder)

	if err := Run(
		cmd,
		[]string{"version", "--long", "--output", "json"},
		StdOutRunOption(buf),
	); err != nil {
		return fmt.Errorf("failed to run version command: %w", err)
	}

	out, err := json.Marshal(struct {
		Version    string          `json:"cosmovisor_version"`
		AppVersion json.RawMessage `json:"app_version"`
	}{
		Version:    getVersion(),
		AppVersion: json.RawMessage(buf.String()),
	})
	if err != nil {
		return fmt.Errorf("can't print version output, expected valid json from APP, got: %s - %w", buf.String(), err)
	}

	fmt.Println(string(out))
	return nil
}
