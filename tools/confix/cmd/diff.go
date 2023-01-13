package cmd

import (
	"github.com/spf13/cobra"
)

func DiffCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "diff [config]",
		Short: "Outputs all config values that are different from the defaults.",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {

			return nil
		},
	}
}
