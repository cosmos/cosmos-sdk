package checkers

type HeightChecker interface {
	GetLatestBlockHeight() (uint64, error)
}

type HeightWatcher struct{}

func (h HeightWatcher) Updated() <-chan uint64 {
	// TODO persist to file
	panic("not implemented")
}

func (h HeightWatcher) ReadNow() (uint64, error) {
	// TODO attempt to read actual height, read from file as a fallback
	panic("not implemented")
}
