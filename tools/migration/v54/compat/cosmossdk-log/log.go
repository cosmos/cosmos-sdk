package log

import (
	"io"

	logv2 "cosmossdk.io/log/v2"
	"github.com/rs/zerolog"
)

const ModuleKey = logv2.ModuleKey

var ContextKey = logv2.ContextKey

type Logger = logv2.Logger
type VerboseModeLogger = logv2.VerboseModeLogger
type Config = logv2.Config
type Option = logv2.Option
type FilterFunc = logv2.FilterFunc
type TestingT = logv2.TestingT

func WithJSONMarshal(marshaler func(v any) ([]byte, error)) {
	logv2.WithJSONMarshal(marshaler)
}

func FilterOption(filter FilterFunc) Option {
	return logv2.FilterOption(filter)
}

func LevelOption(level zerolog.Level) Option {
	return logv2.LevelOption(level)
}

func OutputJSONOption() Option {
	return logv2.OutputJSONOption()
}

func ColorOption(val bool) Option {
	return logv2.ColorOption(val)
}

func TimeFormatOption(format string) Option {
	return logv2.TimeFormatOption(format)
}

func TraceOption(val bool) Option {
	return logv2.TraceOption(val)
}

func HooksOption(hooks ...zerolog.Hook) Option {
	return logv2.HooksOption(hooks...)
}

func ParseLogLevel(levelStr string) (FilterFunc, error) {
	return logv2.ParseLogLevel(levelStr)
}

func NewFilterWriter(parent io.Writer, filter FilterFunc) io.Writer {
	return logv2.NewFilterWriter(parent, filter)
}

func NewLogger(dst io.Writer, options ...Option) Logger {
	return logv2.NewLogger(dst, options...)
}

func NewCustomLogger(logger zerolog.Logger) Logger {
	return logv2.NewCustomLogger(logger)
}

func NewNopLogger() Logger {
	return logv2.NewNopLogger()
}

func NewTestLogger(t TestingT) Logger {
	return logv2.NewTestLogger(t)
}

func NewTestLoggerInfo(t TestingT) Logger {
	return logv2.NewTestLoggerInfo(t)
}

func NewTestLoggerError(t TestingT) Logger {
	return logv2.NewTestLoggerError(t)
}
