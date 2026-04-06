package main

import (
	"io"
	"os"

	"cosmossdk.io/tools/cosmovisor/v2/internal"
)

// DefaultRunConfig defines a default RunConfig that writes to os.Stdout and os.Stderr
var DefaultRunConfig = RunConfig{
	StdIn:  os.Stdin,
	StdOut: os.Stdout,
	StdErr: os.Stderr,
}

type RunConfig = internal.RunConfig

type RunOption func(*RunConfig)

// StdInRunOption sets the StdIn reader for the Run command
func StdInRunOption(r io.Reader) RunOption {
	return func(cfg *RunConfig) {
		cfg.StdIn = r
	}
}

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
