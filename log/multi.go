package log

import "context"

type multiLogger struct {
	loggers []Logger
}

// NewMultiLogger returns a Logger that dispatches to all provided loggers.
// SetVerboseMode propagates to any logger that implements VerboseModeLogger.
func NewMultiLogger(loggers ...Logger) Logger {
	return multiLogger{loggers: loggers}
}

func (m multiLogger) Info(msg string, kv ...any) {
	for _, l := range m.loggers {
		l.Info(msg, kv...)
	}
}

func (m multiLogger) Warn(msg string, kv ...any) {
	for _, l := range m.loggers {
		l.Warn(msg, kv...)
	}
}

func (m multiLogger) Error(msg string, kv ...any) {
	for _, l := range m.loggers {
		l.Error(msg, kv...)
	}
}

func (m multiLogger) Debug(msg string, kv ...any) {
	for _, l := range m.loggers {
		l.Debug(msg, kv...)
	}
}

func (m multiLogger) InfoContext(ctx context.Context, msg string, kv ...any) {
	for _, l := range m.loggers {
		l.InfoContext(ctx, msg, kv...)
	}
}

func (m multiLogger) WarnContext(ctx context.Context, msg string, kv ...any) {
	for _, l := range m.loggers {
		l.WarnContext(ctx, msg, kv...)
	}
}

func (m multiLogger) ErrorContext(ctx context.Context, msg string, kv ...any) {
	for _, l := range m.loggers {
		l.ErrorContext(ctx, msg, kv...)
	}
}

func (m multiLogger) DebugContext(ctx context.Context, msg string, kv ...any) {
	for _, l := range m.loggers {
		l.DebugContext(ctx, msg, kv...)
	}
}

func (m multiLogger) With(kv ...any) Logger {
	ls := make([]Logger, len(m.loggers))
	for i, l := range m.loggers {
		ls[i] = l.With(kv...)
	}
	return multiLogger{ls}
}

func (m multiLogger) Impl() any { return m.loggers }

func (m multiLogger) SetVerboseMode(enable bool) {
	for _, l := range m.loggers {
		if v, ok := l.(VerboseModeLogger); ok {
			v.SetVerboseMode(enable)
		}
	}
}

var _ VerboseModeLogger = multiLogger{}
