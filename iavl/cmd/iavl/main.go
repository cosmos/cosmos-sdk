package main

import (
	"os"

	"github.com/spf13/cobra"
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "iavl",
		Short: "IAVL tree inspection and management tool",
	}

	rootCmd.AddCommand(
		newViewCmd(),
		newImportCmd(),
	)

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
