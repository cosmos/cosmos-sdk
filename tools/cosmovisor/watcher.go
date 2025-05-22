package cosmovisor

type watcher[T any] interface {
	Updated() <-chan T
	Errors() <-chan error
}
