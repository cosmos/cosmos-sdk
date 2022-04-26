package cmd

import (
	"strings"

	"github.com/rs/zerolog"
)

// RunCosmovisorCommand executes the desired cosmovisor command.
func RunCosmovisorCommand(logger *zerolog.Logger, args []string) error {
	arg0 := ""
	if len(args) > 0 {
		arg0 = strings.TrimSpace(args[0])
	}

	switch {
	case IsVersionCommand(arg0):
		return PrintVersion(logger, args[1:])

	case ShouldGiveHelp(arg0):
		return DoHelp()

	case IsRunCommand(arg0):
		return Run(logger, args[1:])
	}

	warnRun := func() {
		logger.Warn().Msg("Use of cosmovisor without the 'run' command is deprecated. Use: cosmovisor run [args]")
	}
	warnRun()
	defer warnRun()

	return Run(logger, args)
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
