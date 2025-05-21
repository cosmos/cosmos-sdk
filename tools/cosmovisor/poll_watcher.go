package cosmovisor

import (
	"context"
	"fmt"
	"os"
	"time"
)

type PollWatcher struct {
	outChan chan []byte
	errChan chan error
}

var _ Watcher[[]byte] = (*PollWatcher)(nil)

func NewPollWatcher(ctx context.Context, filename string, pollInterval time.Duration) *PollWatcher {
	outChan := make(chan []byte, 1)
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
				stat, err := os.Stat(filename)
				if err != nil {
					if !os.IsNotExist(err) {
						errChan <- fmt.Errorf("failed to stat upgrade info file: %w", err)
					}
				} else if stat.Size() > 0 {
					bz, err := os.ReadFile(filename)
					if err != nil {
						errChan <- fmt.Errorf("failed to read file %s: %w", filename, err)
					} else {
						outChan <- bz
					}
				}
			}
		}
	}()
	return &PollWatcher{
		outChan: outChan,
		errChan: errChan,
	}
}

func (w *PollWatcher) Updated() <-chan []byte {
	return w.outChan
}

func (w *PollWatcher) Errors() <-chan error {
	return w.errChan
}
