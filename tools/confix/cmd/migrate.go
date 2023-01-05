package cmd

import (
	"context"
	"fmt"
	"log"
	"os"

	"cosmossdk.io/tools/confix"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/spf13/cobra"
	"golang.org/x/exp/maps"
)

var (
	FlagStdOut  bool
	FlagVerbose bool
)

func MigrateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "migrate [target-version] [config-file] (options)",
		Short: "Migrate Cosmos SDK configuration files",
		Long: `Modify the contents of the specified config TOML file to update the names,
locations, and values of configuration settings to the current configuration
layout. The output is written in-place unless --stdout is provided.
In case of any error in updating the file, no output is written.`,
		Args: cobra.MinimumNArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			targetVersion, filename := args[0], args[1]

			clientCtx := client.GetClientContextFromCmd(cmd)
			if clientCtx.HomeDir != "" {
				filename = fmt.Sprintf("%s/config/%s.toml", clientCtx.HomeDir, filename)
			}

			plan, ok := confix.Versions[targetVersion]
			if !ok {
				return fmt.Errorf("unknown version %q, supported versions are: %q", targetVersion, maps.Keys(confix.Versions))
			}

			ctx := context.Background()
			if FlagVerbose {
				ctx = confix.WithLogWriter(ctx, os.Stderr)
			}

			outputPath := filename
			if FlagStdOut {
				outputPath = ""
			}

			if err := confix.Upgrade(ctx, plan, filename, outputPath); err != nil {
				log.Fatalf("Failed to migrate config: %v", err)
			}

			return nil
		},
	}

	cmd.Flags().BoolVar(&FlagStdOut, "stdout", false, "print the updated config to stdout")
	cmd.Flags().BoolVar(&FlagVerbose, "verbose", false, "log changes to stderr")

	return cmd
}
