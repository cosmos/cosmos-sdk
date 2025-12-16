package log

import (
	"context"
	"log/slog"
	"sync/atomic"
)

// filterHandler wraps a slog.Handler and applies a FilterFunc to selectively discard logs.
type filterHandler struct {
	parent slog.Handler
	filter FilterFunc
	module string // current module from attrs
}

// newFilterHandler creates a new filtering handler that wraps the parent handler.
// The filter function is called with the module and level of each log entry.
// If the filter returns true, the log entry is discarded.
func newFilterHandler(parent slog.Handler, filter FilterFunc) slog.Handler {
	if filter == nil {
		return parent
	}
	return &filterHandler{
		parent: parent,
		filter: filter,
	}
}

func (h *filterHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.parent.Enabled(ctx, level)
}

func (h *filterHandler) Handle(ctx context.Context, r slog.Record) error {
	// Get module from record attrs if not already set
	module := h.module
	if module == "" {
		r.Attrs(func(attr slog.Attr) bool {
			if attr.Key == ModuleKey {
				module = attr.Value.String()
				return false // stop iteration
			}
			return true
		})
	}

	// Apply filter
	levelStr := LevelToString(r.Level)
	if h.filter(module, levelStr) {
		return nil // filtered out
	}

	return h.parent.Handle(ctx, r)
}

func (h *filterHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	// Check if module is being set
	newModule := h.module
	for _, attr := range attrs {
		if attr.Key == ModuleKey {
			newModule = attr.Value.String()
			break
		}
	}
	return &filterHandler{
		parent: h.parent.WithAttrs(attrs),
		filter: h.filter,
		module: newModule,
	}
}

func (h *filterHandler) WithGroup(name string) slog.Handler {
	return &filterHandler{
		parent: h.parent.WithGroup(name),
		filter: h.filter,
		module: h.module,
	}
}

// verboseModeHandler wraps a slog.Handler and supports switching between regular and verbose modes.
// In verbose mode, the level is lowered and filtering is disabled.
type verboseModeHandler struct {
	parent       slog.Handler
	regularLevel slog.Level
	verboseLevel slog.Level
	filter       FilterFunc
	module       string
	verbose      *atomic.Bool // shared across all derived handlers
}

// newVerboseModeHandler creates a handler that supports verbose mode switching.
func newVerboseModeHandler(parent slog.Handler, regularLevel, verboseLevel slog.Level, filter FilterFunc) *verboseModeHandler {
	return &verboseModeHandler{
		parent:       parent,
		regularLevel: regularLevel,
		verboseLevel: verboseLevel,
		filter:       filter,
		verbose:      &atomic.Bool{},
	}
}

func (h *verboseModeHandler) SetVerboseMode(enable bool) {
	h.verbose.Store(enable)
}

func (h *verboseModeHandler) Enabled(ctx context.Context, level slog.Level) bool {
	if h.verbose.Load() {
		// In verbose mode, use verbose level
		return level >= h.verboseLevel
	}
	// In regular mode, use regular level
	return level >= h.regularLevel
}

func (h *verboseModeHandler) Handle(ctx context.Context, r slog.Record) error {
	// Check level first
	if !h.Enabled(ctx, r.Level) {
		return nil
	}

	// Apply filter only in non-verbose mode
	if !h.verbose.Load() && h.filter != nil {
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

		levelStr := LevelToString(r.Level)
		if h.filter(module, levelStr) {
			return nil // filtered out
		}
	}

	return h.parent.Handle(ctx, r)
}

func (h *verboseModeHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	newModule := h.module
	for _, attr := range attrs {
		if attr.Key == ModuleKey {
			newModule = attr.Value.String()
			break
		}
	}
	return &verboseModeHandler{
		parent:       h.parent.WithAttrs(attrs),
		regularLevel: h.regularLevel,
		verboseLevel: h.verboseLevel,
		filter:       h.filter,
		module:       newModule,
		verbose:      h.verbose, // share the same atomic bool
	}
}

func (h *verboseModeHandler) WithGroup(name string) slog.Handler {
	return &verboseModeHandler{
		parent:       h.parent.WithGroup(name),
		regularLevel: h.regularLevel,
		verboseLevel: h.verboseLevel,
		filter:       h.filter,
		module:       h.module,
		verbose:      h.verbose, // share the same atomic bool
	}
}
