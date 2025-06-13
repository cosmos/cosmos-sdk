package main

import (
	"fmt"

	"github.com/spf13/cobra"

	"cosmossdk.io/tools/cosmovisor"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Display cosmovisor config.",
	Long: `Display cosmovisor config. If a config file is provided, it will display the config from the file, 
otherwise it will display the config from the environment variables.`,
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := getConfigFromCmd(cmd)
		if err != nil {
			return err
		}

		cmd.Print(cfg.DetailString())
		return nil
	},
}

// getConfigFromCmd retrieves the cosmovisor configuration from the command flags.
func getConfigFromCmd(cmd *cobra.Command) (*cosmovisor.Config, error) {
	configPath, err := cmd.Flags().GetString(cosmovisor.FlagCosmovisorConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to get config flag: %w", err)
	}

	cfg, err := cosmovisor.GetConfigFromFile(configPath)
	if err != nil {
		return nil, err
	}
	return cfg, nil
}
