package log

import (
	"os"
	"time"

	"github.com/rs/zerolog"
)

// Defines commons keys for logging
const ModuleKey = "module"

// ContextKey is used to store the logger in the context
var ContextKey struct{}

// Logger is the Cosmos SDK logger interface
// It aimed to be used as a wrapper around zerolog
// It maintains as much backward compatibility with the CometBFT logger as possible (use server.CometZeroLogWrapper for 100% compatibility)
// All functionalities of Zerolog are available through the Impl() method
type Logger interface {
	// Info takes a message and a set of key/value pairs and logs with level INFO.
	// The key of the tuple must be a string.
	Info(msg string, keyVals ...interface{})

	// Error takes a message and a set of key/value pairs and logs with level DEBUG.
	// The key of the tuple must be a string.
	Error(msg string, keyVals ...interface{})

	// Debug takes a message and a set of key/value pairs and logs with level ERR.
	// The key of the tuple must be a string.
	Debug(msg string, keyVals ...interface{})

	// With returns a new wrapped logger with additional context provided by a set
	With(keyVals ...interface{}) Logger

	// Impl returns the underlying logger implementation
	// It is used to access the full functionalities of the underlying logger
	// Advanced users can type cast the returned value to the actual logger
	Impl() interface{}
}

type ZeroLogWrapper struct {
	*zerolog.Logger
}

// NewNopLogger returns a new logger that does nothing
func NewNopLogger() Logger {
	logger := zerolog.Nop()
	return ZeroLogWrapper{&logger}
}

// NewLogger returns a new logger
func NewLogger() Logger {
	output := zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.Kitchen}
	logger := zerolog.New(output).With().Timestamp().Logger()
	return ZeroLogWrapper{&logger}
}

// NewLoggerWithKV returns a new logger with the given key/value pair
func NewLoggerWithKV(key, value string) Logger {
	output := zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.Kitchen}
	logger := zerolog.New(output).With().Str(key, value).Timestamp().Logger()
	return ZeroLogWrapper{&logger}
}

// NewCustomLogger returns a new logger with the given zerolog logger
func NewCustomLogger(logger zerolog.Logger) Logger {
	return ZeroLogWrapper{&logger}
}

// Info takes a message and a set of key/value pairs and logs with level INFO.
// The key of the tuple must be a string.
func (l ZeroLogWrapper) Info(msg string, keyVals ...interface{}) {
	l.Logger.Info().Fields(keyVals).Msg(msg)
}

// Error takes a message and a set of key/value pairs and logs with level DEBUG.
// The key of the tuple must be a string.
func (l ZeroLogWrapper) Error(msg string, keyVals ...interface{}) {
	l.Logger.Error().Fields(keyVals).Msg(msg)
}

// Debug takes a message and a set of key/value pairs and logs with level ERR.
// The key of the tuple must be a string.
func (l ZeroLogWrapper) Debug(msg string, keyVals ...interface{}) {
	l.Logger.Debug().Fields(keyVals).Msg(msg)
}

// With returns a new wrapped logger with additional context provided by a set
func (l ZeroLogWrapper) With(keyVals ...interface{}) Logger {
	logger := l.Logger.With().Fields(keyVals).Logger()
	return ZeroLogWrapper{&logger}
}

// Impl returns the underlying zerolog logger
// It can be used to used zerolog structured API directly instead of the wrapper
func (l ZeroLogWrapper) Impl() interface{} {
	return l.Logger
}
