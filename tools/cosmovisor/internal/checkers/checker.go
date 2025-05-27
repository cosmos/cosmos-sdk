package checkers

type LatestBlockChecker interface {
	GetLatestBlockHeight() (uint64, error)
}
