package cosmovisor

type Watcher[T any] interface {
	Updated() <-chan T
	Errors() <-chan error
}
