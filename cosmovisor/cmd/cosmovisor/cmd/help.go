package cmd

import (
	"fmt"
	"os"

	"github.com/cosmos/cosmos-sdk/cosmovisor"
	"github.com/cosmos/cosmos-sdk/cosmovisor/errors"
)

// ShouldGiveHelp checks the env and provided args to see if help is needed or being requested.
// Help is needed if the os.Getenv(EnvName) env var isn't set.
// Help is requested if the first arg is "help" or "--help"; or the only arg is "-h"
func ShouldGiveHelp(args []string) bool {
	if len(os.Getenv(cosmovisor.EnvName)) == 0 {
		return true
	}
	if len(args) == 0 {
		return false
	}
	return args[0] == "help" || args[0] == "--help" || (len(args) == 1 && args[0] == "-h")
}

// DoHelp outputs help text, config info, and attempts to run the binary with the --help flag.
func DoHelp() {
	// Output the help text
	fmt.Println(GetHelpText())
	// If the config isn't valid, say what's wrong and we're done.
	cfg, cerr := cosmovisor.GetConfigFromEnv()
	switch err := cerr.(type) {
	case nil:
		// Nothing to do. Move on.
	case *errors.MultiError:
		fmt.Fprintf(os.Stderr, "[cosmovisor] multiple configuration errors found:\n")
		for i, e := range err.GetErrors() {
			fmt.Fprintf(os.Stderr, "  %d: %v\n", i+1, e)
		}
		return
	default:
		fmt.Fprintf(os.Stderr, "[cosmovisor] %v\n", err)
		return
	}
	// If the config's legit, output what we see it as.
	fmt.Println("[cosmovisor] config is valid:")
	fmt.Println(cfg.DetailString())
	// Attempt to run the configured binary with the --help flag.
	if err := cosmovisor.RunHelp(cfg, os.Stdout, os.Stderr); err != nil {
		fmt.Fprintf(os.Stderr, "[cosmovisor] %v\n", err)
	}
}

// GetHelpText creates the help text multi-line string.
func GetHelpText() string {
	return `Cosmosvisor - A process manager for Cosmos SDK application binaries.

Cosmovisor monitors the governance module for incoming chain upgrade proposals.
If it sees a proposal that gets approved, cosmovisor can automatically download
the new binary, stop the current binary, switch from the old binary to the new one,
and finally restart the node with the new binary.

Configuration of Cosmovisor is done through environment variables, which are
documented in: https://github.com/cosmos/cosmos-sdk/tree/master/cosmovisor/README.md
`
}
