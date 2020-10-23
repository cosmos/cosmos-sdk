package server

import "github.com/spf13/cobra"

func RosettaServer() *cobra.Command {
	return &cobra.Command{
		Use: "rosetta",
		RunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
	}
}
