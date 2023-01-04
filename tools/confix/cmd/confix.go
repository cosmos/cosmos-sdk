package cmd

import (
	"context"
	"errors"
	"log"
	"os"

	"cosmossdk.io/tools/confix"
	"github.com/spf13/cobra"
)

var (
	configPath, outputPath string
	doVerbose, forceNext   bool
)

func ConfixCommand() *cobra.Command {
	confixCmd := &cobra.Command{
		Use:   "confix --config <src> [--output <dst>]",
		Short: "Update Cosmos SDK configuration files",
		Long: `Modify the contents of the specified --config TOML file to update the names,
locations, and values of configuration settings to the current configuration
layout. The output is written to --output, or to stdout.
It is valid to set --config and --output to the same path. In that case, the file will
be modified in-place. In case of any error in updating the file, no output is written.`,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			// TODO support directly the home dir and detect the config
			if configPath == "" {
				return errors.New("a non-empty --config path must be specified")
			}

			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			if doVerbose {
				ctx = confix.WithLogWriter(ctx, os.Stderr)
			}
			if err := confix.Upgrade(ctx, configPath, outputPath, forceNext); err != nil {
				log.Fatalf("Upgrading config: %v", err)
			}

			return nil
		},
	}

	confixCmd.Flags().StringVar(&configPath, "config", "", "Config file path (required)")
	confixCmd.Flags().StringVar(&outputPath, "output", "", "output file path (default stdout)")
	confixCmd.Flags().BoolVar(&doVerbose, "verbose", false, "log changes to stderr")
	confixCmd.Flags().BoolVar(&forceNext, "force", false, "disable version check, assumes latest version (dangerous!)")

	return confixCmd
}
