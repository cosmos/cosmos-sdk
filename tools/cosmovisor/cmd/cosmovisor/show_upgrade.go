package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"cosmossdk.io/tools/cosmovisor"
)

func NewShowUpgradeInfoCmd() *cobra.Command {
	showUpgradeInfo := &cobra.Command{
		Use:          "show-upgrade-info",
		Short:        "Show upgrade-info.json into stdout.",
		SilenceUsage: false,
		Args:         cobra.NoArgs,
		RunE:         showUpgradeInfoCmd,
	}

	return showUpgradeInfo
}

func showUpgradeInfoCmd(cmd *cobra.Command, args []string) error {
	configPath, err := cmd.Flags().GetString(cosmovisor.FlagCosmovisorConfig)
	if err != nil {
		return fmt.Errorf("failed to get config flag: %w", err)
	}

	cfg, err := cosmovisor.GetConfigFromFile(configPath)
	if err != nil {
		return err
	}

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
