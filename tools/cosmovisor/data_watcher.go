package cosmovisor

import (
	"context"
	"encoding/json"
)

type dataWatcher[T any] struct {
	outChan chan T
	errChan chan error
}

func newDataWatcher[T any](ctx context.Context, watcher watcher[[]byte]) *dataWatcher[T] {
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
				err := json.Unmarshal(contents, &data)
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
	return &dataWatcher[T]{
		outChan: outChan,
		errChan: errChan,
	}
}

func (d dataWatcher[T]) Updated() <-chan T {
	return d.outChan

}

func (d dataWatcher[T]) Errors() <-chan error {
	return d.errChan
}
