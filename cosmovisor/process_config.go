package cosmovisor

import (
	"io"
	"os"
)

var DefaultRunConfig = RunConfig{
	DisableLogging: false,
	StdOut:         os.Stdout,
	StdErr:         os.Stderr,
}

type RunConfig struct {
	DisableLogging bool

	StdOut io.Writer
	StdErr io.Writer
}

type RunOption func(*RunConfig)

func StdOut(w io.Writer) RunOption {
	return func(cfg *RunConfig) {
		cfg.StdOut = w
	}
}

func StdErr(w io.Writer) RunOption {
	return func(cfg *RunConfig) {
		cfg.StdErr = w
	}
}

func DisableLogging() RunOption {
	return func(cfg *RunConfig) {
		cfg.DisableLogging = true
		DisableLogger()
	}
}
