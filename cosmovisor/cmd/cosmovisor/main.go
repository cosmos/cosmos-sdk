package main

import (
	"fmt"
	"os"

	"github.com/cosmos/cosmos-sdk/cosmovisor"
)

func main() {
	if err := Run(os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "%+v\n", err)
		os.Exit(1)
	}
}

// Run is the main loop, but returns an error
func Run(args []string) error {
	cfg, err := cosmovisor.GetConfigFromEnv()
	if err != nil {
		return err
	}

	var doUpgrade bool
	for {
		if doUpgrade, err = cosmovisor.LaunchProcess(cfg, args, os.Stdout, os.Stderr); !cfg.RestartAfterUpgrade || err != nil || !doUpgrade {
			break
		}
	}

	return err
}
