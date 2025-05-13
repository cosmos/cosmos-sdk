package log

import (
	"encoding"
	"encoding/json"
	"fmt"
	"io"

	"github.com/bytedance/sonic"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/pkgerrors"
)

func init() {
	zerolog.InterfaceMarshalFunc = func(i any) ([]byte, error) {
		switch v := i.(type) {
		case json.Marshaler:
			return sonic.Marshal(i)
		case encoding.TextMarshaler:
			return sonic.Marshal(i)
		case fmt.Stringer:
			return sonic.Marshal(v.String())
		default:
			return sonic.Marshal(i)
		}
	}
}

// ModuleKey defines a module logging key.
const ModuleKey = "module"

// ContextKey is used to store the logger in the context.
var ContextKey contextKey

type contextKey struct{}

// Logger is the Cosmos SDK logger interface.
type Logger interface {
	// Info takes a message and a set of key/value pairs and logs with level INFO.
	// The key of the tuple must be a string.
	Info(msg string, keyVals ...any)

	// Warn takes a message and a set of key/value pairs and logs with level WARN.
	// The key of the tuple must be a string.
	Warn(msg string, keyVals ...any)

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

// VerboseModeLogger is an extension interface of Logger which allows verbosity to be configured.
type VerboseModeLogger interface {
	Logger
	// SetVerboseMode configures whether the logger enters verbose mode or not for
	// special operations where increased observability of log messages is desired
	// (such as chain upgrades).
	SetVerboseMode(bool)
}

// WithJSONMarshal configures zerolog global json encoding.
func WithJSONMarshal(marshaler func(v any) ([]byte, error)) {
	zerolog.InterfaceMarshalFunc = func(i any) ([]byte, error) {
		switch v := i.(type) {
		case json.Marshaler:
			return marshaler(i)
		case encoding.TextMarshaler:
			return marshaler(i)
		case fmt.Stringer:
			return marshaler(v.String())
		default:
			return marshaler(i)
		}
	}
}

type zeroLogWrapper struct {
	*zerolog.Logger
	regularLevel zerolog.Level
	verboseLevel zerolog.Level
	// this field is used to disable filtering during verbose logging
	// and will only be non-nil when we have a filterWriter
	filterWriter *filterWriter
}

// NewLogger returns a new logger that writes to the given destination.
//
// Typical usage from a main function is:
//
//	logger := log.NewLogger(os.Stderr)
//
// Stderr is the typical destination for logs,
// so that any output from your application can still be piped to other processes.
func NewLogger(dst io.Writer, options ...Option) Logger {
	logCfg := defaultConfig
	for _, opt := range options {
		opt(&logCfg)
	}

	output := dst
	if !logCfg.OutputJSON {
		output = zerolog.ConsoleWriter{
			Out:        dst,
			NoColor:    !logCfg.Color,
			TimeFormat: logCfg.TimeFormat,
		}
	}

	var fltWtr *filterWriter
	if logCfg.Filter != nil {
		fltWtr = &filterWriter{
			parent: output,
			filter: logCfg.Filter,
		}
		output = fltWtr
	}

	logger := zerolog.New(output)

	if logCfg.StackTrace {
		zerolog.ErrorStackMarshaler = func(err error) interface{} {
			return pkgerrors.MarshalStack(errors.WithStack(err))
		}

		logger = logger.With().Stack().Logger()
	}

	if logCfg.TimeFormat != "" {
		logger = logger.With().Timestamp().Logger()
	}

	logger = logger.Level(logCfg.Level)
	logger = logger.Hook(logCfg.Hooks...)

	return zeroLogWrapper{
		Logger:       &logger,
		regularLevel: logCfg.Level,
		verboseLevel: logCfg.VerboseLevel,
		filterWriter: fltWtr,
	}
}

// NewCustomLogger returns a new logger with the given zerolog logger.
func NewCustomLogger(logger zerolog.Logger) Logger {
	return zeroLogWrapper{
		Logger:       &logger,
		regularLevel: logger.GetLevel(),
		verboseLevel: zerolog.NoLevel,
		filterWriter: nil,
	}
}

// Info takes a message and a set of key/value pairs and logs with level INFO.
// The key of the tuple must be a string.
func (l zeroLogWrapper) Info(msg string, keyVals ...interface{}) {
	l.Logger.Info().Fields(keyVals).Msg(msg)
}

// Warn takes a message and a set of key/value pairs and logs with level WARN.
// The key of the tuple must be a string.
func (l zeroLogWrapper) Warn(msg string, keyVals ...interface{}) {
	l.Logger.Warn().Fields(keyVals).Msg(msg)
}

// Error takes a message and a set of key/value pairs and logs with level ERROR.
// The key of the tuple must be a string.
func (l zeroLogWrapper) Error(msg string, keyVals ...interface{}) {
	l.Logger.Error().Fields(keyVals).Msg(msg)
}

// Debug takes a message and a set of key/value pairs and logs with level DEBUG.
// The key of the tuple must be a string.
func (l zeroLogWrapper) Debug(msg string, keyVals ...interface{}) {
	l.Logger.Debug().Fields(keyVals).Msg(msg)
}

// With returns a new wrapped logger with additional context provided by a set.
func (l zeroLogWrapper) With(keyVals ...interface{}) Logger {
	logger := l.Logger.With().Fields(keyVals).Logger()
	l.Logger = &logger
	return l
}

// WithContext returns a new wrapped logger with additional context provided by a set.
func (l zeroLogWrapper) WithContext(keyVals ...interface{}) any {
	logger := l.Logger.With().Fields(keyVals).Logger()
	l.Logger = &logger
	return l
}

// Impl returns the underlying zerolog logger.
// It can be used to used zerolog structured API directly instead of the wrapper.
func (l zeroLogWrapper) Impl() interface{} {
	return l.Logger
}

// SetVerboseMode implements VerboseModeLogger interface.
func (l zeroLogWrapper) SetVerboseMode(enable bool) {
	if enable && l.verboseLevel != zerolog.NoLevel {
		*l.Logger = l.Level(l.verboseLevel)
		if l.filterWriter != nil {
			l.filterWriter.disableFilter = true
		}
	} else {
		*l.Logger = l.Level(l.regularLevel)
		if l.filterWriter != nil {
			l.filterWriter.disableFilter = false
		}
	}
}

var _ VerboseModeLogger = zeroLogWrapper{}

// NewNopLogger returns a new logger that does nothing.
func NewNopLogger() Logger {
	// The custom nopLogger is about 3x faster than a zeroLogWrapper with zerolog.Nop().
	return nopLogger{}
}

// nopLogger is a Logger that does nothing when called.
// See the "specialized nop logger" benchmark and compare with the "zerolog nop logger" benchmark.
// The custom implementation is about 3x faster.
type nopLogger struct{}

func (nopLogger) Info(string, ...any)    {}
func (nopLogger) Warn(string, ...any)    {}
func (nopLogger) Error(string, ...any)   {}
func (nopLogger) Debug(string, ...any)   {}
func (nopLogger) With(...any) Logger     { return nopLogger{} }
func (nopLogger) WithContext(...any) any { return nopLogger{} }
func (nopLogger) Impl() any              { return nopLogger{} }
