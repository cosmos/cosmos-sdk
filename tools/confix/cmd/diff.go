package cmd

import (
	"github.com/spf13/cobra"
)

func DiffCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "diff [config]",
		Short: "Display the diff between the current config and the SDK default config",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {

			return nil
		},
	}
}
