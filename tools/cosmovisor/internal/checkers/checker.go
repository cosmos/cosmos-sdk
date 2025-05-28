package checkers

type HeightChecker interface {
	GetLatestBlockHeight() (uint64, error)
}
