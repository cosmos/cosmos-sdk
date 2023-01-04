package cmd

import (
	"context"
	"fmt"
	"log"
	"os"

	"cosmossdk.io/tools/confix"
	"github.com/spf13/cobra"
	"golang.org/x/exp/maps"
)

var (
	outputPath string
	doVerbose  bool
)

func MigrateCommand() *cobra.Command {
	fixCmd := &cobra.Command{
		Use:   "migrate [target-version] [config-file] (options)",
		Short: "Update Cosmos SDK configuration files",
		Long: `Modify the contents of the specified config TOML file to update the names,
locations, and values of configuration settings to the current configuration
layout. The output is written to --output if provided, or to stdout.
It is valid to set the config file and the output to the same path. In that case, the file will
be modified in-place. In case of any error in updating the file, no output is written.`,
		Args: cobra.MinimumNArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			if _, ok := confix.Versions[args[0]]; !ok {
				return fmt.Errorf("unknown version %q, supported versions are: %q", args[0], maps.Keys(confix.Versions))
			}

			ctx := context.Background()
			if doVerbose {
				ctx = confix.WithLogWriter(ctx, os.Stderr)
			}

			if err := confix.Migrate(ctx, args[0], args[1], outputPath); err != nil {
				log.Fatalf("Upgrading config: %v", err)
			}

			return nil
		},
	}

	fixCmd.Flags().StringVar(&outputPath, "output", "", "output file path (default stdout)")
	fixCmd.Flags().BoolVar(&doVerbose, "verbose", false, "log changes to stderr")

	return fixCmd
}
