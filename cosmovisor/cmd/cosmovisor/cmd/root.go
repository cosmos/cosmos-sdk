package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
)

var logger *zerolog.Logger

var rootCmd = &cobra.Command{
	Use:   "cosmovisor",
	Short: "Facilitate your cosmos-sdk upgrades with Cosmovisor",
	Long:  "Cosmovisor is a small process manager for Cosmos SDK application binaries that monitors the governance module for incoming chain upgrade proposals.",
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute(log *zerolog.Logger) {
	logger = log

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

// isOneOf returns true if the given arg equals one of the provided options (ignoring case).
func isOneOf(arg string, options []string) bool {
	for _, opt := range options {
		if strings.EqualFold(arg, opt) {
			return true
		}
	}
	return false
}
