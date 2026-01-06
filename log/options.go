package log

import (
	"io"
	"log/slog"
	"time"

	"github.com/rs/zerolog"
	otellog "go.opentelemetry.io/otel/log"
)

// NoLevel indicates that verbose mode should not change logging behavior.
const NoLevel slog.Level = -100

// defaultConfig has sensible defaults for the logger.
var defaultConfig = Config{
	Level:        slog.LevelInfo,
	VerboseLevel: NoLevel, // disabled by default
	Color:        true,
	TimeFormat:   time.Kitchen,
}

// Config defines configuration for the logger.
type Config struct {
	// Level is the minimum log level. Messages with a lower level will be discarded.
	Level slog.Level
	// VerboseLevel is the logging level to use when verbose mode is enabled.
	// If there is a filter enabled, it will be disabled when verbose mode is enabled
	// and all log messages will be emitted at the VerboseLevel.
	// If this is set to NoLevel, then no changes to the logging level or filter will be made
	// when verbose mode is enabled.
	VerboseLevel slog.Level
	// Filter is an optional filter function to selectively discard logs.
	Filter FilterFunc
	// OutputJSON configures the console handler to output JSON instead of text.
	OutputJSON bool
	// Color enables/disables ANSI color codes in console output.
	Color bool
	// TimeFormat is the time format string for console output.
	// If empty, defaults to time.Kitchen ("3:04PM").
	TimeFormat string
	// StackTrace enables stack trace logging on error.
	StackTrace bool
	// Hooks are zerolog hooks to add to the logger.
	Hooks []zerolog.Hook

	// ConsoleWriter is the writer for console output.
	// If nil and DisableConsole is false, defaults to os.Stderr.
	ConsoleWriter io.Writer
	// DisableConsole disables console output entirely.
	DisableConsole bool

	// LoggerProvider is the OpenTelemetry logger provider to use.
	// If nil, the global logger provider is used.
	LoggerProvider otellog.LoggerProvider
	// EnableOTEL controls OpenTelemetry log forwarding.
	// nil = check LoggerProvider.
	// true = force enable
	// false = force disable
	EnableOTEL *bool
}

// Option configures a Logger.
type Option func(*Config)

// WithLevel sets the minimum log level.
// Messages with a lower level will be discarded.
func WithLevel(level slog.Level) Option {
	return func(cfg *Config) {
		cfg.Level = level
	}
}

// WithVerboseLevel sets the verbose level for the Logger.
// When verbose mode is enabled via SetVerboseMode(true), the logger will be switched to this level
// and any filters will be disabled.
// Set to NoLevel to disable verbose mode changes entirely.
func WithVerboseLevel(level slog.Level) Option {
	return func(cfg *Config) {
		cfg.VerboseLevel = level
	}
}

// WithFilter sets the filter for the Logger.
// The filter function is called with the module and level of each log entry.
// If the filter returns true, the log entry is discarded.
func WithFilter(filter FilterFunc) Option {
	return func(cfg *Config) {
		cfg.Filter = filter
	}
}

// WithJSONOutput configures the console handler to output JSON instead of text.
func WithJSONOutput() Option {
	return func(cfg *Config) {
		cfg.OutputJSON = true
	}
}

// WithColor enables/disables ANSI color codes in console output.
// Defaults to true (colors enabled).
func WithColor(enabled bool) Option {
	return func(cfg *Config) {
		cfg.Color = enabled
	}
}

// WithTimeFormat sets the time format string for console output.
// The format uses Go's time layout format (e.g., time.Kitchen, time.RFC3339).
// If not set, defaults to time.Kitchen ("3:04PM").
func WithTimeFormat(format string) Option {
	return func(cfg *Config) {
		cfg.TimeFormat = format
	}
}

// WithStackTrace enables stack trace logging on error.
func WithStackTrace(enabled bool) Option {
	return func(cfg *Config) {
		cfg.StackTrace = enabled
	}
}

// WithHooks appends hooks to the Logger hooks.
func WithHooks(hooks ...zerolog.Hook) Option {
	return func(cfg *Config) {
		cfg.Hooks = append(cfg.Hooks, hooks...)
	}
}

// WithConsoleWriter overrides the default console writer (os.Stderr).
// Use this to redirect console output to a different writer.
func WithConsoleWriter(w io.Writer) Option {
	return func(cfg *Config) {
		cfg.ConsoleWriter = w
	}
}

// WithoutConsole disables console output entirely.
// When enabled, logs are only sent to OpenTelemetry (if OTEL is enabled).
func WithoutConsole() Option {
	return func(cfg *Config) {
		cfg.DisableConsole = true
	}
}

// WithOTEL forces OpenTelemetry log forwarding to be enabled.
// By default, OTEL is auto-detected based on environment variables
// (OTEL_EXPORTER_OTLP_ENDPOINT, OTEL_LOGS_EXPORTER).
// Use this to explicitly enable OTEL regardless of environment.
func WithOTEL() Option {
	return func(cfg *Config) {
		t := true
		cfg.EnableOTEL = &t
	}
}

// WithoutOTEL forces OpenTelemetry log forwarding to be disabled.
// Use this to explicitly disable OTEL regardless of environment variables.
// This enables the fast zerolog path with zero allocations.
func WithoutOTEL() Option {
	return func(cfg *Config) {
		f := false
		cfg.EnableOTEL = &f
	}
}

// WithLoggerProvider sets a custom OpenTelemetry LoggerProvider.
// If not provided, the global LoggerProvider is used.
// This also forces OTEL forwarding to be enabled.
func WithLoggerProvider(provider otellog.LoggerProvider) Option {
	return func(cfg *Config) {
		cfg.LoggerProvider = provider
		t := true
		cfg.EnableOTEL = &t
	}
}
