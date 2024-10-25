package serverv2

import (
	"context"
	"errors"
	"os"
	"os/signal"
	"path/filepath"
	"runtime/pprof"
	"strings"
	"syscall"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"

	"cosmossdk.io/core/server"
	"cosmossdk.io/core/transaction"
	"cosmossdk.io/log"
)

type RootCmdBuilder struct {
	DefaultHome string
	EnvPrefix   string

	use     string
	fixture CommandFixture
	vipr    *viper.Viper
	logger  log.Logger
}

type CommandFixture interface {
	Bootstrap(cmd *cobra.Command) (WritesConfig, error)
	RootCommand(
		rootCommand *cobra.Command,
		subCommand *cobra.Command,
		logger log.Logger,
		config server.ConfigMap,
	) (*cobra.Command, error)
}

func NewRootCmdBuilder(
	fixture CommandFixture,
	use string,
	defaultHomeBasename string,
) (*RootCmdBuilder, error) {
	f := &RootCmdBuilder{
		fixture: fixture,
		use:     use,
	}
	var err error
	f.DefaultHome, err = f.DefaultHomeDir(defaultHomeBasename)
	if err != nil {
		return nil, err
	}
	return f, nil
}

func (b *RootCmdBuilder) DefaultHomeDir(name string) (string, error) {
	// get the home directory from the environment variable
	// to not clash with the $HOME system variable, when no prefix is set
	// we check the NODE_HOME environment variable
	homeDir, envHome := "", "HOME"
	if len(b.EnvPrefix) > 0 {
		homeDir = os.Getenv(b.EnvPrefix + "_" + envHome)
	} else {
		homeDir = os.Getenv("NODE_" + envHome)
	}
	if homeDir != "" {
		return filepath.Clean(homeDir), nil
	}

	// get user home directory
	userHomeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(userHomeDir, name), nil
}

func (b *RootCmdBuilder) Command() (*cobra.Command, error) {
	cmd := &cobra.Command{Use: b.use, SilenceErrors: true}
	SetPersistentFlags(cmd.PersistentFlags(), b.DefaultHome)
	// update the global viper with the root command's configuration
	viper.SetEnvPrefix(b.EnvPrefix)
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_", "-", "_"))
	viper.AutomaticEnv()
	if b.vipr != nil && b.logger != nil {
		err := SetCmdServerContext(cmd, b.vipr, b.logger)
		if err != nil {
			return nil, err
		}
	}
	return cmd, nil
}

func (b *RootCmdBuilder) Build(
	args []string,
) (*cobra.Command, error) {
	bootstrapCmd, err := b.Command()
	if err != nil {
		return nil, err
	}
	configWriter, err := b.fixture.Bootstrap(bootstrapCmd)
	if err != nil {
		return nil, err
	}
	cmd, _, err := bootstrapCmd.Traverse(args)
	if err != nil {
		return nil, err
	}
	if err = cmd.ParseFlags(args); err != nil {
		// help requested, return the command early
		if errors.Is(err, pflag.ErrHelp) {
			return cmd, nil
		}
		return nil, err
	}
	home, err := cmd.Flags().GetString(FlagHome)
	if err != nil {
		return nil, err
	}
	configDir := filepath.Join(home, "config")
	// create app.toml if it does not already exist
	if _, err = os.Stat(filepath.Join(configDir, "app.toml")); os.IsNotExist(err) {
		if err = configWriter.WriteConfig(configDir); err != nil {
			return nil, err
		}
	}
	b.vipr, err = ReadConfig(configDir)
	if err != nil {
		return nil, err
	}
	if err = b.vipr.BindPFlags(cmd.Flags()); err != nil {
		return nil, err
	}
	b.logger, err = NewLogger(b.vipr, cmd.OutOrStdout())
	if err != nil {
		return nil, err
	}
	rootCmd, err := b.Command()
	if err != nil {
		return nil, err
	}
	return b.fixture.RootCommand(rootCmd, cmd, b.logger, b.vipr.AllSettings())
}

// SetPersistentFlags sets persistent flags which should be used by all server (and client)
// commands.  They control logging behavior and home directory.
func SetPersistentFlags(pflags *pflag.FlagSet, defaultHome string) {
	pflags.String(FlagLogLevel, "info", "The logging level (trace|debug|info|warn|error|fatal|panic|disabled or '*:<level>,<key>:<level>')")
	pflags.String(FlagLogFormat, "plain", "The logging format (json|plain)")
	pflags.Bool(FlagLogNoColor, false, "Disable colored logs")
	pflags.StringP(FlagHome, "", defaultHome, "directory for config and data")
}

// AddCommands add the server commands to the root command
// It configures the config handling and the logger handling
func AddCommands[T transaction.Tx](
	rootCmd *cobra.Command,
	logger log.Logger,
	globalAppConfig server.ConfigMap,
	globalServerConfig ServerConfig,
	components ...ServerComponent[T],
) (WritesConfig, error) {
	if len(components) == 0 {
		return nil, errors.New("no components provided")
	}
	srv := NewServer(globalServerConfig, components...)
	cmds := srv.CLICommands()
	startCmd := createStartCommand(srv, globalAppConfig, logger)
	// TODO necessary? won't the parent context be inherited?
	startCmd.SetContext(rootCmd.Context())
	cmds.Commands = append(cmds.Commands, startCmd)
	rootCmd.AddCommand(cmds.Commands...)

	if len(cmds.Queries) > 0 {
		if queryCmd := findSubCommand(rootCmd, "query"); queryCmd != nil {
			queryCmd.AddCommand(cmds.Queries...)
		} else {
			queryCmd := topLevelCmd(rootCmd.Context(), "query", "Querying subcommands")
			queryCmd.Aliases = []string{"q"}
			queryCmd.AddCommand(cmds.Queries...)
			rootCmd.AddCommand(queryCmd)
		}
	}

	if len(cmds.Txs) > 0 {
		if txCmd := findSubCommand(rootCmd, "tx"); txCmd != nil {
			txCmd.AddCommand(cmds.Txs...)
		} else {
			txCmd := topLevelCmd(rootCmd.Context(), "tx", "Transactions subcommands")
			txCmd.AddCommand(cmds.Txs...)
			rootCmd.AddCommand(txCmd)
		}
	}

	return srv, nil
}

// createStartCommand creates the start command for the application.
func createStartCommand[T transaction.Tx](
	server *Server[T],
	config server.ConfigMap,
	logger log.Logger,
) *cobra.Command {
	flags := server.StartFlags()

	cmd := &cobra.Command{
		Use:         "start",
		Short:       "Run the application",
		Annotations: map[string]string{"needs-app": "true"},
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, cancelFn := context.WithCancel(cmd.Context())
			go func() {
				sigCh := make(chan os.Signal, 1)
				signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
				select {
				case sig := <-sigCh:
					cancelFn()
					cmd.Printf("caught %s signal\n", sig.String())
				case <-ctx.Done():
					// If the root context is canceled (which is likely to happen in tests involving cobra commands),
					// don't block waiting for the OS signal before stopping the server.
					cancelFn()
				}

				if err := server.Stop(ctx); err != nil {
					cmd.PrintErrln("failed to stop servers:", err)
				}
			}()

			return wrapCPUProfile(logger, config, func() error {
				return server.Start(ctx)
			})
		},
	}

	// add the start flags to the command
	for _, startFlags := range flags {
		cmd.Flags().AddFlagSet(startFlags)
	}

	return cmd
}

// wrapCPUProfile starts CPU profiling, if enabled, and executes the provided
// callbackFn, then waits for it to return.
func wrapCPUProfile(logger log.Logger, cfg server.ConfigMap, callbackFn func() error) error {
	cpuProfileFile, ok := cfg[FlagCPUProfiling]
	if !ok {
		// if cpu profiling is not enabled, just run the callback
		return callbackFn()
	}

	f, err := os.Create(cpuProfileFile.(string))
	if err != nil {
		return err
	}

	logger.Info("starting CPU profiler", "profile", cpuProfileFile)
	if err := pprof.StartCPUProfile(f); err != nil {
		_ = f.Close()
		return err
	}

	defer func() {
		logger.Info("stopping CPU profiler", "profile", cpuProfileFile)
		pprof.StopCPUProfile()
		if err := f.Close(); err != nil {
			logger.Info("failed to close cpu-profile file", "profile", cpuProfileFile, "err", err.Error())
		}
	}()

	return callbackFn()
}

// findSubCommand finds a sub-command of the provided command whose Use
// string is or begins with the provided subCmdName.
// It verifies the command's aliases as well.
func findSubCommand(cmd *cobra.Command, subCmdName string) *cobra.Command {
	for _, subCmd := range cmd.Commands() {
		use := subCmd.Use
		if use == subCmdName || strings.HasPrefix(use, subCmdName+" ") {
			return subCmd
		}

		for _, alias := range subCmd.Aliases {
			if alias == subCmdName || strings.HasPrefix(alias, subCmdName+" ") {
				return subCmd
			}
		}
	}
	return nil
}

// topLevelCmd creates a new top-level command with the provided name and
// description. The command will have DisableFlagParsing set to false and
// SuggestionsMinimumDistance set to 2.
func topLevelCmd(ctx context.Context, use, short string) *cobra.Command {
	cmd := &cobra.Command{
		Use:                        use,
		Short:                      short,
		DisableFlagParsing:         false,
		SuggestionsMinimumDistance: 2,
	}
	cmd.SetContext(ctx)

	return cmd
}
