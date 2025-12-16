package log

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"strings"
	"time"

	"go.opentelemetry.io/contrib/bridges/otelslog"
)

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

	// InfoContext takes a context, message and key/value pairs and logs with level INFO.
	// The context is used for trace/span correlation when using OpenTelemetry.
	InfoContext(ctx context.Context, msg string, keyVals ...any)

	// WarnContext takes a context, message and key/value pairs and logs with level WARN.
	// The context is used for trace/span correlation when using OpenTelemetry.
	WarnContext(ctx context.Context, msg string, keyVals ...any)

	// ErrorContext takes a context, message and key/value pairs and logs with level ERROR.
	// The context is used for trace/span correlation when using OpenTelemetry.
	ErrorContext(ctx context.Context, msg string, keyVals ...any)

	// DebugContext takes a context, message and key/value pairs and logs with level DEBUG.
	// The context is used for trace/span correlation when using OpenTelemetry.
	DebugContext(ctx context.Context, msg string, keyVals ...any)

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

// slogLogger satisfies Logger with logging backed by an instance of *slog.Logger.
type slogLogger struct {
	log *slog.Logger
}

var _ Logger = slogLogger{}

// verboseModeLogger wraps a slogLogger and adds verbose mode support.
type verboseModeLogger struct {
	slogLogger
	handler *verboseModeHandler
}

var _ VerboseModeLogger = &verboseModeLogger{}

func (l *verboseModeLogger) SetVerboseMode(enable bool) {
	l.handler.SetVerboseMode(enable)
}

func (l *verboseModeLogger) With(keyVals ...any) Logger {
	// Return a new verboseModeLogger that shares the same handler
	return &verboseModeLogger{
		slogLogger: slogLogger{log: l.log.With(keyVals...)},
		handler:    l.handler,
	}
}

// NewLogger creates a Logger that exports logs to both console and OpenTelemetry.
// The name identifies the instrumentation scope (e.g., "cosmos-sdk").
//
// By default, logs are written to os.Stderr and exported to OpenTelemetry using
// the global LoggerProvider set by telemetry.InitializeOpenTelemetry().
//
// Use WithConsoleWriter to redirect console output to a different writer.
// Use WithoutConsole to disable console output entirely (OTEL-only).
// Use WithLoggerProvider to override with a custom OTEL provider.
//
// Example:
//
//	// Default: console (stderr) + OTEL
//	logger := log.NewLogger("cosmos-sdk")
//
//	// Custom console writer + OTEL
//	logger := log.NewLogger("cosmos-sdk", log.WithConsoleWriter(os.Stdout))
//
//	// OTEL-only (no console output)
//	logger := log.NewLogger("cosmos-sdk", log.WithoutConsole())
func NewLogger(name string, opts ...Option) Logger {
	cfg := &Config{
		Level:        slog.LevelInfo,
		VerboseLevel: NoLevel, // disabled by default
	}
	for _, opt := range opts {
		opt(cfg)
	}

	// Build otelslog options
	var otelOpts []otelslog.Option
	if cfg.LoggerProvider != nil {
		otelOpts = append(otelOpts, otelslog.WithLoggerProvider(cfg.LoggerProvider))
	}

	otelHandler := otelslog.NewHandler(name, otelOpts...)

	var handler slog.Handler

	// Determine output configuration
	if cfg.DisableConsole {
		// OTEL-only
		handler = otelHandler
	} else {
		// Default: console + OTEL
		consoleWriter := cfg.ConsoleWriter
		if consoleWriter == nil {
			consoleWriter = os.Stderr
		}

		var consoleHandler slog.Handler
		if cfg.OutputJSON {
			consoleHandler = slog.NewJSONHandler(consoleWriter, &slog.HandlerOptions{Level: cfg.Level})
		} else {
			consoleHandler = &prettyHandler{
				out:        consoleWriter,
				level:      cfg.Level,
				noColor:    !cfg.Color,
				timeFormat: cfg.TimeFormat,
			}
		}
		handler = &multiHandler{handlers: []slog.Handler{consoleHandler, otelHandler}}
	}

	// If verbose level is configured, use verboseModeHandler for dynamic level/filter switching
	if cfg.VerboseLevel != NoLevel {
		vmHandler := newVerboseModeHandler(handler, cfg.Level, cfg.VerboseLevel, cfg.Filter)
		return &verboseModeLogger{
			slogLogger: slogLogger{log: slog.New(vmHandler)},
			handler:    vmHandler,
		}
	}

	// Apply filter if configured (non-verbose mode)
	if cfg.Filter != nil {
		handler = newFilterHandler(handler, cfg.Filter)
	}

	return slogLogger{log: slog.New(handler)}
}

// NewCustomLogger returns a Logger backed by an existing slog.Logger instance.
// All logging methods are called directly on the *slog.Logger;
// therefore it is the caller's responsibility to configure message filtering,
// level filtering, output format, and so on.
func NewCustomLogger(log *slog.Logger) Logger {
	return slogLogger{log: log}
}

func (l slogLogger) Info(msg string, keyVals ...any) {
	l.log.Info(msg, keyVals...)
}

func (l slogLogger) Warn(msg string, keyVals ...any) {
	l.log.Warn(msg, keyVals...)
}

func (l slogLogger) Error(msg string, keyVals ...any) {
	l.log.Error(msg, keyVals...)
}

func (l slogLogger) Debug(msg string, keyVals ...any) {
	l.log.Debug(msg, keyVals...)
}

func (l slogLogger) InfoContext(ctx context.Context, msg string, keyVals ...any) {
	l.log.InfoContext(ctx, msg, keyVals...)
}

func (l slogLogger) WarnContext(ctx context.Context, msg string, keyVals ...any) {
	l.log.WarnContext(ctx, msg, keyVals...)
}

func (l slogLogger) ErrorContext(ctx context.Context, msg string, keyVals ...any) {
	l.log.ErrorContext(ctx, msg, keyVals...)
}

func (l slogLogger) DebugContext(ctx context.Context, msg string, keyVals ...any) {
	l.log.DebugContext(ctx, msg, keyVals...)
}

func (l slogLogger) With(keyVals ...any) Logger {
	return slogLogger{log: l.log.With(keyVals...)}
}

// Impl returns l's underlying [*slog.Logger].
func (l slogLogger) Impl() any {
	return l.log
}

// multiHandler fans out log records to multiple slog.Handlers.
// It uses best-effort semantics: all handlers are attempted even if some fail.
type multiHandler struct {
	handlers []slog.Handler
}

func (h *multiHandler) Enabled(ctx context.Context, level slog.Level) bool {
	for _, handler := range h.handlers {
		if handler.Enabled(ctx, level) {
			return true
		}
	}
	return false
}

func (h *multiHandler) Handle(ctx context.Context, r slog.Record) error {
	var errs []error
	for _, handler := range h.handlers {
		if handler.Enabled(ctx, r.Level) {
			if err := handler.Handle(ctx, r.Clone()); err != nil {
				errs = append(errs, err)
			}
		}
	}
	return errors.Join(errs...)
}

func (h *multiHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	newHandlers := make([]slog.Handler, len(h.handlers))
	for i, handler := range h.handlers {
		newHandlers[i] = handler.WithAttrs(attrs)
	}
	return &multiHandler{handlers: newHandlers}
}

func (h *multiHandler) WithGroup(name string) slog.Handler {
	newHandlers := make([]slog.Handler, len(h.handlers))
	for i, handler := range h.handlers {
		newHandlers[i] = handler.WithGroup(name)
	}
	return &multiHandler{handlers: newHandlers}
}

// ANSI color codes
const (
	colorReset  = "\033[0m"
	colorRed    = "\033[31m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorBlue   = "\033[34m"
	colorCyan   = "\033[36m"
	colorGray   = "\033[90m"
)

// prettyHandler is a simple slog.Handler that outputs logs in a human-readable format
// similar to zerolog's console output.
type prettyHandler struct {
	out        io.Writer
	level      slog.Level
	attrs      []slog.Attr
	groups     []string
	noColor    bool
	timeFormat string
}

func (h *prettyHandler) Enabled(_ context.Context, level slog.Level) bool {
	return level >= h.level
}

func (h *prettyHandler) Handle(_ context.Context, r slog.Record) error {
	// Format: 3:04PM INF message key=value key=value module=xxx
	timeFormat := h.timeFormat
	if timeFormat == "" {
		timeFormat = time.Kitchen
	}
	timeStr := r.Time.Format(timeFormat)
	levelStr, levelColor := h.levelToShortAndColor(r.Level)

	// Start with time and level
	if h.noColor {
		fmt.Fprintf(h.out, "%s %s %s", timeStr, levelStr, r.Message)
	} else {
		fmt.Fprintf(h.out, "%s%s%s %s%s%s %s", colorGray, timeStr, colorReset, levelColor, levelStr, colorReset, r.Message)
	}

	// Add pre-set attrs (these already have group prefixes applied)
	for _, attr := range h.attrs {
		h.writeAttr(attr, "")
	}

	// Build group prefix for record attrs
	groupPrefix := h.buildGroupPrefix()

	// Add record attrs (skip complex objects like "impl" that contain nested loggers)
	r.Attrs(func(attr slog.Attr) bool {
		// Skip attributes that are likely to be huge internal state objects
		if attr.Key == "impl" {
			return true
		}
		h.writeAttr(attr, groupPrefix)
		return true
	})

	fmt.Fprintln(h.out)
	return nil
}

func (h *prettyHandler) buildGroupPrefix() string {
	if len(h.groups) == 0 {
		return ""
	}
	return strings.Join(h.groups, ".") + "."
}

func (h *prettyHandler) writeAttr(attr slog.Attr, prefix string) {
	val := attr.Value.Resolve()

	// Skip empty values
	if val.Kind() == slog.KindAny && val.Any() == nil {
		return
	}

	key := prefix + attr.Key

	// Print key
	if h.noColor {
		fmt.Fprintf(h.out, " %s=", key)
	} else {
		fmt.Fprintf(h.out, " %s%s=%s", colorCyan, key, colorReset)
	}

	// For simple types, print value
	switch val.Kind() {
	case slog.KindString:
		fmt.Fprint(h.out, val.String())
	case slog.KindInt64:
		fmt.Fprintf(h.out, "%d", val.Int64())
	case slog.KindUint64:
		fmt.Fprintf(h.out, "%d", val.Uint64())
	case slog.KindFloat64:
		fmt.Fprintf(h.out, "%g", val.Float64())
	case slog.KindBool:
		fmt.Fprintf(h.out, "%t", val.Bool())
	case slog.KindDuration:
		fmt.Fprint(h.out, val.Duration())
	case slog.KindTime:
		fmt.Fprint(h.out, val.Time().Format(time.RFC3339))
	case slog.KindGroup:
		// For groups, print nested attrs with the group name as prefix
		groupPrefix := key + "."
		for _, a := range val.Group() {
			h.writeAttr(a, groupPrefix)
		}
	default:
		// For KindAny, check if it implements Stringer or can be formatted
		if v := val.Any(); v != nil {
			if s, ok := v.(fmt.Stringer); ok {
				fmt.Fprint(h.out, s.String())
			} else {
				fmt.Fprintf(h.out, "%v", v)
			}
		}
	}
}

func (h *prettyHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	// Apply current group prefix to new attrs
	groupPrefix := h.buildGroupPrefix()
	prefixedAttrs := make([]slog.Attr, len(attrs))
	for i, attr := range attrs {
		prefixedAttrs[i] = slog.Attr{Key: groupPrefix + attr.Key, Value: attr.Value}
	}
	return &prettyHandler{
		out:        h.out,
		level:      h.level,
		attrs:      append(h.attrs, prefixedAttrs...),
		groups:     h.groups,
		noColor:    h.noColor,
		timeFormat: h.timeFormat,
	}
}

func (h *prettyHandler) WithGroup(name string) slog.Handler {
	return &prettyHandler{
		out:        h.out,
		level:      h.level,
		attrs:      h.attrs,
		groups:     append(h.groups, name),
		noColor:    h.noColor,
		timeFormat: h.timeFormat,
	}
}

func (h *prettyHandler) levelToShortAndColor(level slog.Level) (string, string) {
	if h.noColor {
		switch {
		case level < slog.LevelInfo:
			return "DBG", ""
		case level < slog.LevelWarn:
			return "INF", ""
		case level < slog.LevelError:
			return "WRN", ""
		default:
			return "ERR", ""
		}
	}
	switch {
	case level < slog.LevelInfo:
		return "DBG", colorBlue
	case level < slog.LevelWarn:
		return "INF", colorGreen
	case level < slog.LevelError:
		return "WRN", colorYellow
	default:
		return "ERR", colorRed
	}
}

// NewNopLogger returns a new logger that does nothing.
func NewNopLogger() Logger {
	// The custom nopLogger is about 3x faster than a slogLogger with a discard handler.
	return nopLogger{}
}

// nopLogger is a Logger that does nothing when called.
type nopLogger struct{}

func (nopLogger) Info(string, ...any)                          {}
func (nopLogger) Warn(string, ...any)                          {}
func (nopLogger) Error(string, ...any)                         {}
func (nopLogger) Debug(string, ...any)                         {}
func (nopLogger) InfoContext(context.Context, string, ...any)  {}
func (nopLogger) WarnContext(context.Context, string, ...any)  {}
func (nopLogger) ErrorContext(context.Context, string, ...any) {}
func (nopLogger) DebugContext(context.Context, string, ...any) {}
func (nopLogger) With(...any) Logger                           { return nopLogger{} }
func (nopLogger) Impl() any                                    { return nopLogger{} }
