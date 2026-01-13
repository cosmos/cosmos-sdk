package log2

import (
	"context"
	"io"
	"log/slog"
	"os"

	"go.opentelemetry.io/contrib/bridges/otelslog"
)

// Option configures New.
type Option func(*config)

type config struct {
	consoleWriter io.Writer
	consoleJSON   bool
	consoleOpts   *slog.HandlerOptions
	otelOpts      otelslog.Option
}

// WithConsoleWriter sets the destination for console logs (defaults to os.Stdout).
func WithConsoleWriter(w io.Writer) Option {
	return func(c *config) { c.consoleWriter = w }
}

// WithConsoleJSON makes console logs JSON instead of text.
func WithConsoleJSON(enabled bool) Option {
	return func(c *config) { c.consoleJSON = enabled }
}

// WithConsoleHandlerOptions sets slog HandlerOptions for the console handler.
func WithConsoleHandlerOptions(opts *slog.HandlerOptions) Option {
	return func(c *config) { c.consoleOpts = opts }
}

// WithOtelHandlerOptions sets otelslog HandlerOptions.
func WithOtelHandlerOptions(opts otelslog.Option) Option {
	return func(c *config) { c.otelOpts = opts }
}

// New returns a logger that fans out to both:
//  1. a console slog handler, and
//  2. an OpenTelemetry slog handler (otelslog).
//
// name is used as the otelslog "instrumentation scope"/logger name.
func New(name string, opts ...Option) *slog.Logger {
	cfg := config{
		consoleWriter: os.Stderr,
	}
	for _, opt := range opts {
		opt(&cfg)
	}

	// Console handler (default: text handler to stdout/stderr chosen by caller via WithConsoleWriter)
	var console slog.Handler
	if cfg.consoleJSON {
		console = slog.NewJSONHandler(cfg.consoleWriter, cfg.consoleOpts)
	} else {
		console = slog.NewTextHandler(cfg.consoleWriter, cfg.consoleOpts)
	}

	// OTel handler (uses global LoggerProvider already set elsewhere)
	var otel *otelslog.Handler
	if cfg.otelOpts != nil {
		otel = otelslog.NewHandler(name, cfg.otelOpts)
	} else {
		otel = otelslog.NewHandler(name)
	}

	return slog.New(teeHandler{hs: []slog.Handler{console, otel}})
}

type teeHandler struct{ hs []slog.Handler }

func (t teeHandler) Enabled(ctx context.Context, level slog.Level) bool {
	for _, h := range t.hs {
		if h.Enabled(ctx, level) {
			return true
		}
	}
	return false
}

func (t teeHandler) Handle(ctx context.Context, r slog.Record) error {
	// Best-effort fan-out. Do not block console logging if OTLP export fails.
	for _, h := range t.hs {
		_ = h.Handle(ctx, r)
	}
	return nil
}

func (t teeHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	out := make([]slog.Handler, 0, len(t.hs))
	for _, h := range t.hs {
		out = append(out, h.WithAttrs(attrs))
	}
	return teeHandler{hs: out}
}

func (t teeHandler) WithGroup(name string) slog.Handler {
	out := make([]slog.Handler, 0, len(t.hs))
	for _, h := range t.hs {
		out = append(out, h.WithGroup(name))
	}
	return teeHandler{hs: out}
}
