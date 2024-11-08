package serverv2

import (
	"context"
	"errors"
	"io"
	"os"
	"os/signal"
	"runtime/pprof"
	"slices"
	"strings"
	"syscall"

	"github.com/spf13/cobra"

	"cosmossdk.io/core/server"
	"cosmossdk.io/core/transaction"
	"cosmossdk.io/log"
)

// AddCommands add the server commands to the root command
// It configures the config handling and the logger handling
func AddCommands[T transaction.Tx](
	rootCmd *cobra.Command,
	logger log.Logger,
	appCloser io.Closer,
	globalAppConfig server.ConfigMap,
	globalServerConfig ServerConfig,
	components ...ServerComponent[T],
) (ConfigWriter, error) {
	if len(components) == 0 {
		return nil, errors.New("no components provided")
	}
	srv := NewServer(globalServerConfig, components...)
	cmds := srv.CLICommands()
	startCmd := createStartCommand(srv, appCloser, globalAppConfig, logger)
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
	appCloser io.Closer,
	config server.ConfigMap,
	logger log.Logger,
) *cobra.Command {
	flags := server.StartFlags()

	cmd := &cobra.Command{
		Use:   "start",
		Short: "Run the application",
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
			}()

			return wrapCPUProfile(logger, config, func() error {
				defer func() {
					if err := server.Stop(cmd.Context()); err != nil {
						cmd.PrintErrln("failed to stop servers:", err)
					}

					if err := appCloser.Close(); err != nil {
						cmd.PrintErrln("failed to close application:", err)
					}
				}()

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

// appBuildingCommands are the commands which need a full application to be built
var appBuildingCommands = [][]string{
	{"start"},
	{"genesis", "export"},
}

// IsAppRequired determines if a command requires a full application to be built by
// recursively checking the command hierarchy against known command paths.
//
// The function works by:
// 1. Combining default appBuildingCommands with additional required commands
// 2. Building command paths by traversing up the command tree
// 3. Checking if any known command path matches the current command path
//
// Time Complexity: O(d * p) where d is command depth and p is number of paths
// Space Complexity: O(p) where p is total number of command paths
func IsAppRequired(cmd *cobra.Command, required ...[]string) bool {
	m := make(map[string]bool)
	cmds := append(appBuildingCommands, required...)
	for _, c := range cmds {
		slices.Reverse(c)
		m[strings.Join(c, "")] = true
	}
	cmdPath := make([]string, 0, 5) // Pre-allocate with reasonable capacity
	for {
		cmdPath = append(cmdPath, cmd.Use)
		if _, ok := m[strings.Join(cmdPath, "")]; ok {
			return true
		}
		if cmd.Parent() == nil {
			return false
		}
		cmd = cmd.Parent()
	}
}
