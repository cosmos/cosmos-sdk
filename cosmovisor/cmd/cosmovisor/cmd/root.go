package cmd

import (
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use: "cosmovisor",
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		//application commands which arent cosmovisor commands throw error and we can ignore that
	}
}
