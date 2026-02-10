package watchers

import (
	"context"
	"time"
)

// HybridWatcher combines fsnotify-based file watching with periodic polling.
// This dual approach provides both fast detection (via fsnotify) and reliability (via polling).
// fsnotify can be unreliable in certain environments (NFS, Docker volumes, some kernels),
// so polling serves as a fallback. Whichever mechanism detects a change first will
// trigger the notification. It is noted that this hybrid approach is overly and maybe unnecessarily
// defensive, but we hedge on the side of being overly cautious for critical file watching.
type HybridWatcher struct {
	outChan chan []byte
}

var _ Watcher[[]byte] = &HybridWatcher{}

// NewHybridWatcher creates a watcher that monitors a file using both fsnotify and polling.
// Updates from either source are forwarded to the output channel.
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
