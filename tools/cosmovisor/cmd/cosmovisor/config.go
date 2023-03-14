package main

import (
	"fmt"

	"cosmossdk.io/tools/cosmovisor"
	"github.com/spf13/cobra"
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

		fmt.Fprint(cmd.OutOrStdout(), cfg.DetailString())
		return nil
	},
}
