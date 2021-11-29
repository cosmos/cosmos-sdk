package cmd

import (
	"strings"

	"github.com/cosmos/cosmos-sdk/cosmovisor"
)

// RunCosmovisorCommand executes the desired cosmovisor command.
func RunCosmovisorCommand(args []string) error {
	arg0 := ""
	if len(args) > 0 {
		arg0 = strings.TrimSpace(args[0])
	}
	switch {
	case IsVersionCommand(arg0):
		return PrintVersion()
	case ShouldGiveHelp(arg0):
		DoHelp()
		return nil
	case IsRunCommand(arg0):
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
