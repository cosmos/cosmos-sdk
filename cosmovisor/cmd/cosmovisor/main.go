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
	if cosmovisor.ShouldGiveHelp(args) {
		DoHelp()
		return nil
	}
	cfg, err := cosmovisor.GetConfigFromEnv()
	if err != nil {
		return err
	}
	launcher, err := cosmovisor.NewLauncher(cfg)
	if err != nil {
		return err
	}

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

func DoHelp() {
	// Output the help text
	fmt.Println(cosmovisor.GetHelpText())
	// If the config isn't valid, there's nothing else to do.
	cfg, cerr := cosmovisor.GetConfigFromEnv()
	switch err := cerr.(type) {
	case nil:
		// Nothing to do. Move on.
	case *cosmovisor.MultiError:
		fmt.Fprintf(os.Stderr, "[cosmovisor] multiple configuration errors found:\n")
		for i, e := range err.GetErrors() {
			fmt.Fprintf(os.Stderr, "  %d: %v", i+1, e)
		}
	default:
		fmt.Fprintf(os.Stderr, "[cosmovisor] %v", err)
	}
	fmt.Printf("[cosmovisor] config is valid:")
	fmt.Println(cfg.DetailString())
	if err := cosmovisor.RunHelp(cfg, os.Stdout, os.Stderr); err != nil {
		fmt.Fprintf(os.Stderr, "[cosmovisor] %v", err)
	}
}
