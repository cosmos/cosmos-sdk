package ibc

import (
	"github.com/tendermint/tendermint/lite"
	liteErr "github.com/tendermint/tendermint/lite/errors"
)

// this really shouldn't be here... removed from upstream lib
// without checking
type missingProvider struct{}

// NewMissingProvider returns a provider which does not store anything and always misses.
func NewMissingProvider() lite.Provider {
	return missingProvider{}
}

func (missingProvider) StoreCommit(lite.FullCommit) error { return nil }
func (missingProvider) GetByHeight(int64) (lite.FullCommit, error) {
	return lite.FullCommit{}, liteErr.ErrCommitNotFound()
}
func (missingProvider) GetByHash([]byte) (lite.FullCommit, error) {
	return lite.FullCommit{}, liteErr.ErrCommitNotFound()
}
func (missingProvider) LatestCommit() (lite.FullCommit, error) {
	return lite.FullCommit{}, liteErr.ErrCommitNotFound()
}
