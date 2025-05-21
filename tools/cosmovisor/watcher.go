package cosmovisor

import (
	"encoding/json"
)

type DataWatcher[T any] struct {
	watcher        Watcher[[]byte]
	ReceivedUpdate chan T
}

func NewDataWatcher[T any](watcher Watcher[[]byte]) *DataWatcher[T] {
	ch := make(chan T)
	go func() {
		for {
			select {
			case contents := <-watcher.Updated():
				var data T
				err := json.Unmarshal(contents, &data)
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

type Watcher[T any] interface {
	Updated() <-chan T
	Errors() <-chan error
}
