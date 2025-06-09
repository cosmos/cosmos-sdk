package watchers

import (
	"context"
	"time"

	"cosmossdk.io/log"
)

type Watcher[T any] interface {
	Updated() <-chan T
}

func InitWatcher[T any](ctx context.Context, logger log.Logger, pollInterval time.Duration, dirWatcher *FSNotifyWatcher, filename string, unmarshal func([]byte) (T, error)) Watcher[T] {
	if dirWatcher != nil {
		hybridWatcher := NewHybridWatcher(ctx, logger, dirWatcher, filename, pollInterval)
		return NewDataWatcher[T](ctx, logger, hybridWatcher, unmarshal)
	} else {
		pollWatcher := NewFilePollWatcher(ctx, logger, filename, pollInterval)
		return NewDataWatcher[T](ctx, logger, pollWatcher, unmarshal)
	}
}
