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
	cosmovisor.Logger.Warn().Msg("Use of cosmovisor without the 'run' command is deprecated.")
	cosmovisor.Logger.Warn().Msg("Update your command to: cosmovisor run [args]")
	return Run(args)
}
