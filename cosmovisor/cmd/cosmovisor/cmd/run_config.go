package cmd

import (
	"io"
	"os"

	"github.com/cosmos/cosmos-sdk/cosmovisor/logging"
)

// DefaultRunConfig defintes a default RunConfig that writes to os.Stdout and os.Stderr
var DefaultRunConfig = RunConfig{
	DisableLogging: false,
	StdOut:         os.Stdout,
	StdErr:         os.Stderr,
}

// RunConfig defines the configuration for running a command
type RunConfig struct {
	DisableLogging bool

	StdOut io.Writer
	StdErr io.Writer
}

type RunOption func(*RunConfig)

// StdOutRunOption sets the StdOut writer for the Run command
func StdOutRunOption(w io.Writer) RunOption {
	return func(cfg *RunConfig) {
		cfg.StdOut = w
	}
}

// SdErrRunOption sets the StdErr writer for the Run command
func StdErrRunOption(w io.Writer) RunOption {
	return func(cfg *RunConfig) {
		cfg.StdErr = w
	}
}

// DisableLoggingRunOption disables logging for the specific Run command
func DisableLogging(logger logging.Logger) RunOption {
	return func(cfg *RunConfig) {
		cfg.DisableLogging = true
		logger.DisableLogger()
	}
}
