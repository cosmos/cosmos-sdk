package main

import (
	"github.com/spf13/cobra"

	"cosmossdk.io/tools/cosmovisor"
)

var configCmd = &cobra.Command{
	Use:          "config",
	Short:        "Display cosmovisor config (prints environment variables used by cosmovisor).",
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := cosmovisor.GetConfigFromEnv()
		if err != nil {
			return err
		}

		cmd.Print(cfg.DetailString())
		return nil
	},
}
