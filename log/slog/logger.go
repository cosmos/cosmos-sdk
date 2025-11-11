// Package slog contains a Logger type that satisfies [cosmossdk.io/log.Logger],
// backed by a standard library [*log/slog.Logger].
package slog

import (
	"log/slog"

	"cosmossdk.io/log"
)

// Logger satisfies [log.Logger] with logging backed by
// an instance of [*slog.Logger].
type Logger = log.SlogLogger

// NewCustomLogger returns a Logger backed by an existing slog.Logger instance.
// All logging methods are called directly on the *slog.Logger;
// therefore it is the caller's responsibility to configure message filtering,
// level filtering, output format, and so on.
func NewCustomLogger(logger *slog.Logger) Logger {
	return log.NewSlogLogger(logger)
}
