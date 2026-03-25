package iavlx

import (
	"cosmossdk.io/log/v2"

	"github.com/cosmos/cosmos-sdk/iavlx/internal"
)

type CommitMultiTree = internal.CommitMultiTree
type CommitBranch = internal.CommitBranch
type CommitFinalizer = internal.CommitFinalizer

// LoadCommitMultiTree loads a CommitMultiStore from the given path, using the provided options.
// The returned store is compatible with github.com/cosmos/cosmos-sdk/store/v2/types.CommitMultiStore
// and can be used as a drop-in replacement for it.
// NOTE: with some minimal work, a compatibility wrapper for cosmossdk.io/store/types.CommitMultiStore
// could be created.
func LoadCommitMultiTree(path string, opts Options, logger log.Logger) (*internal.CommitMultiTree, error) {
	return internal.LoadCommitMultiTree(path, opts.toInternalOpts(), logger)
}
