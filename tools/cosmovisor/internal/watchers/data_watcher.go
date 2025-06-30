package watchers

import (
	"context"
)

type DataWatcher[T any] struct {
	outChan chan T
}

func NewDataWatcher[T, I any](ctx context.Context, errorHandler ErrorHandler, watcher Watcher[I], unmarshal func(I) (T, error)) *DataWatcher[T] {
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
					errorHandler.Warn("failed to unmarshal data", err)
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
