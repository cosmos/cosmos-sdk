package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"cosmossdk.io/tools/cosmovisor"
)

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Run an APP command.",
	Long: `Run an APP command. This command is intended to be used by the cosmovisor binary.
Provide '--cosmovisor-config' file path in command args or set env variables to load configuration.
`,
	SilenceUsage:       true,
	DisableFlagParsing: true,
	RunE: func(_ *cobra.Command, args []string) error {
		cfgPath, args, err := parseCosmovisorConfig(args)
		if err != nil {
			return fmt.Errorf("failed to parse cosmovisor config: %w", err)
		}

		return run(cfgPath, args)
	},
}

// run runs the configured program with the given args and monitors it for upgrades.
func run(cfgPath string, args []string, options ...RunOption) error {
	cfg, err := cosmovisor.GetConfigFromFile(cfgPath)
	if err != nil {
		return err
	}

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
	launcher, err := cosmovisor.NewLauncher(logger, cfg)
	if err != nil {
		return err
	}

	doUpgrade, err := launcher.Run(args, runCfg.StdIn, runCfg.StdOut, runCfg.StdErr)
	// if RestartAfterUpgrade, we launch after a successful upgrade (given that condition launcher.Run returns nil)
	for cfg.RestartAfterUpgrade && err == nil && doUpgrade {
		logger.Info("upgrade detected, relaunching", "app", cfg.Name)
		doUpgrade, err = launcher.Run(args, runCfg.StdIn, runCfg.StdOut, runCfg.StdErr)
	}

	if doUpgrade && err == nil {
		logger.Info("upgrade detected, DAEMON_RESTART_AFTER_UPGRADE is off. Verify new upgrade and start cosmovisor again.")
	}

	return err
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
