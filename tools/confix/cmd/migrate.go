package cmd

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/spf13/cobra"
	"golang.org/x/exp/maps"

	"cosmossdk.io/tools/confix"
)


var (
	FlagStdOut       bool
	FlagVerbose      bool
	FlagSkipValidate bool
)

func MigrateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "migrate [target-version] <config-path> [config-type]",
		Short: "Migrate Cosmos SDK configuration file to the specified version",
		Long: `Migrate the contents of the Cosmos SDK configuration (app.toml or client.toml) to the specified version. Configuration type is app by default.
The output is written in-place unless --stdout is provided.
In case of any error in updating the file, no output is written.`,
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			var configPath string
			
			clientCtx := client.GetClientContextFromCmd(cmd)
			targetVersion := args[0]
			configType := confix.AppConfigType // Default to app configuration

			if len(args) > 2 {
				configType = strings.ToLower(args[2])
				if configType != confix.AppConfigType && configType != confix.ClientConfigType {
					return errors.New("config type must be 'app' or 'client'")
				}
			}

			switch  {
			case len(args) > 1:
				configPath = args[1]
			case clientCtx.HomeDir != "":
				configPath = fmt.Sprintf("%s/config/%s.toml",clientCtx.HomeDir, configType)
			default:
				return errors.New("must provide a path to the config file")	
			}


			plan, ok := confix.Migrations[targetVersion]
			if !ok {
				return fmt.Errorf("unknown version %q, supported versions are: %q", targetVersion, maps.Keys(confix.Migrations))
			}

			rawFile, err := confix.LoadConfig(configPath)
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}

			ctx := context.Background()
			if FlagVerbose {
				ctx = confix.WithLogWriter(ctx, cmd.ErrOrStderr())
			}

			outputPath := configPath
			if FlagStdOut {
				outputPath = ""
			}

			if err := confix.Upgrade(ctx, plan(rawFile, targetVersion, configType), configPath, outputPath, FlagSkipValidate); err != nil {
				return fmt.Errorf("failed to migrate config: %w", err)
			}

			return nil
		},
	}

	cmd.Flags().BoolVar(&FlagStdOut, "stdout", false, "print the updated config to stdout")
	cmd.Flags().BoolVar(&FlagVerbose, "verbose", false, "log changes to stderr")
	cmd.Flags().BoolVar(&FlagSkipValidate, "skip-validate", false, "skip configuration validation (allows to migrate unknown configurations)")

	return cmd
}
