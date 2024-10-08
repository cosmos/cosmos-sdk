package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"cosmossdk.io/tools/cosmovisor"
)

func NewShowUpgradeInfoCmd() *cobra.Command {
	return &cobra.Command{
		Use:          "show-upgrade-info",
		Short:        "Display current upgrade-info.json from <app> data directory",
		SilenceUsage: false,
		Args:         cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
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
					cmd.Printf("No upgrade info found at %s\n", cfg.UpgradeInfoFilePath())
					return nil
				}
				return fmt.Errorf("failed to read upgrade-info.json: %w", err)
			}

			cmd.Println(string(data))
			return nil
		},
	}
}
