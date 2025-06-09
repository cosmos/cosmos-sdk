package watchers

import (
	"context"
	"time"

	"cosmossdk.io/log"
)

type HybridWatcher struct {
	outChan chan []byte
	errChan chan error
}

var _ Watcher[[]byte] = &HybridWatcher{}

func NewHybridWatcher(ctx context.Context, logger log.Logger, dirWatcher *FSNotifyWatcher, filename string, backupPollInterval time.Duration) *HybridWatcher {
	pollWatcher := NewFilePollWatcher(ctx, logger, filename, backupPollInterval)
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
			}
		}
	}()

	return &HybridWatcher{
		outChan: outChan,
		errChan: errChan,
	}
}

func (h HybridWatcher) Updated() <-chan []byte {
	return h.outChan
}

func (h HybridWatcher) Errors() <-chan error {
	return h.errChan
}
