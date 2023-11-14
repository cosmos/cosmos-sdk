package cmd

import (
	"errors"
	"fmt"
	"strings"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/spf13/cobra"
	"golang.org/x/exp/maps"

	"cosmossdk.io/tools/confix"
)

// DiffCommand creates a new command for comparing configuration files
func DiffCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "diff [target-version] <config-path> [config-type]",
		Short: "Outputs all config values that are different from the default.",
		Long:  "This command compares the specified configuration file (app.toml or client.toml) with the defaults and outputs any differences.",
		Args:  cobra.MinimumNArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			targetVersion := args[0]
			configPath := args[1]
			configType := confix.AppConfigType // Default to app configuration
			clientCtx := client.GetClientContextFromCmd(cmd)

			if len(args) > 2 {
				configType = strings.ToLower(args[2])
				if configType != confix.AppConfigType && configType != confix.ClientConfigType {
					return errors.New("config type must be 'app' or 'client'")
				}
			}

			if _, ok := confix.Migrations[targetVersion]; !ok {
				return fmt.Errorf("unknown version %q, supported versions are: %q", targetVersion, maps.Keys(confix.Migrations))
			}

			targetVersionFile, err := confix.LoadLocalConfig(targetVersion, configType)
			if err != nil {
				panic(fmt.Errorf("failed to load internal config: %w", err))
			}

			rawFile, err := confix.LoadConfig(configPath)
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}

			diff := confix.DiffValues(rawFile, targetVersionFile)
			if len(diff) == 0 {
				return clientCtx.PrintString("All config values are the same as the defaults.\n")
			}

			if err := clientCtx.PrintString("The following config values are different from the defaults:\n"); err != nil {
				return err
			}

			confix.PrintDiff(cmd.OutOrStdout(), diff)
			return nil
		},
	}
}
