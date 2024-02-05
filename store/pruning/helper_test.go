package pruning

type busyPruningStore struct {
	ch chan struct{}

	count int // count of Prune calls
}

func newBusyPruningStore() *busyPruningStore {
	return &busyPruningStore{
		ch: make(chan struct{}),
	}
}

func (s *busyPruningStore) Prune(version uint64) error {
	<-s.ch
	s.count++
	return nil
}

func (s *busyPruningStore) finishPruning() {
	s.ch <- struct{}{}
}

func (s *busyPruningStore) Close() {
	close(s.ch)
}
