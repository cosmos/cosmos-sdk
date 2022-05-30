package main

import (
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "cosmovisor",
	Short: "A process manager for Cosmos SDK application binaries.",
	Long:  GetHelpText(),
}
