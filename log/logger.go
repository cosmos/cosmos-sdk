package log

import (
	"io"
	"time"

	"github.com/rs/zerolog"
)

// Defines commons keys for logging.
const ModuleKey = "module"

// ContextKey is used to store the logger in the context.
var ContextKey struct{}

// Logger is the Cosmos SDK logger interface.
// It maintains as much backward compatibility with the CometBFT logger as possible.
// All functionalities of the logger are available through the Impl() method.
type Logger interface {
	// Info takes a message and a set of key/value pairs and logs with level INFO.
	// The key of the tuple must be a string.
	Info(msg string, keyVals ...any)

	// Error takes a message and a set of key/value pairs and logs with level ERR.
	// The key of the tuple must be a string.
	Error(msg string, keyVals ...any)

	// Debug takes a message and a set of key/value pairs and logs with level DEBUG.
	// The key of the tuple must be a string.
	Debug(msg string, keyVals ...any)

	// With returns a new wrapped logger with additional context provided by a set.
	With(keyVals ...any) Logger

	// Impl returns the underlying logger implementation.
	// It is used to access the full functionalities of the underlying logger.
	// Advanced users can type cast the returned value to the actual logger.
	Impl() any
}

type zeroLogWrapper struct {
	*zerolog.Logger
}

// NewNopLogger returns a new logger that does nothing.
func NewNopLogger() Logger {
	// The custom nopLogger is about 3x faster than a zeroLogWrapper with zerolog.Nop().
	return nopLogger{}
}

// NewLogger returns a new logger that writes to the given destination.
//
// Typical usage from a main function is:
// logger := log.NewLogger(os.Stderr)
//
// Stderr is the typical destination for logs,
// so that any output from your application can still be piped to other processes.
func NewLogger(dst io.Writer) Logger {
	output := zerolog.ConsoleWriter{Out: dst, TimeFormat: time.Kitchen}
	logger := zerolog.New(output).With().Timestamp().Logger()
	return zeroLogWrapper{&logger}
}

// NewLoggerWithKV is shorthand for NewLogger(dst).With(key, value).
func NewLoggerWithKV(dst io.Writer, key, value string) Logger {
	return NewLogger(dst).With(key, value)
}

// NewCustomLogger returns a new logger with the given zerolog logger.
func NewCustomLogger(logger zerolog.Logger) Logger {
	return zeroLogWrapper{&logger}
}

// Info takes a message and a set of key/value pairs and logs with level INFO.
// The key of the tuple must be a string.
func (l zeroLogWrapper) Info(msg string, keyVals ...interface{}) {
	l.Logger.Info().Fields(keyVals).Msg(msg)
}

// Error takes a message and a set of key/value pairs and logs with level DEBUG.
// The key of the tuple must be a string.
func (l zeroLogWrapper) Error(msg string, keyVals ...interface{}) {
	l.Logger.Error().Fields(keyVals).Msg(msg)
}

// Debug takes a message and a set of key/value pairs and logs with level ERR.
// The key of the tuple must be a string.
func (l zeroLogWrapper) Debug(msg string, keyVals ...interface{}) {
	l.Logger.Debug().Fields(keyVals).Msg(msg)
}

// With returns a new wrapped logger with additional context provided by a set.
func (l zeroLogWrapper) With(keyVals ...interface{}) Logger {
	logger := l.Logger.With().Fields(keyVals).Logger()
	return zeroLogWrapper{&logger}
}

// Impl returns the underlying zerolog logger.
// It can be used to used zerolog structured API directly instead of the wrapper.
func (l zeroLogWrapper) Impl() interface{} {
	return l.Logger
}

// FilterKeys returns a new logger that filters out all key/value pairs that do not match the filter.
// This functions assumes that the logger is a zerolog.Logger, which is the case for the logger returned by log.NewLogger().
// NOTE: filtering has a performance impact on the logger.
func FilterKeys(logger Logger, filter FilterFunc) Logger {
	zl, ok := logger.Impl().(*zerolog.Logger)
	if !ok {
		panic("logger is not a zerolog.Logger")
	}

	filteredLogger := zl.Hook(zerolog.HookFunc(func(e *zerolog.Event, lvl zerolog.Level, _ string) {
		// TODO(@julienrbrt) wait for https://github.com/rs/zerolog/pull/527 to be merged
		// keys, err := e.GetKeys()
		// if err != nil {
		// 	panic(err)
		// }

		keys := []string{}

		for _, key := range keys {
			if filter(key, lvl.String()) {
				e.Discard()
				break
			}
		}
	}))

	return NewCustomLogger(filteredLogger)
}

// nopLogger is a Logger that does nothing when called.
// See the "specialized nop logger" benchmark and compare with the "zerolog nop logger" benchmark.
// The custom implementation is about 3x faster.
type nopLogger struct{}

func (nopLogger) Info(string, ...any)  {}
func (nopLogger) Error(string, ...any) {}
func (nopLogger) Debug(string, ...any) {}
func (nopLogger) With(...any) Logger   { return nopLogger{} }
func (nopLogger) Impl() any            { return nopLogger{} }
