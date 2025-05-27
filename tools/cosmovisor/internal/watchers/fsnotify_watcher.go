package watchers

import (
	"context"
	"fmt"
	"os"

	"github.com/fsnotify/fsnotify"
)

type FSNotifyWatcher struct {
	watcher *fsnotify.Watcher
	outChan chan FileUpdate
	errChan chan error
}

var _ Watcher[FileUpdate] = (*FSNotifyWatcher)(nil)

func NewFSNotifyWatcher(ctx context.Context, dir string, filenames []string) (*FSNotifyWatcher, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	err = watcher.Add(dir)
	if err != nil {
		return nil, fmt.Errorf("failed to watch directory %s: %w", dir, err)
	}

	// TODO check that filenames are in dir & fully qualified
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
				errChan <- fmt.Errorf("fsnotify error: %w", err)
			}
		}
	}()

	return &FSNotifyWatcher{
		watcher: watcher,
		outChan: outChan,
		errChan: errChan,
	}, nil
}

type FileUpdate struct {
	Filename string
	Contents []byte
}

func (w *FSNotifyWatcher) Updated() <-chan FileUpdate {
	return w.outChan
}

func (w *FSNotifyWatcher) Errors() <-chan error {
	return w.errChan
}
