package watchers

import (
	"context"
	"os"
	"reflect"
	"time"

	"cosmossdk.io/log"
)

type PollWatcher[T any] struct {
	outChan chan T
}

func NewPollWatcher[T any](ctx context.Context, logger log.Logger, checker func() (T, error), pollInterval time.Duration) *PollWatcher[T] {
	outChan := make(chan T, 1)
	ticker := time.NewTicker(pollInterval)
	go func() {
		defer ticker.Stop()
		defer close(outChan)

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				x, err := checker()
				if err != nil {
					if !os.IsNotExist(err) {
						logger.Error("failed to check for updates: %w", "error", err)
					}
				} else {
					// to make PollWatcher generic on any type T (including []byte), we use reflect.DeepEqual and the default zero value of T
					var zero T
					if !reflect.DeepEqual(x, zero) {
						outChan <- x
					}
				}
			}
		}
	}()
	return &PollWatcher[T]{
		outChan: outChan,
	}
}

func (w *PollWatcher[T]) Updated() <-chan T {
	return w.outChan
}
