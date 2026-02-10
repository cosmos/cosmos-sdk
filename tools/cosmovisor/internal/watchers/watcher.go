package watchers

import (
	"context"
	"time"

	"cosmossdk.io/log"
)

// Watcher is an interface that defines a generic watcher that emits updates of type T.
type Watcher[T any] interface {
	// Updated returns a channel that emits updates of type T.
	Updated() <-chan T
}

// ErrorHandler is an interface for handling errors and warnings in watchers.
// This abstraction decouples watchers from logging policy decisions. Watchers report
// issues without deciding severity, while the handler determines the appropriate action
// (e.g., the default implementation downgrades errors to warnings since file watching
// errors are typically non-fatal and shouldn't alarm operators).
// But rather than hardcoding this behavior at the call site, this abstraction allows for
// an easy way to swap this out in the future.
type ErrorHandler interface {
	// Error handles an error as an error.
	Error(msg string, err error)
	// Warn handles an error as a warning.
	Warn(msg string, err error)
}

type debugLoggerErrorHandler struct {
	logger log.Logger
}

func (h *debugLoggerErrorHandler) Error(msg string, err error) {
	h.logger.Warn(msg, "error", err)
}

func (h *debugLoggerErrorHandler) Warn(msg string, err error) {
	h.logger.Debug(msg, "error", err)
}

// DebugLoggerErrorHandler returns an ErrorHandler that logs errors and warnings using the provided logger,
// but downgrades errors to warnings and warnings to debug logs.
func DebugLoggerErrorHandler(logger log.Logger) ErrorHandler {
	return &debugLoggerErrorHandler{logger: logger}
}

// InitFileWatcher initializes a file watcher which uses either both fsnotify and polling (hybrid watcher) or just polling,
// depending on whether a fsnotify directory watcher is provided.
func InitFileWatcher[T any](ctx context.Context, errorHandler ErrorHandler, pollInterval time.Duration, dirWatcher *FSNotifyWatcher, filename string, unmarshal func([]byte) (T, error)) Watcher[T] {
	if dirWatcher != nil {
		hybridWatcher := NewHybridWatcher(ctx, errorHandler, dirWatcher, filename, pollInterval)
		return NewDataWatcher[T](ctx, errorHandler, hybridWatcher, unmarshal)
	} else {
		pollWatcher := NewFilePollWatcher(ctx, errorHandler, filename, pollInterval)
		return NewDataWatcher[T](ctx, errorHandler, pollWatcher, unmarshal)
	}
}
