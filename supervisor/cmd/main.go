package main

import (
	"fmt"
	"os"

	"github.com/cosmos/cosmos-sdk/supervisor"
)

func main() {
	err := Run(os.Args[1:])
	if err != nil {
		fmt.Printf("%+v\n", err)
		os.Exit(1)
	}
}

// Run is the main loop, but returns an error
func Run(args []string) error {
	cfg, err := supervisor.GetConfigFromEnv()
	if err != nil {
		return err
	}
	doUpgrade, err := supervisor.LaunchProcess(cfg, args, os.Stdout, os.Stderr)

	// if RestartAfterUpgrade, we launch after a successful upgrade (only condition LaunchProcess returns nil)
	for cfg.RestartAfterUpgrade && err == nil && doUpgrade {
		doUpgrade, err = supervisor.LaunchProcess(cfg, args, os.Stdout, os.Stderr)
	}
	return err
}
