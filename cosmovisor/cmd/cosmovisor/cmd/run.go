package cmd

import (
	"os"

	"github.com/cosmos/cosmos-sdk/cosmovisor"
)

// RunArgs are the strings that indicate a cosmovisor run command.
var RunArgs = []string{"run"}

// IsRunCommand checks if the given args indicate that a run is desired.
func IsRunCommand(arg string) bool {
	return isOneOf(arg, RunArgs)
}

// Run runs the configured program with the given args and monitors it for upgrades.
func Run(args []string) error {
	cfg, cerr := cosmovisor.GetConfigFromEnv()
	cosmovisor.LogConfigOrError(cosmovisor.Logger, cfg, cerr)
	if cerr != nil {
		return cerr
	}
	launcher, err := cosmovisor.NewLauncher(cfg)
	if err != nil {
		return err
	}

	doUpgrade, err := launcher.Run(args, os.Stdout, os.Stderr)
	// if RestartAfterUpgrade, we launch after a successful upgrade (only condition LaunchProcess returns nil)
	for cfg.RestartAfterUpgrade && err == nil && doUpgrade {
		cosmovisor.Logger.Info().Str("app", cfg.Name).Msg("upgrade detected, relaunching")
		doUpgrade, err = launcher.Run(args, os.Stdout, os.Stderr)
	}
	if doUpgrade && err == nil {
		cosmovisor.Logger.Info().Msg("upgrade detected, DAEMON_RESTART_AFTER_UPGRADE is off. Verify new upgrade and start cosmovisor again.")
	}

	return err
}
