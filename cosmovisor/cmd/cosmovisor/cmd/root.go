package cmd

import (
	"strings"

	"github.com/cosmos/cosmos-sdk/cosmovisor"
)

// RunCosmovisorCommand executes the desired cosmovisor command.
func RunCosmovisorCommand(args []string) error {
	switch {
	case ShouldGiveHelp(args):
		DoHelp()
		return nil
	case IsVersionCommand(args):
		PrintVersion()
		return nil
	case IsRunCommand(args):
		return Run(args[1:])
	}
	warnRun := func() {
		cosmovisor.Logger.Warn().Msg("Use of cosmovisor without the 'run' command is deprecated. Use: cosmovisor run [args]")
	}
	warnRun()
	defer warnRun()
	return Run(args)
}

// isOneOf returns true if the given arg equals one of the provided options (ignoring case).
func isOneOf(arg string, options []string) bool {
	for _, opt := range options {
		if strings.EqualFold(arg, opt) {
			return true
		}
	}
	return false
}
