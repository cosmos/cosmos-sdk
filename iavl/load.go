package iavl

import (
	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/iavl/internal"
)

// LoadCommitMultiTree loads a CommitMultiStore from the given path, using the provided options.
func LoadCommitMultiTree(path string, opts Options) (storetypes.CommitMultiStore, error) {
	return internal.LoadCommitMultiTree(path, opts.toInternalOpts())
}
