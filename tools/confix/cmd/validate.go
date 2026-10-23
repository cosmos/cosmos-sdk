package cmd

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/server/config"
)

const flagConfig = "config"

// ValidateCommand creates a read-only command that validates app.toml using
// SDK defaults + Config.ValidateBasic(), without starting the node or writing files.
func ValidateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "validate [app]",
		Short: "Validate app.toml without starting the node",
		Long: `Validate an application configuration file (app.toml) in a read-only way.

The command starts from SDK defaults, parses the selected file, and runs
Config.ValidateBasic(). It never creates, rewrites, or starts the node.

If --config is set, that path is used. Otherwise:
  - with an optional "app" argument (or no args) and a client home, validates
    <home>/config/app.toml
  - with an explicit path argument, validates that file (relative paths use cwd)

A missing file is an error.`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			configPath, err := resolveAppConfigPath(cmd, args)
			if err != nil {
				return err
			}

			if _, err := os.Stat(configPath); err != nil {
				if os.IsNotExist(err) {
					return fmt.Errorf("app.toml not found: %s", configPath)
				}
				return err
			}

			v := viper.New()
			v.SetConfigType("toml")
			v.SetConfigFile(configPath)
			if err := v.ReadInConfig(); err != nil {
				return fmt.Errorf("failed to read %s: %w", configPath, err)
			}

			cfg, err := config.GetConfig(v)
			if err != nil {
				return fmt.Errorf("failed to parse %s: %w", configPath, err)
			}

			if err := cfg.ValidateBasic(); err != nil {
				return fmt.Errorf("invalid app.toml (%s): %w", configPath, err)
			}

			return client.GetClientContextFromCmd(cmd).PrintString(
				fmt.Sprintf("app.toml is valid: %s\n", configPath),
			)
		},
	}

	cmd.Flags().String(flagConfig, "", "path to app.toml (overrides home/args)")

	return cmd
}

func resolveAppConfigPath(cmd *cobra.Command, args []string) (string, error) {
	if p, _ := cmd.Flags().GetString(flagConfig); p != "" {
		return p, nil
	}

	clientCtx := client.GetClientContextFromCmd(cmd)

	if len(args) == 0 {
		if clientCtx.HomeDir == "" {
			return "", errors.New("must provide a path to app.toml (or set --home / --config)")
		}
		return filepath.Join(clientCtx.HomeDir, "config", "app.toml"), nil
	}

	arg := args[0]
	// "app" means <home>/config/app.toml when home is set (matches view/get style)
	if arg == "app" || arg == "app.toml" {
		if clientCtx.HomeDir != "" {
			return filepath.Join(clientCtx.HomeDir, "config", AppConfigName(arg)), nil
		}
		if arg == "app" {
			return "", errors.New("must provide a path to app.toml when --home is unset")
		}
	}

	if strings.HasSuffix(arg, "client.toml") {
		return "", errors.New("app.toml file expected, got client.toml")
	}

	return arg, nil
}

// AppConfigName normalizes the app config basename.
func AppConfigName(arg string) string {
	if strings.HasSuffix(arg, tomlSuffix) {
		return arg
	}
	return arg + tomlSuffix
}
