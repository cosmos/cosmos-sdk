package log

import "log/slog"

var _ Logger = SlogLogger{}

// SlogLogger satisfies [log.Logger] with logging backed by
// an instance of [*slog.Logger].
type SlogLogger struct {
	log *slog.Logger
}

// NewSlogLogger returns a Logger backed by an existing slog.Logger instance.
// All logging methods are called directly on the *slog.Logger;
// therefore it is the caller's responsibility to configure message filtering,
// level filtering, output format, and so on.
func NewSlogLogger(log *slog.Logger) SlogLogger {
	return SlogLogger{log: log}
}

func (l SlogLogger) Info(msg string, keyVals ...any) {
	l.log.Info(msg, keyVals...)
}

func (l SlogLogger) Warn(msg string, keyVals ...any) {
	l.log.Warn(msg, keyVals...)
}

func (l SlogLogger) Error(msg string, keyVals ...any) {
	l.log.Error(msg, keyVals...)
}

func (l SlogLogger) Debug(msg string, keyVals ...any) {
	l.log.Debug(msg, keyVals...)
}

func (l SlogLogger) With(keyVals ...any) Logger {
	return SlogLogger{log: l.log.With(keyVals...)}
}

// Impl returns l's underlying [*slog.Logger].
func (l SlogLogger) Impl() any {
	return l.log
}
