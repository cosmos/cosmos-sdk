// Package slog contains a Logger type that satisfies [cosmossdk.io/log.Logger],
// backed by a standard library [*log/slog.Logger].
package slog

import (
	"log/slog"

	"cosmossdk.io/log"
)

var _ log.Logger = Logger{}

// Logger satisfies [log.Logger] with logging backed by
// an instance of [*slog.Logger].
type Logger struct {
	log *slog.Logger
}

// NewCustomLogger returns a Logger backed by an existing slog.Logger instance.
// All logging methods are called directly on the *slog.Logger;
// therefore it is the caller's responsibility to configure message filtering,
// level filtering, output format, and so on.
func NewCustomLogger(log *slog.Logger) Logger {
	return Logger{log: log}
}

func (l Logger) Info(msg string, keyVals ...any) {
	l.log.Info(msg, keyVals...)
}

func (l Logger) Warn(msg string, keyVals ...any) {
	l.log.Warn(msg, keyVals...)
}

func (l Logger) Error(msg string, keyVals ...any) {
	l.log.Error(msg, keyVals...)
}

func (l Logger) Debug(msg string, keyVals ...any) {
	l.log.Debug(msg, keyVals...)
}

func (l Logger) With(keyVals ...any) log.Logger {
	return Logger{log: l.log.With(keyVals...)}
}

// Impl returns l's underlying [*slog.Logger].
func (l Logger) Impl() any {
	return l.log
}
