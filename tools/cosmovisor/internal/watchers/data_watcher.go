package watchers

import (
	"context"
)

type DataWatcher[T any] struct {
	outChan chan T
	errChan chan error
}

func NewDataWatcher[T any, I any](ctx context.Context, watcher Watcher[I], unmarshal func(I) (T, error)) *DataWatcher[T] {
	outChan := make(chan T, 1)
	errChan := make(chan error, 1)
	go func() {
		defer close(outChan)
		defer close(errChan)
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
				// ignore errors because failing JSON unmarshal probably just means the file is incomplete
				if err == nil {
					outChan <- data
				}
			case err, ok := <-watcher.Errors():
				if !ok {
					return
				}
				errChan <- err

			}
		}
	}()
	return &DataWatcher[T]{
		outChan: outChan,
		errChan: errChan,
	}
}

func (d DataWatcher[T]) Updated() <-chan T {
	return d.outChan

}

func (d DataWatcher[T]) Errors() <-chan error {
	return d.errChan
}
