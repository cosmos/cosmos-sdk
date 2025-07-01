package watchers

import (
	"context"
	"fmt"
	"os"

	"github.com/fsnotify/fsnotify"

	"cosmossdk.io/log"
)

type FSNotifyWatcher struct {
	watcher *fsnotify.Watcher
	outChan chan FileUpdate
}

var _ Watcher[FileUpdate] = (*FSNotifyWatcher)(nil)

func NewFSNotifyWatcher(ctx context.Context, logger log.Logger, dir string, filenames []string) (*FSNotifyWatcher, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	err = watcher.Add(dir)
	if err != nil {
		return nil, fmt.Errorf("failed to watch directory %s: %w", dir, err)
	}

    // TODO check that filenames are in dir & fully qualified
    // Validate filenames are absolute paths within the watched directory
    filenameSet := make(map[string]struct{})
    for _, filename := range filenames {
        if !filepath.IsAbs(filename) {
            return nil, fmt.Errorf("filename must be absolute path: %s", filename)
        }
        if !strings.HasPrefix(filename, dir) {
            return nil, fmt.Errorf("filename must be within watched directory: %s", filename)
        }
        filenameSet[filename] = struct{}{}
    }
	filenameSet := make(map[string]struct{})
	for _, filename := range filenames {
		filenameSet[filename] = struct{}{}
	}

	outChan := make(chan FileUpdate, 1)
	errChan := make(chan error, 1)
	go func() {
		// close the watcher and channels
		// when the goroutines exits via return's
		defer func(watcher *fsnotify.Watcher) {
			_ = watcher.Close()
		}(watcher)
		defer close(outChan)
		defer close(errChan)

		for {
			select {
			case <-ctx.Done():
				return
			case event, ok := <-watcher.Events:
				if !ok { // channel closed
					return
				}
				if event.Has(fsnotify.Write) {
					if _, ok := filenameSet[event.Name]; !ok {
						continue
					}
					filename := event.Name
					bz, err := os.ReadFile(filename)
					if err != nil {
						errChan <- fmt.Errorf("failed to read file %s: %w", filename, err)
					} else {
						outChan <- FileUpdate{Filename: filename, Contents: bz}
					}
				}
			case err, ok := <-watcher.Errors:
				if !ok { // channel closed
					return
				}
				logger.Error("fsnotify error", "error", err)
			}
		}
	}()

	return &FSNotifyWatcher{
		watcher: watcher,
		outChan: outChan,
	}, nil
}

type FileUpdate struct {
	Filename string
	Contents []byte
}

func (w *FSNotifyWatcher) Updated() <-chan FileUpdate {
	return w.outChan
}
