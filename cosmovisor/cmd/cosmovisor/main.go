package main

import (
	"fmt"
	"os"

	"github.com/cosmos/cosmos-sdk/cosmovisor"
	"github.com/cosmos/cosmos-sdk/cosmovisor/cmd/cosmovisor/cmd"
	"github.com/cosmos/cosmos-sdk/cosmovisor/errors"
)

func main() {
	cosmovisor.SetupLogging()
	if err := Run(os.Args[1:]); err != nil {
		cosmovisor.Logger.Error().Err(err).Msg("")
		os.Exit(1)
	}
}

// Run is the main loop, but returns an error
func Run(args []string) error {
	cmd.RunCosmovisorCommands(args)

	cfg, cerr := cosmovisor.GetConfigFromEnv()
	if cerr != nil {
		switch err := cerr.(type) {
		case *errors.MultiError:
			cosmovisor.Logger.Error().Msg("multiple configuration errors found:")
			for i, e := range err.GetErrors() {
				cosmovisor.Logger.Error().Err(e).Msg(fmt.Sprintf("  %d:", i+1))
			}
		default:
			cosmovisor.Logger.Error().Err(err).Msg("configuration error:")
		}
		return cerr
	}
	cosmovisor.Logger.Info().Msg("Configuration is valid:\n" + cfg.DetailString())
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
