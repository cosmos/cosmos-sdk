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
		Args:         cobra.ExactArgs(1),
		RunE:         showUpgradeInfoCmd,
	}

	return showUpgradeInfo
}

func showUpgradeInfoCmd(cmd *cobra.Command, args []string) error {
	if len(args) != 1 {
		return fmt.Errorf("expected exactly one argument")
	}

	data, err := os.ReadFile(args[0])
	if err != nil {
		return fmt.Errorf("failed to read upgrade-info.json: %w", err)
	}

	fmt.Println(string(data))

	return nil
}
