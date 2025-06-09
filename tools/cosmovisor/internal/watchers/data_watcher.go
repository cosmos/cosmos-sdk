package watchers

import (
	"context"

	"cosmossdk.io/log"
)

type DataWatcher[T any] struct {
	outChan chan T
}

func NewDataWatcher[T any, I any](ctx context.Context, logger log.Logger, watcher Watcher[I], unmarshal func(I) (T, error)) *DataWatcher[T] {
	outChan := make(chan T, 1)
	go func() {
		defer close(outChan)
		for {
			select {
			case <-ctx.Done():
				return
			case contents, ok := <-watcher.Updated():
				if !ok {
					return
				}
				var data T
				data, err := unmarshal(contents)
				if err == nil {
					outChan <- data
				} else {
					logger.Warn("failed to unmarshal data", "error", err)
				}
			}
		}
	}()
	return &DataWatcher[T]{
		outChan: outChan,
	}
}

func (d DataWatcher[T]) Updated() <-chan T {
	return d.outChan

}
