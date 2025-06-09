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
type ErrorHandler interface {
	// Error handles an error as an error.
	Error(msg string, err error)
	// Warn handles an error as a warning.
	Warn(msg string, err error)
}

type loggerErrorHandler struct {
	logger log.Logger
}

func (h *loggerErrorHandler) Error(msg string, err error) {
	h.logger.Error(msg, "error", err)
}

func (h *loggerErrorHandler) Warn(msg string, err error) {
	h.logger.Warn(msg, "error", err)
}

// LoggerErrorHandler returns an ErrorHandler that logs errors and warnings using the provided logger.
func LoggerErrorHandler(logger log.Logger) ErrorHandler {
	return &loggerErrorHandler{logger: logger}
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
