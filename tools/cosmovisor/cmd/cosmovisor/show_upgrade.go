package main

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
)

func NewShowManualUpgradesCmd() *cobra.Command {
	return &cobra.Command{
		Use:          "show-manual-upgrades",
		Short:        "Display planned manual upgrades",
		SilenceUsage: false,
		Args:         cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := getConfigFromCmd(cmd)
			if err != nil {
				return err
			}

			data, err := cfg.ReadManualUpgrades()
			if err != nil {
				return fmt.Errorf("failed to read upgrade-info.json.batch: %w", err)
			}

			bz, err := json.MarshalIndent(data, "", "  ")
			if err != nil {
				return fmt.Errorf("failed to marshal manual upgrade info as json: %w", err)
			}

			cmd.Println(string(bz))
			return nil
		},
	}
}
