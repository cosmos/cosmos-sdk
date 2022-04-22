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
func Run(args []string, options ...cosmovisor.RunOption) error {
	cfg, err := cosmovisor.GetConfigFromEnv()
	if err != nil {
		return err
	}

	runCfg := cosmovisor.DefaultRunConfig
	for _, opt := range options {
		opt(&runCfg)
	}

	launcher, err := cosmovisor.NewLauncher(cfg, &runCfg)
	if err != nil {
		return err
	}

	doUpgrade, err := launcher.Run(args)
	// if RestartAfterUpgrade, we launch after a successful upgrade (only condition LaunchProcess returns nil)
	for cfg.RestartAfterUpgrade && err == nil && doUpgrade {
		cosmovisor.Logger.Info().Str("app", cfg.Name).Msg("upgrade detected, relaunching")
		doUpgrade, err = launcher.Run(args)
	}
	if doUpgrade && err == nil {
		cosmovisor.Logger.Info().Msg("upgrade detected, DAEMON_RESTART_AFTER_UPGRADE is off. Verify new upgrade and start cosmovisor again.")
	}

	return err
}
