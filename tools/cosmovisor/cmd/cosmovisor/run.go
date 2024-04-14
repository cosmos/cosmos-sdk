package main

import (
	"os"

	"github.com/spf13/cobra"

	"cosmossdk.io/tools/cosmovisor"
)

const (
	cfgENV = "CONFIG"
)

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Run an APP command.",
	Long: `Run an APP command. This command is intended to be used by the cosmovisor binary.
Provide config file path via CONFIG environment variable.
`,
	SilenceUsage:       true,
	DisableFlagParsing: true,
	RunE: func(_ *cobra.Command, args []string) error {
		return run(args)
	},
}

// run runs the configured program with the given args and monitors it for upgrades.
func run(args []string, options ...RunOption) error {
	cfg, err := cosmovisor.GetConfigFromFile(os.Getenv(cfgENV))
	if err != nil {
		return err
	}

	runCfg := DefaultRunConfig
	for _, opt := range options {
		opt(&runCfg)
	}

	logger := cfg.Logger(runCfg.StdOut)
	launcher, err := cosmovisor.NewLauncher(logger, cfg)
	if err != nil {
		return err
	}

	doUpgrade, err := launcher.Run(args, runCfg.StdOut, runCfg.StdErr)
	// if RestartAfterUpgrade, we launch after a successful upgrade (given that condition launcher.Run returns nil)
	for cfg.RestartAfterUpgrade && err == nil && doUpgrade {
		logger.Info("upgrade detected, relaunching", "app", cfg.Name)
		doUpgrade, err = launcher.Run(args, runCfg.StdOut, runCfg.StdErr)
	}

	if doUpgrade && err == nil {
		logger.Info("upgrade detected, DAEMON_RESTART_AFTER_UPGRADE is off. Verify new upgrade and start cosmovisor again.")
	}

	return err
}
