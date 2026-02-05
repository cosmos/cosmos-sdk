package watchers

import (
	"context"
	"time"
)

type HybridWatcher struct {
	outChan chan []byte
}

var _ Watcher[[]byte] = &HybridWatcher{}

func NewHybridWatcher(ctx context.Context, errorHandler ErrorHandler, dirWatcher *FSNotifyWatcher, filename string, backupPollInterval time.Duration) *HybridWatcher {
	pollWatcher := NewFilePollWatcher(ctx, errorHandler, filename, backupPollInterval)
	outChan := make(chan []byte, 1)

	go func() {
		defer close(outChan)
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
	}
}

func (h HybridWatcher) Updated() <-chan []byte {
	return h.outChan
}
