package log

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"
)

// TestingT is the interface required for logging in tests.
// It is a subset of testing.T to avoid a direct dependency on the testing package.
type TestingT interface {
	Log(args ...any)
	Helper()
}

// NewTestLogger returns a logger that calls t.Log to write entries.
//
// The returned logger emits messages at any level.
// For active debugging of a test with verbose logs,
// the [NewTestLoggerInfo] and [NewTestLoggerError] functions
// only emit messages at or above the corresponding log levels.
//
// If the logs may help debug a test failure,
// you may want to use NewTestLogger(t) in your test.
// Otherwise, use NewNopLogger().
func NewTestLogger(t TestingT) Logger {
	return newTestLogger(t, slog.LevelDebug)
}

// NewTestLoggerInfo returns a test logger that filters out messages
// below info level.
//
// This is primarily helpful during active debugging of a test
// with verbose logs.
func NewTestLoggerInfo(t TestingT) Logger {
	return newTestLogger(t, slog.LevelInfo)
}

// NewTestLoggerError returns a test logger that filters out messages
// below Error level.
//
// This is primarily helpful during active debugging of a test
// with verbose logs.
func NewTestLoggerError(t TestingT) Logger {
	return newTestLogger(t, slog.LevelError)
}

func newTestLogger(t TestingT, lvl slog.Level) Logger {
	handler := &testHandler{
		t:     t,
		level: lvl,
	}
	return NewCustomLogger(slog.New(handler))
}

// testHandler is a slog.Handler that writes to testing.T.Log
type testHandler struct {
	t      TestingT
	level  slog.Level
	attrs  []slog.Attr
	groups []string
}

func (h *testHandler) Enabled(_ context.Context, level slog.Level) bool {
	return level >= h.level
}

func (h *testHandler) Handle(_ context.Context, r slog.Record) error {
	h.t.Helper()

	var buf strings.Builder

	// Format: 3:04PM INF message key=value key=value
	buf.WriteString(r.Time.Format(time.Kitchen))
	buf.WriteString(" ")

	switch {
	case r.Level < slog.LevelInfo:
		buf.WriteString("DBG")
	case r.Level < slog.LevelWarn:
		buf.WriteString("INF")
	case r.Level < slog.LevelError:
		buf.WriteString("WRN")
	default:
		buf.WriteString("ERR")
	}

	buf.WriteString(" ")
	buf.WriteString(r.Message)

	// Add pre-set attrs
	for _, attr := range h.attrs {
		buf.WriteString(" ")
		buf.WriteString(attr.Key)
		buf.WriteString("=")
		buf.WriteString(formatAttrValue(attr.Value))
	}

	// Add record attrs
	r.Attrs(func(attr slog.Attr) bool {
		buf.WriteString(" ")
		if len(h.groups) > 0 {
			buf.WriteString(strings.Join(h.groups, "."))
			buf.WriteString(".")
		}
		buf.WriteString(attr.Key)
		buf.WriteString("=")
		buf.WriteString(formatAttrValue(attr.Value))
		return true
	})

	h.t.Log(buf.String())
	return nil
}

func formatAttrValue(v slog.Value) string {
	v = v.Resolve()
	switch v.Kind() {
	case slog.KindString:
		return v.String()
	case slog.KindInt64:
		return fmt.Sprintf("%d", v.Int64())
	case slog.KindUint64:
		return fmt.Sprintf("%d", v.Uint64())
	case slog.KindFloat64:
		return fmt.Sprintf("%g", v.Float64())
	case slog.KindBool:
		return fmt.Sprintf("%t", v.Bool())
	case slog.KindDuration:
		return v.Duration().String()
	case slog.KindTime:
		return v.Time().Format(time.RFC3339)
	default:
		if val := v.Any(); val != nil {
			if s, ok := val.(fmt.Stringer); ok {
				return s.String()
			}
			return fmt.Sprintf("%v", val)
		}
		return ""
	}
}

func (h *testHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &testHandler{
		t:      h.t,
		level:  h.level,
		attrs:  append(h.attrs, attrs...),
		groups: h.groups,
	}
}

func (h *testHandler) WithGroup(name string) slog.Handler {
	return &testHandler{
		t:      h.t,
		level:  h.level,
		attrs:  h.attrs,
		groups: append(h.groups, name),
	}
}
