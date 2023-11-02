package main

import (
	"io"
	"os"
)

// DefaultRunConfig defintes a default RunConfig that writes to os.Stdout and os.Stderr
var DefaultRunConfig = RunConfig{
	StdOut: os.Stdout,
	StdErr: os.Stderr,
}

// RunConfig defines the configuration for running a command
type RunConfig struct {
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

// StdErrRunOption sets the StdErr writer for the Run command
func StdErrRunOption(w io.Writer) RunOption {
	return func(cfg *RunConfig) {
		cfg.StdErr = w
	}
}
