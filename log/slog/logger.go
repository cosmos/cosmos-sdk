// Package slog contains a Logger type that satisfies [cosmossdk.io/log.Logger],
// backed by a standard library [*log/slog.Logger].
package slog

import (
	"log/slog"

	"cosmossdk.io/log/v2"
)

var _ log.Logger = Logger{}

// Logger satisfies [log.Logger] with logging backed by
// an instance of [*slog.Logger].
type Logger struct {
	// we MUST embed slog, otherwise source info will be wrong and always refer to this file, not the caller's file
	*slog.Logger
}

// NewCustomLogger returns a Logger backed by an existing slog.Logger instance.
// All logging methods are called directly on the *slog.Logger;
// therefore it is the caller's responsibility to configure message filtering,
// level filtering, output format, and so on.
func NewCustomLogger(log *slog.Logger) Logger {
	return Logger{Logger: log}
}

func (l Logger) With(keyVals ...any) log.Logger {
	return Logger{Logger: l.Logger.With(keyVals...)}
}

// Impl returns l's underlying [*slog.Logger].
func (l Logger) Impl() any {
	return l.Logger
}
