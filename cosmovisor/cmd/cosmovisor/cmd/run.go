package cmd

import (
	"github.com/cosmos/cosmos-sdk/cosmovisor"
)

// RunArgs are the strings that indicate a cosmovisor run command.
var RunArgs = []string{"run"}

// IsRunCommand checks if the given args indicate that a run is desired.
func IsRunCommand(arg string) bool {
	return isOneOf(arg, RunArgs)
}

// Run runs the configured program with the given args and monitors it for upgrades.
func (cv *Cosmovisor) Run(args []string, options ...RunOption) error {
	cfg, err := cosmovisor.GetConfigFromEnv()
	if err != nil {
		return err
	}

	runCfg := DefaultRunConfig
	for _, opt := range options {
		opt(&runCfg)
	}

	launcher, err := cosmovisor.NewLauncher(cv.logger, cfg)
	if err != nil {
		return err
	}

	doUpgrade, err := launcher.Run(args, runCfg.StdOut, runCfg.StdErr)
	// if RestartAfterUpgrade, we launch after a successful upgrade (only condition LaunchProcess returns nil)
	for cfg.RestartAfterUpgrade && err == nil && doUpgrade {
		cv.logger.Info().Str("app", cfg.Name).Msg("upgrade detected, relaunching")
		doUpgrade, err = launcher.Run(args, runCfg.StdOut, runCfg.StdErr)
	}
	if doUpgrade && err == nil {
		cv.logger.Info().Msg("upgrade detected, DAEMON_RESTART_AFTER_UPGRADE is off. Verify new upgrade and start cosmovisor again.")
	}

	return err
}
