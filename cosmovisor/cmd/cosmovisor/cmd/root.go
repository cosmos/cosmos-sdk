package cmd

import (
	"strings"

	"github.com/cosmos/cosmos-sdk/cosmovisor/logging"
)

type Cosmovisor struct {
	logger *logging.Logger
}

func NewCosmovisor(logger *logging.Logger) Cosmovisor {
	return Cosmovisor{
		logger: logger,
	}
}

// RunCosmovisorCommand executes the desired cosmovisor command.
func (cv *Cosmovisor) RunCosmovisorCommand(args []string) error {
	arg0 := ""
	if len(args) > 0 {
		arg0 = strings.TrimSpace(args[0])
	}

	switch {
	case IsVersionCommand(arg0):
		return cv.PrintVersion(args[1:])

	case ShouldGiveHelp(arg0):
		return cv.DoHelp()

	case IsRunCommand(arg0):
		return cv.Run(args[1:])
	}

	warnRun := func() {
		cv.logger.Warn().Msg("Use of cosmovisor without the 'run' command is deprecated. Use: cosmovisor run [args]")
	}
	warnRun()
	defer warnRun()

	return cv.Run(args)
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
