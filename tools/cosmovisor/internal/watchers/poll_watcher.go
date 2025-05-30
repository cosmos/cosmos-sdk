package watchers

import (
	"context"
	"fmt"
	"os"
	"time"
)

type PollWatcher[T any] struct {
	outChan chan T
	errChan chan error
}

func NewPollWatcher[T any](ctx context.Context, checker func() (T, error), pollInterval time.Duration) *PollWatcher[T] {
	outChan := make(chan T, 1)
	errChan := make(chan error, 1)
	ticker := time.NewTicker(pollInterval)
	go func() {
		defer ticker.Stop()
		defer close(outChan)
		defer close(errChan)

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				_, err := checker()
				if err != nil {
					if !os.IsNotExist(err) {
						errChan <- fmt.Errorf("failed to check for updates: %w", err)
					}
				} else {
					panic("TODO: generically check for empty value")
					//if x != nil {
					//	outChan <- x
					//}
				}
			}
		}
	}()
	return &PollWatcher[T]{
		outChan: outChan,
		errChan: errChan,
	}
}

func (w *PollWatcher[T]) Updated() <-chan T {
	return w.outChan
}

func (w *PollWatcher[T]) Errors() <-chan error {
	return w.errChan
}
