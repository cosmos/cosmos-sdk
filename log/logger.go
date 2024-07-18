package log

import (
	"encoding"
	"encoding/json"
	"fmt"
	"io"

	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/pkgerrors"

	corelog "cosmossdk.io/core/log"
	coretesting "cosmossdk.io/core/testing"
)

func init() {
	zerolog.InterfaceMarshalFunc = func(i any) ([]byte, error) {
		switch v := i.(type) {
		case json.Marshaler:
			return json.Marshal(i)
		case encoding.TextMarshaler:
			return json.Marshal(i)
		case fmt.Stringer:
			return json.Marshal(v.String())
		default:
			return json.Marshal(i)
		}
	}
}

// ModuleKey defines a module logging key.
const ModuleKey = corelog.ModuleKey

// ContextKey is used to store the logger in the context.
var ContextKey struct{}

// Logger is the Cosmos SDK logger interface.
// It maintains as much backward compatibility with the CometBFT logger as possible.
// All functionalities of the logger are available through the Impl() method.
type Logger = corelog.Logger

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

	if logCfg.Filter != nil {
		output = NewFilterWriter(output, logCfg.Filter)
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

	if logCfg.Level != zerolog.NoLevel {
		logger = logger.Level(logCfg.Level)
	}

	logger = logger.Hook(logCfg.Hooks...)

	return zeroLogWrapper{&logger}
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
	return zeroLogWrapper{&logger}
}

// Impl returns the underlying zerolog logger.
// It can be used to used zerolog structured API directly instead of the wrapper.
func (l zeroLogWrapper) Impl() interface{} {
	return l.Logger
}

// NewNopLogger returns a new logger that does nothing.
var NewNopLogger = coretesting.NewNopLogger

// LogWrapper wraps a Logger and implements the Logger interface.
// it is only meant to avoid breakage of legacy versions of the Logger interface.
type LogWrapper struct {
	corelog.Logger
}

func NewLogWrapper(logger corelog.Logger) Logger {
	return LogWrapper{logger}
}

func (l LogWrapper) Impl() interface{} {
	return l.Logger
}

func (l LogWrapper) With(keyVals ...interface{}) Logger {
	return NewLogWrapper(l.Logger.With(keyVals...))
}

func (l LogWrapper) Info(msg string, keyVals ...interface{}) {
	l.Logger.Info(msg, keyVals...)
}

func (l LogWrapper) Warn(msg string, keyVals ...interface{}) {
	l.Logger.Warn(msg, keyVals...)
}

func (l LogWrapper) Error(msg string, keyVals ...interface{}) {
	l.Logger.Error(msg, keyVals...)
}

func (l LogWrapper) Debug(msg string, keyVals ...interface{}) {
	l.Logger.Debug(msg, keyVals...)
}
