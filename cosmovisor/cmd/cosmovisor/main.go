package main

import (
	"fmt"
	"os"

	"github.com/cosmos/cosmos-sdk/cosmovisor"
)

func main() {
	if err := Run(os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "[cosmovisor] %+v\n", err)
		os.Exit(1)
	}
}

// Run is the main loop, but returns an error
func Run(args []string) error {
	help := cosmovisor.HelpRequested(args)
	if help {
		fmt.Println(cosmovisor.GetHelpText())
	}
	cfg, err := cosmovisor.GetConfigFromEnv()
	if err != nil {
		return err
	}
	launcher, err := cosmovisor.NewLauncher(cfg)
	if err != nil {
		return err
	}

	// TODO: 10126 - if help requested, call help on the binary.
	doUpgrade, err := launcher.Run(args, os.Stdout, os.Stderr)
	// if RestartAfterUpgrade, we launch after a successful upgrade (only condition LaunchProcess returns nil)
	for cfg.RestartAfterUpgrade && err == nil && doUpgrade {
		fmt.Println("[cosmovisor] upgrade detected, relaunching the app ", cfg.Name)
		doUpgrade, err = launcher.Run(args, os.Stdout, os.Stderr)
	}
	if doUpgrade && err == nil {
		fmt.Println("[cosmovisor] upgrade detected, DAEMON_RESTART_AFTER_UPGRADE is off. Verify new upgrade and start cosmovisor again.")
	}
	return err
}
