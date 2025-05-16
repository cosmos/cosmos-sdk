package cosmovisor

import (
	"encoding/json"
	"time"
)

type DataWatcher[T any] struct {
	watcher        FileWatcher
	ReceivedUpdate chan T
}

func NewDataWatcher[T any](watcher FileWatcher) *DataWatcher[T] {
	ch := make(chan T)
	go func() {
		for {
			select {
			case contents := <-watcher.FileContentsUpdated():
				var data T
				err := json.Unmarshal([]byte(contents), &data)
				if err == nil {
					ch <- data
				}
			}
		}
	}()
	return &DataWatcher[T]{
		watcher: watcher,
	}
}

type FileWatcher interface {
	FileContentsUpdated() <-chan string
	Stop()
}

type PollWatcher struct {
}

var _ FileWatcher = (*PollWatcher)(nil)

func NewPollWatcher(filename string, pollInterval time.Duration) *PollWatcher {
	return &PollWatcher{}
}

func (w *PollWatcher) FileContentsUpdated() <-chan string {
	panic("not implemented")
}

func (w *PollWatcher) Stop() {
	//TODO implement me
	panic("implement me")
}

type FSNotifyWatcher struct {
}

var _ FileWatcher = (*FSNotifyWatcher)(nil)

func NewFSNotifyWatcher(filename string) *FSNotifyWatcher {
	return &FSNotifyWatcher{}
}

func (w *FSNotifyWatcher) FileContentsUpdated() <-chan string {
	panic("not implemented")
}

func (w *FSNotifyWatcher) Stop() {
	//TODO implement me
	panic("implement me")
}
