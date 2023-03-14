package main

import (
	"cosmossdk.io/log"
	"cosmossdk.io/tools/cosmovisor"
	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
)

var runCmd = &cobra.Command{
	Use:                "run",
	Short:              "Run an APP command.",
	SilenceUsage:       true,
	DisableFlagParsing: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		return Run(cmd, args)
	},
}

// Run runs the configured program with the given args and monitors it for upgrades.
func Run(cmd *cobra.Command, args []string, options ...RunOption) error {
	cfg, err := cosmovisor.GetConfigFromEnv()
	if err != nil {
		return err
	}

	logger := cmd.Context().Value(log.ContextKey).(log.Logger)

	if cfg.DisableLogs {
		logger = log.NewCustomLogger(zerolog.Nop())
	}

	runCfg := DefaultRunConfig
	for _, opt := range options {
		opt(&runCfg)
	}

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
