package serverv2

import (
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"

	"cosmossdk.io/core/server"
	"cosmossdk.io/log"
)

// CommandFactory is a factory help create server/v2 root commands.
// For example usage see simapp/v2/cmd/root_di.go
type CommandFactory struct {
	defaultHomeDir string
	envPrefix      string
	configWriter   ConfigWriter
	loggerFactory  func(server.ConfigMap, io.Writer) (log.Logger, error)

	logger log.Logger
	// TODO remove this field
	// this viper handle is kept because certain commands in server/v2 fetch a viper instance
	// from the command context in order to read the config.
	// After merging #22267 this is no longer required, and server.ConfigMap can be used instead.
	// See issue #22388
	vipr *viper.Viper
}

type CommandFactoryOption func(*CommandFactory) error

// NewCommandFactory creates a new CommandFactory with the given options.
func NewCommandFactory(opts ...CommandFactoryOption) (*CommandFactory, error) {
	f := &CommandFactory{}
	for _, opt := range opts {
		err := opt(f)
		if err != nil {
			return nil, err
		}
	}
	return f, nil
}

// WithEnvPrefix sets the environment variable prefix for the command factory.
func WithEnvPrefix(envPrefix string) CommandFactoryOption {
	return func(f *CommandFactory) error {
		f.envPrefix = envPrefix
		return nil
	}
}

// WithStdDefaultHomeDir sets the server's default home directory `user home directory`/`defaultHomeBasename`.
func WithStdDefaultHomeDir(defaultHomeBasename string) CommandFactoryOption {
	return func(f *CommandFactory) error {
		// get the home directory from the environment variable
		// to not clash with the $HOME system variable, when no prefix is set
		// we check the NODE_HOME environment variable
		homeDir, envHome := "", "HOME"
		if len(f.envPrefix) > 0 {
			homeDir = os.Getenv(f.envPrefix + "_" + envHome)
		} else {
			homeDir = os.Getenv("NODE_" + envHome)
		}
		if homeDir != "" {
			f.defaultHomeDir = filepath.Clean(homeDir)
			return nil
		}

		// get user home directory
		userHomeDir, err := os.UserHomeDir()
		if err != nil {
			return err
		}

		f.defaultHomeDir = filepath.Join(userHomeDir, defaultHomeBasename)
		return nil
	}
}

// WithDefaultHomeDir sets the server's default home directory.
func WithDefaultHomeDir(homedir string) CommandFactoryOption {
	return func(f *CommandFactory) error {
		f.defaultHomeDir = homedir
		return nil
	}
}

// WithConfigWriter sets the config writer for the command factory.
// If set the config writer will be used to write TOML config files during ParseCommand invocations.
func WithConfigWriter(configWriter ConfigWriter) CommandFactoryOption {
	return func(f *CommandFactory) error {
		f.configWriter = configWriter
		return nil
	}
}

// WithLoggerFactory sets the logger factory for the command factory.
func WithLoggerFactory(loggerFactory func(server.ConfigMap, io.Writer) (log.Logger, error)) CommandFactoryOption {
	return func(f *CommandFactory) error {
		f.loggerFactory = loggerFactory
		return nil
	}
}

// enhanceCommand adds the following flags to the command:
//
// --log-level: The logging level (trace|debug|info|warn|error|fatal|panic|disabled or '*:<level>,<key>:<level>')
// --log-format: The logging format (json|plain)
// --log-no-color: Disable colored logs
// --home: directory for config and data
//
// It also sets the environment variable prefix for the viper instance.
func (f *CommandFactory) enhanceCommand(cmd *cobra.Command) {
	pflags := cmd.PersistentFlags()
	pflags.String(FlagLogLevel, "info", "The logging level (trace|debug|info|warn|error|fatal|panic|disabled or '*:<level>,<key>:<level>')")
	pflags.String(FlagLogFormat, "plain", "The logging format (json|plain)")
	pflags.Bool(FlagLogNoColor, false, "Disable colored logs")
	pflags.StringP(FlagHome, "", f.defaultHomeDir, "directory for config and data")
	viper.SetEnvPrefix(f.envPrefix)
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_", "-", "_"))
	viper.AutomaticEnv()
}

// EnhanceRootCommand sets the viper and logger in the command context.
func (f *CommandFactory) EnhanceRootCommand(cmd *cobra.Command) {
	f.enhanceCommand(cmd)
	SetCmdServerContext(cmd, f.vipr, f.logger)
}

// ParseCommand parses args against the input rootCmd CLI skeleton then returns the target subcommand,
// a fully realized config map, and a properly configured logger.
// If `WithConfigWriter` was set in the factory options, the config writer will be used to write the app.toml file.
// Internally a viper instance is created and used to bind the flags to the config map.
// Future invocations of EnhanceCommandContext will set the viper instance and logger in the command context.
func (f *CommandFactory) ParseCommand(
	rootCmd *cobra.Command,
	args []string,
) (*cobra.Command, server.ConfigMap, log.Logger, error) {
	f.enhanceCommand(rootCmd)
	cmd, _, err := rootCmd.Find(args)
	if err != nil {
		return nil, nil, nil, err
	}
	// AutoCLI will set this to true for its commands to disable flag parsing in execution.  We want to parse flags here.
	cmd.DisableFlagParsing = false
	if err = cmd.ParseFlags(args); err != nil {
		// help requested, return the command early
		if errors.Is(err, pflag.ErrHelp) {
			return cmd, nil, nil, err
		}
		return nil, nil, nil, err
	}
	home, err := cmd.Flags().GetString(FlagHome)
	if err != nil {
		return nil, nil, nil, err
	}
	configDir := filepath.Join(home, "config")
	if f.configWriter != nil {
		// create app.toml if it does not already exist
		if _, err = os.Stat(filepath.Join(configDir, "app.toml")); os.IsNotExist(err) {
			if err = f.configWriter.WriteConfig(configDir); err != nil {
				return nil, nil, nil, err
			}
		}
	}
	f.vipr, err = ReadConfig(configDir)
	if err != nil {
		return nil, nil, nil, err
	}
	if err = f.vipr.BindPFlags(cmd.Flags()); err != nil {
		return nil, nil, nil, err
	}

	if f.loggerFactory != nil {
		f.logger, err = f.loggerFactory(f.vipr.AllSettings(), cmd.OutOrStdout())
		if err != nil {
			return nil, nil, nil, err
		}
	}

	return cmd, f.vipr.AllSettings(), f.logger, nil
}
