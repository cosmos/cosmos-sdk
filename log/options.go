package log

import (
	"io"
	"log/slog"

	otellog "go.opentelemetry.io/otel/log"
)

// NoLevel indicates that verbose mode should not change logging behavior.
const NoLevel slog.Level = -100

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
	// LoggerProvider is the OpenTelemetry logger provider to use.
	// If nil, the global logger provider is used.
	LoggerProvider otellog.LoggerProvider
	// ConsoleWriter is the writer for console output.
	// If nil and DisableConsole is false, defaults to os.Stderr.
	ConsoleWriter io.Writer
	// ConsoleHandler is a custom slog.Handler for console output.
	// Takes precedence over ConsoleWriter if both are set.
	ConsoleHandler slog.Handler
	// DisableConsole disables console output entirely.
	// When true, logs are only sent to OpenTelemetry.
	DisableConsole bool
	// OutputJSON configures the console handler to output JSON instead of text.
	OutputJSON bool
	// Color enables/disables ANSI color codes in console output.
	Color bool
	// TimeFormat is the time format string for console output.
	// If empty, defaults to time.Kitchen ("3:04PM").
	TimeFormat string
	// Filter is an optional filter function to selectively discard logs.
	Filter FilterFunc
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

// WithLoggerProvider sets a custom OpenTelemetry LoggerProvider.
// If not provided, the global LoggerProvider is used.
func WithLoggerProvider(provider otellog.LoggerProvider) Option {
	return func(cfg *Config) {
		cfg.LoggerProvider = provider
	}
}

// WithConsoleWriter overrides the default console writer (os.Stderr).
// By default, logs are written to both console and OpenTelemetry.
// Use this to redirect console output to a different writer.
func WithConsoleWriter(w io.Writer) Option {
	return func(cfg *Config) {
		cfg.ConsoleWriter = w
	}
}

// WithoutConsole disables console output entirely.
// When enabled, logs are only sent to OpenTelemetry.
func WithoutConsole() Option {
	return func(cfg *Config) {
		cfg.DisableConsole = true
	}
}

// WithConsoleHandler sets a custom slog.Handler for console output.
// This takes precedence over WithConsoleWriter if both are set.
func WithConsoleHandler(h slog.Handler) Option {
	return func(cfg *Config) {
		cfg.ConsoleHandler = h
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

// WithFilter sets the filter for the Logger.
// The filter function is called with the module and level of each log entry.
// If the filter returns true, the log entry is discarded.
func WithFilter(filter FilterFunc) Option {
	return func(cfg *Config) {
		cfg.Filter = filter
	}
}
