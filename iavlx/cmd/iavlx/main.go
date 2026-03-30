package main

import (
	"os"

	"github.com/spf13/cobra"
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "iavlx",
		Short: "iavlx tree inspection and management tool",
	}

	rootCmd.AddCommand(
		newViewCmd(),
		newImportCmd(),
		newRollbackCmd(),
	)

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
