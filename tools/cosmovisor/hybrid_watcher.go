package cosmovisor

import (
	"context"
	"time"
)

type hybridWatcher struct {
	outChan chan []byte
	errChan chan error
}

var _ watcher[[]byte] = &hybridWatcher{}

func newHybridWatcher(ctx context.Context, dirWatcher *fsNotifyWatcher, filename string, backupPollInterval time.Duration) *hybridWatcher {
	pollWatcher := newPollWatcher(ctx, filename, backupPollInterval)
	outChan := make(chan []byte, 1)
	errChan := make(chan error, 1)

	go func() {
		defer close(outChan)
		defer close(errChan)
		for {
			select {
			case <-ctx.Done():
				return
			case update, ok := <-dirWatcher.Updated():
				if !ok {
					return
				}
				if update.Filename == filename {
					outChan <- update.Contents
				}
			case update, ok := <-pollWatcher.Updated():
				if !ok {
					return
				}
				outChan <- update
			case err, ok := <-dirWatcher.Errors():
				if !ok {
					return
				}
				errChan <- err
			case err, ok := <-pollWatcher.Errors():
				if !ok {
					return
				}
				errChan <- err
			}
		}
	}()

	return &hybridWatcher{
		outChan: outChan,
		errChan: errChan,
	}
}

func (h hybridWatcher) Updated() <-chan []byte {
	return h.outChan
}

func (h hybridWatcher) Errors() <-chan error {
	return h.errChan
}
