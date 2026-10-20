package server

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/cosmos/cosmos-sdk/client/flags"
	serverconfig "github.com/cosmos/cosmos-sdk/server/config"
)

const flagAppConfig = "config"

// validateAppConfigRunE resolves the candidate path and validates it without
// creating anything.
//
// `--home` is resolved directly from the flag rather than through the standard
// client/server context: building that context is what runs `interceptConfigs`,
// which writes a default `app.toml` into the node home. The whole point of this
// command is that it does not do that.
func validateAppConfigRunE(cmd *cobra.Command, _ []string) error {
	configPath, err := cmd.Flags().GetString(flagAppConfig)
	if err != nil {
		return err
	}

	if configPath == "" {
		home, err := cmd.Flags().GetString(flags.FlagHome)
		if err != nil {
			return err
		}
		if home == "" {
			return fmt.Errorf("cannot resolve the node home: --home is unset")
		}
		configPath = filepath.Join(home, "config", "app.toml")
	}

	// Refuse a missing file rather than letting viper fall back to defaults,
	// which would validate a default config and report success for a file that
	// is not there.
	if _, err := os.Stat(configPath); err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return fmt.Errorf("config file not found: %s", configPath)
		}
		return fmt.Errorf("cannot read config file %s: %w", configPath, err)
	}

	v := viper.New()
	// Pin the format instead of letting viper infer it from the extension. The
	// startup loader reads app.toml as TOML explicitly, so a staged candidate
	// such as `app.toml.new` (or an extensionless temp file) must be checked by
	// the same rules; inference would either fail on the name or parse a JSON
	// candidate that startup would reject.
	v.SetConfigType("toml")
	v.SetConfigFile(configPath)
	if err := v.ReadInConfig(); err != nil {
		return fmt.Errorf("cannot parse config file %s: %w", configPath, err)
	}

	conf, err := serverconfig.GetConfig(v)
	if err != nil {
		return err
	}

	if err := conf.ValidateBasic(); err != nil {
		return fmt.Errorf("invalid app.toml at %s: %w", configPath, err)
	}

	cmd.Println("OK:", configPath)
	return nil
}

// ValidateAppConfigCmd returns a read-only command that parses and validates an
// `app.toml` without starting or initializing a node.
//
// It never writes. The node's own startup path (`interceptConfigs`) creates a
// default `app.toml` when the file is missing, which makes it unusable for
// checking a candidate configuration: validating a file that is not there would
// silently produce one and report success. This command instead fails on a
// missing file and touches nothing, so CI or deploy automation can validate a
// candidate and promote it only if the check passes.
//
// It replaces the inherited `PersistentPreRunE` with a no-op. That hook is what
// calls `interceptConfigs` for every child of the root command, and letting it
// run here would create `<home>/config/app.toml` before this image reads the
// candidate -- the command would then validate a file it had just written and
// report OK for one that never existed.
func ValidateAppConfigCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "validate-app",
		Short: "Validate app.toml without starting the node",
		Long: `Parse and validate app.toml, reporting the first problem found.

The command is read-only: it never creates a configuration file. If the file is
missing it fails. Without --config, <home>/config/app.toml is validated.`,
		Args: cobra.NoArgs,
		// Override the root's mutating setup hook: see the doc comment above.
		PersistentPreRunE: func(_ *cobra.Command, _ []string) error { return nil },
		RunE:              validateAppConfigRunE,
		SilenceUsage:      true,
	}

	cmd.Flags().String(flagAppConfig, "", "path to app.toml (defaults to <home>/config/app.toml)")

	return cmd
}
