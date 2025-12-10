package slog

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"time"

	"go.opentelemetry.io/contrib/bridges/otelslog"
	otellog "go.opentelemetry.io/otel/log"

	"cosmossdk.io/log"
)

// OtelLoggerOption configures an OpenTelemetry logger.
type OtelLoggerOption func(*otelLoggerConfig)

type otelLoggerConfig struct {
	loggerProvider otellog.LoggerProvider
	consoleWriter  io.Writer
	consoleHandler slog.Handler
	level          slog.Level
	jsonOutput     bool
}

// WithLoggerProvider sets a custom OpenTelemetry LoggerProvider.
// If not provided, the global LoggerProvider is used.
func WithLoggerProvider(provider otellog.LoggerProvider) OtelLoggerOption {
	return func(cfg *otelLoggerConfig) {
		cfg.loggerProvider = provider
	}
}

// WithConsoleWriter sets a writer for console output in addition to OTEL export.
// This enables dual output: logs go to both the console and OpenTelemetry.
func WithConsoleWriter(w io.Writer) OtelLoggerOption {
	return func(cfg *otelLoggerConfig) {
		cfg.consoleWriter = w
	}
}

// WithConsoleHandler sets a custom slog.Handler for console output.
// This takes precedence over WithConsoleWriter if both are set.
func WithConsoleHandler(h slog.Handler) OtelLoggerOption {
	return func(cfg *otelLoggerConfig) {
		cfg.consoleHandler = h
	}
}

// WithLevel sets the minimum log level. Defaults to slog.LevelInfo.
func WithLevel(level slog.Level) OtelLoggerOption {
	return func(cfg *otelLoggerConfig) {
		cfg.level = level
	}
}

// WithJSONOutput configures the console handler to output JSON instead of text.
func WithJSONOutput() OtelLoggerOption {
	return func(cfg *otelLoggerConfig) {
		cfg.jsonOutput = true
	}
}

// NewOtelLogger creates a Logger that exports logs to OpenTelemetry.
// The name identifies the instrumentation scope (e.g., "cosmos-sdk").
//
// By default, it uses the global LoggerProvider set by telemetry.InitializeOpenTelemetry().
// Use WithLoggerProvider to override with a custom provider.
//
// To enable dual output (console + OTEL), use WithConsoleWriter or WithConsoleHandler.
//
// Example:
//
//	// OTEL-only logging
//	logger := slog.NewOtelLogger("cosmos-sdk")
//
//	// Dual output to stderr and OTEL
//	logger := slog.NewOtelLogger("cosmos-sdk", slog.WithConsoleWriter(os.Stderr))
func NewOtelLogger(name string, opts ...OtelLoggerOption) log.Logger {
	cfg := &otelLoggerConfig{
		level: slog.LevelInfo,
	}
	for _, opt := range opts {
		opt(cfg)
	}

	// Build otelslog options
	var otelOpts []otelslog.Option
	if cfg.loggerProvider != nil {
		otelOpts = append(otelOpts, otelslog.WithLoggerProvider(cfg.loggerProvider))
	}

	otelHandler := otelslog.NewHandler(name, otelOpts...)

	var handler slog.Handler

	// Determine if we need dual output
	if cfg.consoleHandler != nil {
		// Use provided console handler + OTEL handler
		handler = &multiHandler{handlers: []slog.Handler{cfg.consoleHandler, otelHandler}}
	} else if cfg.consoleWriter != nil {
		// Create a handler for console output
		handlerOpts := &slog.HandlerOptions{Level: cfg.level}
		var consoleHandler slog.Handler
		if cfg.jsonOutput {
			consoleHandler = slog.NewJSONHandler(cfg.consoleWriter, handlerOpts)
		} else {
			consoleHandler = &prettyHandler{
				out:   cfg.consoleWriter,
				level: cfg.level,
			}
		}
		handler = &multiHandler{handlers: []slog.Handler{consoleHandler, otelHandler}}
	} else {
		// OTEL-only
		handler = otelHandler
	}

	return Logger{log: slog.New(handler)}
}

// multiHandler fans out log records to multiple slog.Handlers.
type multiHandler struct {
	handlers []slog.Handler
	attrs    []slog.Attr
	group    string
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
	for _, handler := range h.handlers {
		if handler.Enabled(ctx, r.Level) {
			if err := handler.Handle(ctx, r.Clone()); err != nil {
				return err
			}
		}
	}
	return nil
}

func (h *multiHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	newHandlers := make([]slog.Handler, len(h.handlers))
	for i, handler := range h.handlers {
		newHandlers[i] = handler.WithAttrs(attrs)
	}
	return &multiHandler{
		handlers: newHandlers,
		attrs:    append(h.attrs, attrs...),
		group:    h.group,
	}
}

func (h *multiHandler) WithGroup(name string) slog.Handler {
	newHandlers := make([]slog.Handler, len(h.handlers))
	for i, handler := range h.handlers {
		newHandlers[i] = handler.WithGroup(name)
	}
	return &multiHandler{
		handlers: newHandlers,
		attrs:    h.attrs,
		group:    name,
	}
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
	out    io.Writer
	level  slog.Level
	attrs  []slog.Attr
	groups []string
}

func (h *prettyHandler) Enabled(_ context.Context, level slog.Level) bool {
	return level >= h.level
}

func (h *prettyHandler) Handle(_ context.Context, r slog.Record) error {
	// Format: 3:04PM INF message key=value key=value module=xxx
	timeStr := r.Time.Format(time.Kitchen)
	levelStr, levelColor := levelToShortAndColor(r.Level)

	// Start with time and level
	fmt.Fprintf(h.out, "%s%s%s %s%s%s %s", colorGray, timeStr, colorReset, levelColor, levelStr, colorReset, r.Message)

	// Add pre-set attrs
	for _, attr := range h.attrs {
		h.writeAttr(attr)
	}

	// Add record attrs (skip complex objects like "impl" that contain nested loggers)
	r.Attrs(func(attr slog.Attr) bool {
		// Skip attributes that are likely to be huge internal state objects
		if attr.Key == "impl" {
			return true
		}
		h.writeAttr(attr)
		return true
	})

	fmt.Fprintln(h.out)
	return nil
}

func (h *prettyHandler) writeAttr(attr slog.Attr) {
	val := attr.Value.Resolve()

	// Skip empty values
	if val.Kind() == slog.KindAny && val.Any() == nil {
		return
	}

	// Print key in cyan
	fmt.Fprintf(h.out, " %s%s=%s", colorCyan, attr.Key, colorReset)

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
		// For groups, print nested attrs
		for _, a := range val.Group() {
			h.writeAttr(a)
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
	return &prettyHandler{
		out:    h.out,
		level:  h.level,
		attrs:  append(h.attrs, attrs...),
		groups: h.groups,
	}
}

func (h *prettyHandler) WithGroup(name string) slog.Handler {
	return &prettyHandler{
		out:    h.out,
		level:  h.level,
		attrs:  h.attrs,
		groups: append(h.groups, name),
	}
}

func levelToShortAndColor(level slog.Level) (string, string) {
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
