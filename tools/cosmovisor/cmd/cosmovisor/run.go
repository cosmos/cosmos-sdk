package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/spf13/cobra"

	"cosmossdk.io/tools/cosmovisor/v2"
	"cosmossdk.io/tools/cosmovisor/v2/internal"
)

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Run an APP command.",
	Long: `Run an APP command. This command is intended to be used by the cosmovisor binary.
Provide '--cosmovisor-config' file path in command args or set env variables to load configuration.
`,
	SilenceUsage:       true,
	DisableFlagParsing: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfgPath, args, err := parseCosmovisorConfig(args)
		if err != nil {
			return fmt.Errorf("failed to parse cosmovisor config: %w", err)
		}

		return run(cmd.Context(), cfgPath, args)
	},
}

// run runs the configured program with the given args and monitors it for upgrades.
func run(ctx context.Context, cfgPath string, args []string, options ...RunOption) error {
	cfg, err := cosmovisor.GetConfigFromFile(cfgPath)
	if err != nil {
		return err
	}

	ctx, _ = signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGTERM)
	shutdownChan := make(chan os.Signal, 1)
	signal.Notify(shutdownChan, syscall.SIGINT, syscall.SIGTERM)
	// ensure we shutdown if the process is killed and context cancellation doesn't cause an exit on its own
	go func() {
		<-shutdownChan
		fmt.Println("Received shutdown signal, exiting gracefully...")
		time.Sleep(cfg.ShutdownGrace)
		fmt.Println("Forcing process shutdown")
		os.Exit(0)
	}()

	runCfg := DefaultRunConfig
	for _, opt := range options {
		opt(&runCfg)
	}

	// set current working directory to $DAEMON_NAME/cosmosvisor
	// to allow current symlink to be relative
	if err = os.Chdir(cfg.Root()); err != nil {
		return err
	}

	logger := cfg.Logger(runCfg.StdOut)
	runner := internal.NewRunner(cfg, runCfg, logger)
	return runner.Start(ctx, args)
}

func parseCosmovisorConfig(args []string) (string, []string, error) {
	var configFilePath string
	for i, arg := range args {
		// Check if the argument is the config flag
		if strings.EqualFold(arg, fmt.Sprintf("--%s", cosmovisor.FlagCosmovisorConfig)) ||
			strings.EqualFold(arg, fmt.Sprintf("-%s", cosmovisor.FlagCosmovisorConfig)) {
			// Check if there is an argument after the flag which should be the config file path
			if i+1 >= len(args) {
				return "", nil, fmt.Errorf("--%s requires an argument", cosmovisor.FlagCosmovisorConfig)
			}

			configFilePath = args[i+1]
			// Remove the flag and its value from the arguments
			args = append(args[:i], args[i+2:]...)
			break
		}
	}

	return configFilePath, args, nil
}
