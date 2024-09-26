package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func NewShowUpgradeInfoCmd() *cobra.Command {
	showUpgradeInfo := &cobra.Command{
		Use:          "show-upgrade-info",
		Short:        "Show upgrade-info.json into stdout.",
		SilenceUsage: true,
		RunE:         showUpgradeInfoCmd,
	}

	return showUpgradeInfo
}

func showUpgradeInfoCmd(cmd *cobra.Command) error {
	data, err := os.ReadFile(cfg.UpgradeInfoFilePath())
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("upgrade-info.json not found at %s: %w", args[0], err)
		}
		return fmt.Errorf("failed to read upgrade-info.json: %w", err)
	}

	_, err = fmt.Fprintln(cmd.OutOrStdout(), string(data))
	if err != nil {
		return fmt.Errorf("failed to write output: %w", err)
	}

	return nil
}
