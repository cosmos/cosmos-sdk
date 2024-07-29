package main

import (
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
		cfg, err := cosmovisor.GetConfigFromFile(cmd.Flag(cosmovisor.FlagCosmovisorConfig).Value.String())
		if err != nil {
			return err
		}

		cmd.Print(cfg.DetailString())
		return nil
	},
}
