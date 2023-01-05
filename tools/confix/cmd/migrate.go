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
	FlagStdOut       bool
	FlagVerbose      bool
	FlagSkipValidate bool
)

func MigrateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "migrate [target-version] [app-toml-path] (options)",
		Short: "Migrate Cosmos SDK app configuration file to the specified version",
		Long: `Migrate the contents of the Cosmos SDK app configuration (app.toml) to the specified version.
The output is written in-place unless --stdout is provided.
In case of any error in updating the file, no output is written.`,
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			var filename string
			targetVersion := args[0]

			clientCtx := client.GetClientContextFromCmd(cmd)
			if clientCtx.HomeDir != "" {
				filename = fmt.Sprintf("%s/config/app.toml", clientCtx.HomeDir)
			} else if len(args) > 1 {
				filename = args[1]
			} else {
				return fmt.Errorf("must provide a path to the app.toml file")
			}

			plan, ok := confix.Migrations[targetVersion]
			if !ok {
				return fmt.Errorf("unknown version %q, supported versions are: %q", targetVersion, maps.Keys(confix.Migrations))
			}

			ctx := context.Background()
			if FlagVerbose {
				ctx = confix.WithLogWriter(ctx, os.Stderr)
			}

			outputPath := filename
			if FlagStdOut {
				outputPath = ""
			}

			if err := confix.Upgrade(ctx, plan(filename, targetVersion), filename, outputPath, FlagSkipValidate); err != nil {
				log.Fatalf("Failed to migrate config: %v", err)
			}

			return nil
		},
	}

	cmd.Flags().BoolVar(&FlagStdOut, "stdout", false, "print the updated config to stdout")
	cmd.Flags().BoolVar(&FlagVerbose, "verbose", false, "log changes to stderr")
	cmd.Flags().BoolVar(&FlagSkipValidate, "skip-validate", false, "skip configuration validation (allows to mutate unknown configurations)")

	return cmd
}
