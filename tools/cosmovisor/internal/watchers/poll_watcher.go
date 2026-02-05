package watchers

import (
	"context"
	"os"
	"reflect"
	"time"
)

type PollWatcher[T any] struct {
	outChan      chan T
	errorHandler ErrorHandler
	checker      func() (T, error)
	pollInterval time.Duration
}

func NewPollWatcher[T any](errorHandler ErrorHandler, checker func() (T, error), pollInterval time.Duration) *PollWatcher[T] {
	outChan := make(chan T, 1)
	return &PollWatcher[T]{
		errorHandler: errorHandler,
		checker:      checker,
		pollInterval: pollInterval,
		outChan:      outChan,
	}
}

func (w *PollWatcher[T]) Start(ctx context.Context) {
	ticker := time.NewTicker(w.pollInterval)
	go func() {
		defer ticker.Stop()
		defer close(w.outChan)

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				x, err := w.checker()
				if err != nil && !os.IsNotExist(err) {
					w.errorHandler.Error("failed to check for updates", err)
				} else if err == nil {
					// to make PollWatcher generic on any type T (including []byte), we use reflect.DeepEqual and the default zero value of T
					var zero T
					if !reflect.DeepEqual(x, zero) {
						w.outChan <- x
					}
				}
			}
		}
	}()
}

func (w *PollWatcher[T]) Updated() <-chan T {
	return w.outChan
}
