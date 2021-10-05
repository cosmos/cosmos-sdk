package cmd

import (
	"github.com/cosmos/cosmos-sdk/cosmovisor"
)

// RunCosmovisorCommand executes the desired cosmovisor command.
func RunCosmovisorCommand(args []string) error {
	switch {
	case ShouldGiveHelp(args):
		DoHelp()
		return nil
	case isVersionCommand(args):
		printVersion()
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
