package iavl

import (
	"cosmossdk.io/log/v2"
	storetypes "github.com/cosmos/cosmos-sdk/store/v2/types"

	"github.com/cosmos/cosmos-sdk/iavl/internal"
)

// LoadCommitMultiTree loads a CommitMultiStore from the given path, using the provided options.
func LoadCommitMultiTree(path string, opts Options, logger log.Logger) (storetypes.CommitMultiStore, error) {
	return internal.LoadCommitMultiTree(path, opts.toInternalOpts(), logger)
}
