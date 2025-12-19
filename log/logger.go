package log

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"sync/atomic"
	"time"

	"github.com/rs/zerolog"
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
	handler *zerologHandler
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
		Color:        true,    // colors enabled by default
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
			// Use zerolog for console formatting with proper Stringer support
			timeFormat := cfg.TimeFormat
			if timeFormat == "" {
				timeFormat = time.Kitchen
			}
			zConsole := zerolog.ConsoleWriter{
				Out:        consoleWriter,
				TimeFormat: timeFormat,
				NoColor:    !cfg.Color,
			}
			zLogger := zerolog.New(zConsole).With().Timestamp().Logger()

			// Determine verbose level (use NoLevel sentinel if not configured)
			verboseLevel := cfg.VerboseLevel
			if verboseLevel == NoLevel {
				verboseLevel = cfg.Level // fallback to regular level
			}

			zHandler := &zerologHandler{
				logger:       zLogger,
				regularLevel: cfg.Level,
				verboseLevel: verboseLevel,
				filter:       cfg.Filter,
				verbose:      &atomic.Bool{},
			}
			consoleHandler = zHandler

			// If verbose mode is configured, return a VerboseModeLogger
			if cfg.VerboseLevel != NoLevel {
				handler = &multiHandler{handlers: []slog.Handler{consoleHandler, otelHandler}}
				return &verboseModeLogger{
					slogLogger: slogLogger{log: slog.New(handler)},
					handler:    zHandler,
				}
			}
		}
		handler = &multiHandler{handlers: []slog.Handler{consoleHandler, otelHandler}}
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

// zerologHandler is a slog.Handler that uses zerolog for console output.
// It properly handles fmt.Stringer types by calling String() before logging.
// It also supports verbose mode switching and filtering.
type zerologHandler struct {
	logger       zerolog.Logger
	regularLevel slog.Level
	verboseLevel slog.Level
	filter       FilterFunc
	module       string       // current module from attrs (for filtering)
	verbose      *atomic.Bool // shared across all derived handlers
	attrs        []slog.Attr
	groups       []string
}

func (h *zerologHandler) Enabled(_ context.Context, level slog.Level) bool {
	if h.verbose != nil && h.verbose.Load() {
		return level >= h.verboseLevel
	}
	return level >= h.regularLevel
}

func (h *zerologHandler) Handle(_ context.Context, r slog.Record) error {
	// Check level first
	if h.verbose != nil && h.verbose.Load() {
		if r.Level < h.verboseLevel {
			return nil
		}
	} else {
		if r.Level < h.regularLevel {
			return nil
		}
	}

	// Apply filter only in non-verbose mode
	if h.filter != nil && (h.verbose == nil || !h.verbose.Load()) {
		module := h.module
		if module == "" {
			r.Attrs(func(attr slog.Attr) bool {
				if attr.Key == ModuleKey {
					module = attr.Value.String()
					return false
				}
				return true
			})
		}
		if h.filter(module, LevelToString(r.Level)) {
			return nil // filtered out
		}
	}

	var evt *zerolog.Event
	switch {
	case r.Level < slog.LevelInfo:
		evt = h.logger.Debug()
	case r.Level < slog.LevelWarn:
		evt = h.logger.Info()
	case r.Level < slog.LevelError:
		evt = h.logger.Warn()
	default:
		evt = h.logger.Error()
	}

	// Add pre-set attrs
	for _, attr := range h.attrs {
		evt = h.addAttr(evt, attr)
	}

	// Add record attrs
	r.Attrs(func(attr slog.Attr) bool {
		evt = h.addAttr(evt, attr)
		return true
	})

	evt.Msg(r.Message)
	return nil
}

// SetVerboseMode enables or disables verbose mode for this handler and all derived handlers.
func (h *zerologHandler) SetVerboseMode(enable bool) {
	if h.verbose != nil {
		h.verbose.Store(enable)
	}
}

func (h *zerologHandler) addAttr(evt *zerolog.Event, attr slog.Attr) *zerolog.Event {
	key := h.prefixKey(attr.Key)
	val := attr.Value.Resolve()

	switch val.Kind() {
	case slog.KindString:
		return evt.Str(key, val.String())
	case slog.KindInt64:
		return evt.Int64(key, val.Int64())
	case slog.KindUint64:
		return evt.Uint64(key, val.Uint64())
	case slog.KindFloat64:
		return evt.Float64(key, val.Float64())
	case slog.KindBool:
		return evt.Bool(key, val.Bool())
	case slog.KindDuration:
		return evt.Dur(key, val.Duration())
	case slog.KindTime:
		return evt.Time(key, val.Time())
	case slog.KindGroup:
		// Handle nested groups
		for _, a := range val.Group() {
			evt = h.addAttrWithPrefix(evt, a, key+".")
		}
		return evt
	default:
		// KindAny - check for Stringer first
		if v := val.Any(); v != nil {
			if s, ok := v.(fmt.Stringer); ok {
				return evt.Str(key, s.String())
			}
			return evt.Interface(key, v)
		}
		return evt
	}
}

func (h *zerologHandler) addAttrWithPrefix(evt *zerolog.Event, attr slog.Attr, prefix string) *zerolog.Event {
	key := prefix + attr.Key
	val := attr.Value.Resolve()

	switch val.Kind() {
	case slog.KindString:
		return evt.Str(key, val.String())
	case slog.KindInt64:
		return evt.Int64(key, val.Int64())
	case slog.KindUint64:
		return evt.Uint64(key, val.Uint64())
	case slog.KindFloat64:
		return evt.Float64(key, val.Float64())
	case slog.KindBool:
		return evt.Bool(key, val.Bool())
	case slog.KindDuration:
		return evt.Dur(key, val.Duration())
	case slog.KindTime:
		return evt.Time(key, val.Time())
	case slog.KindGroup:
		for _, a := range val.Group() {
			evt = h.addAttrWithPrefix(evt, a, key+".")
		}
		return evt
	default:
		if v := val.Any(); v != nil {
			if s, ok := v.(fmt.Stringer); ok {
				return evt.Str(key, s.String())
			}
			return evt.Interface(key, v)
		}
		return evt
	}
}

func (h *zerologHandler) prefixKey(key string) string {
	if len(h.groups) == 0 {
		return key
	}
	prefix := ""
	for _, g := range h.groups {
		prefix += g + "."
	}
	return prefix + key
}

func (h *zerologHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	// Check if module is being set
	newModule := h.module
	for _, attr := range attrs {
		if attr.Key == ModuleKey {
			newModule = attr.Value.String()
			break
		}
	}
	return &zerologHandler{
		logger:       h.logger,
		regularLevel: h.regularLevel,
		verboseLevel: h.verboseLevel,
		filter:       h.filter,
		module:       newModule,
		verbose:      h.verbose, // share the same atomic bool
		attrs:        append(h.attrs, attrs...),
		groups:       h.groups,
	}
}

func (h *zerologHandler) WithGroup(name string) slog.Handler {
	return &zerologHandler{
		logger:       h.logger,
		regularLevel: h.regularLevel,
		verboseLevel: h.verboseLevel,
		filter:       h.filter,
		module:       h.module,
		verbose:      h.verbose, // share the same atomic bool
		attrs:        h.attrs,
		groups:       append(h.groups, name),
	}
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
