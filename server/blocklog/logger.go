package blocklog

import (
	"context"
	"fmt"
	"time"

	"cosmossdk.io/log/v2"
)

// Logger is a pure passthrough around another log.Logger that also records
// every call, at every level, into the Store while a block is open. It never
// filters, never returns an error, and never touches application state.
type Logger struct {
	inner  log.Logger
	store  *Store
	fields []any // accumulated With() pairs, used for attribution only
}

var (
	_ log.Logger            = (*Logger)(nil)
	_ log.VerboseModeLogger = (*Logger)(nil)
)

// Wrap returns inner wrapped for capture into store. fields are key/value
// pairs already present in inner's context (they are recorded on every entry
// but not re-applied to inner).
func Wrap(inner log.Logger, store *Store, fields ...any) log.Logger {
	return &Logger{inner: inner, store: store, fields: fields}
}

func (l *Logger) Debug(msg string, kv ...any) { l.record("debug", msg, kv); l.inner.Debug(msg, kv...) }
func (l *Logger) Info(msg string, kv ...any)  { l.record("info", msg, kv); l.inner.Info(msg, kv...) }
func (l *Logger) Warn(msg string, kv ...any)  { l.record("warn", msg, kv); l.inner.Warn(msg, kv...) }
func (l *Logger) Error(msg string, kv ...any) { l.record("error", msg, kv); l.inner.Error(msg, kv...) }

func (l *Logger) DebugContext(ctx context.Context, msg string, kv ...any) {
	l.record("debug", msg, kv)
	l.inner.DebugContext(ctx, msg, kv...)
}

func (l *Logger) InfoContext(ctx context.Context, msg string, kv ...any) {
	l.record("info", msg, kv)
	l.inner.InfoContext(ctx, msg, kv...)
}

func (l *Logger) WarnContext(ctx context.Context, msg string, kv ...any) {
	l.record("warn", msg, kv)
	l.inner.WarnContext(ctx, msg, kv...)
}

func (l *Logger) ErrorContext(ctx context.Context, msg string, kv ...any) {
	l.record("error", msg, kv)
	l.inner.ErrorContext(ctx, msg, kv...)
}

// With returns a child that keeps capturing, so module-tagged keeper loggers
// stay attributed.
func (l *Logger) With(kv ...any) log.Logger {
	fields := make([]any, 0, len(l.fields)+len(kv))
	fields = append(append(fields, l.fields...), kv...)
	return &Logger{inner: l.inner.With(kv...), store: l.store, fields: fields}
}

// Impl returns the inner logger's implementation.
func (l *Logger) Impl() any { return l.inner.Impl() }

// SetVerboseMode forwards to the inner logger when it supports verbose mode.
func (l *Logger) SetVerboseMode(v bool) {
	if vl, ok := l.inner.(log.VerboseModeLogger); ok {
		vl.SetVerboseMode(v)
	}
}

// BeginBlockLog opens capture for height. Block executors call it when they
// start executing a block.
func (l *Logger) BeginBlockLog(height int64) { l.store.Begin(height) }

// CommitBlockLog closes capture for the open height. Block executors call it
// once the block is committed.
func (l *Logger) CommitBlockLog() { l.store.Commit() }

func (l *Logger) record(level, msg string, kv []any) {
	if !l.store.open.Load() {
		return // no block open: skip the formatting work entirely
	}
	e := Entry{Time: time.Now().UTC().Format(time.RFC3339Nano), Level: level, Msg: msg}
	for _, pairs := range [][]any{l.fields, kv} {
		for i := 0; i < len(pairs); i += 2 {
			var v any
			if i+1 < len(pairs) {
				v = pairs[i+1]
			}
			k := stringify(pairs[i])
			if k == log.ModuleKey {
				e.Module = stringify(v)
				continue
			}
			if e.Fields == nil {
				e.Fields = make(map[string]string)
			}
			e.Fields[k] = stringify(v)
		}
	}
	l.store.Write(e)
}

func stringify(v any) string {
	switch x := v.(type) {
	case string:
		return x
	case error:
		return x.Error()
	case fmt.Stringer:
		return x.String()
	default:
		return fmt.Sprintf("%v", v)
	}
}
