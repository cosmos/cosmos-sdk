package cosmovisor

import (
	"context"
	"fmt"
	"os"
	"time"
)

type pollWatcher struct {
	outChan chan []byte
	errChan chan error
}

var _ watcher[[]byte] = (*pollWatcher)(nil)

func newPollWatcher(ctx context.Context, filename string, pollInterval time.Duration) *pollWatcher {
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
	return &pollWatcher{
		outChan: outChan,
		errChan: errChan,
	}
}

func (w *pollWatcher) Updated() <-chan []byte {
	return w.outChan
}

func (w *pollWatcher) Errors() <-chan error {
	return w.errChan
}
