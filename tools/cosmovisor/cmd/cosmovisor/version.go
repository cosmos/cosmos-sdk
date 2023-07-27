package main

import (
	"encoding/json"
	"fmt"
	"runtime/debug"
	"strings"

	"github.com/spf13/cobra"

	"cosmossdk.io/tools/cosmovisor"
)

func NewVersionCmd() *cobra.Command {
	versionCmd := &cobra.Command{
		Use:          "version",
		Short:        "Display cosmovisor and APP version.",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			noAppVersion, _ := cmd.Flags().GetBool(cosmovisor.FlagCosmovisorOnly)
			if val, err := cmd.Flags().GetString(cosmovisor.FlagOutput); val == "json" && err == nil {
				return printVersionJSON(cmd, args, noAppVersion)
			}

			return printVersion(cmd, args, noAppVersion)
		},
	}

	versionCmd.Flags().StringP(cosmovisor.FlagOutput, "o", "text", "Output format (text|json)")
	versionCmd.Flags().Bool(cosmovisor.FlagCosmovisorOnly, false, "Print cosmovisor version only")

	return versionCmd
}

func getVersion() string {
	version, ok := debug.ReadBuildInfo()
	if !ok {
		panic("failed to get cosmovisor version")
	}

	return strings.TrimSpace(version.Main.Version)
}

func printVersion(cmd *cobra.Command, args []string, noAppVersion bool) error {
	cmd.Printf("cosmovisor version: %s\n", getVersion())
	if noAppVersion {
		return nil
	}

	if err := run(append([]string{"version"}, args...)); err != nil {
		return fmt.Errorf("failed to run version command: %w", err)
	}

	return nil
}

func printVersionJSON(cmd *cobra.Command, args []string, noAppVersion bool) error {
	if noAppVersion {
		cmd.Printf(`{"cosmovisor_version":"%s"}`+"\n", getVersion())
		return nil
	}

	buf := new(strings.Builder)
	if err := run(
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

	cmd.Println(string(out))
	return nil
}
