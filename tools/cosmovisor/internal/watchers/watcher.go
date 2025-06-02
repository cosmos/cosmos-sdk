package watchers

import (
	"context"
	"time"
)

type Watcher[T any] interface {
	Updated() <-chan T
	Errors() <-chan error
}

func InitWatcher[T any](ctx context.Context, pollInterval time.Duration, dirWatcher *FSNotifyWatcher, filename string, unmarshal func([]byte) (T, error)) Watcher[T] {
	if dirWatcher != nil {
		hybridWatcher := NewHybridWatcher(ctx, dirWatcher, filename, pollInterval)
		return NewDataWatcher[T](ctx, hybridWatcher, unmarshal)
	} else {
		pollWatcher := NewFilePollWatcher(ctx, filename, pollInterval)
		return NewDataWatcher[T](ctx, pollWatcher, unmarshal)
	}
}
