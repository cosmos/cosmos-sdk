package cmd

import (
	"fmt"
	"github.com/cosmos/cosmos-sdk/cosmovisor"
	"os"
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

func DoHelp() {
	// Output the help text
	fmt.Println(GetHelpText())
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

// GetHelpText creates the help text multi-line string.
func GetHelpText() string {
	return fmt.Sprintf(`Cosmosvisor - A process manager for Cosmos SDK application binaries.

Cosmovisor monitors the governance module for incoming chain upgrade proposals.
If it sees a proposal that gets approved, cosmovisor can automatically download
the new binary, stop the current binary, switch from the old binary to the new one,
and finally restart the node with the new binary.

Command line arguments are passed on to the configured binary.

Configuration of Cosmoviser is done through the following environment variables:
    %s
        The location where the cosmovisor/ directory is kept that contains the genesis binary,
        the upgrade binaries, and any additional auxiliary files associated with each binary
        (e.g. $HOME/.gaiad, $HOME/.regend, $HOME/.simd, etc.).
    %s
        The name of the binary itself (e.g. gaiad, regend, simd, etc.).
    %s
        Optional, default is "false" (cosmovisor will not auto-download new binaries).
        Enables auto-downloading of new binaries. For security reasons, this is intended
        for full nodes rather than validators.
        Valid values: true, false.
    %s
        Optional, default is "true" (cosmovisor will restart the subprocess after upgrade).
        If true, will restart the subprocess with the same command-line arguments and flags
        (but with the new binary) after a successful upgrade.
        If false, cosmovisor stops running after an upgrade and requires the system administrator
        to manually restart it.
        Note that cosmovisor will not auto-restart the subprocess if there was an error.
        Valid values: true, false.
    %s
        Optional, default is "300".
        The interval length in milliseconds for polling the upgrade plan file.
        Valid values: Integers greater than 0.
    %s
        Optional, default is "false" (cosmovisor will not auto-download new binaries).
        If false, data will be backed up before trying the upgrade.
        If true, data will NOT be backed up before trying the upgrade.
        This is useful (and recommended) in case of failures and when needed to rollback.
        It is advised to use backup option, i.e. UNSAFE_SKIP_BACKUP=false
        Valid values: true, false.

`, cosmovisor.EnvHome, cosmovisor.EnvName, cosmovisor.EnvDownloadBin,
cosmovisor.EnvRestartUpgrade, cosmovisor.EnvInterval, cosmovisor.EnvSkipBackup)
}