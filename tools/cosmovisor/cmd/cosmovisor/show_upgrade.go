package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func NewShowUpgradeInfoCmd() *cobra.Command {
	showUpgradeInfo := &cobra.Command{
		Use:          "show-upgrade-info <path to executable>",
		Short:        "Show upgrade-info.json into stdout.",
		SilenceUsage: true,
		RunE:         showUpgradeInfoCmd,
	}

	return showUpgradeInfo
}

func showUpgradeInfoCmd(cmd *cobra.Command, path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read upgrade-info.json: %w", err)
	}

	fmt.Println(string(data))

	return nil
}
