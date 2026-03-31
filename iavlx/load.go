package iavlx

import (
	"cosmossdk.io/log/v2"

	"github.com/cosmos/cosmos-sdk/iavlx/internal"
)

// CommitMultiTree manages multiple named IAVL trees with atomic cross-tree commits.
// It implements storetypes.CommitMultiStore. During block execution, callers interact
// with it through a CommitBranch (obtained via CommitBranch()) which provides cached
// writes, and then commit atomically via a CommitFinalizer.
type CommitMultiTree = internal.CommitMultiTree

// CommitBranch is a writeable cached MultiTree that bridges pending in-memory writes
// from block execution and the underlying CommitMultiTree. When block execution is
// complete, call StartCommit to kick off parallel per-tree commits, which returns a
// CommitFinalizer for the finalize-or-rollback decision.
type CommitBranch = internal.CommitBranch

// CommitFinalizer coordinates the multi-tree commit after StartCommit has been called.
// It waits for per-tree hashes, writes commit info to disk, and then either finalizes
// (makes the commit permanent) or rolls back (truncates per-tree WALs) based on the
// caller's signal via StartFinalize/Finalize or Rollback.
type CommitFinalizer = internal.CommitFinalizer

// LoadCommitMultiTree loads a CommitMultiStore from the given path, using the provided options.
// path should be a directory that either already contains iavlx data or will be created on first use.
// The returned store is compatible with github.com/cosmos/cosmos-sdk/store/v2/types.CommitMultiStore
// and can be used as a drop-in replacement for it.
// The caller must call Close on the returned CommitMultiTree when done to release file handles and
// background goroutines.
// NOTE: with some minimal work, a compatibility wrapper for cosmossdk.io/store/types.CommitMultiStore
// could be created.
func LoadCommitMultiTree(path string, opts Options, logger log.Logger) (*internal.CommitMultiTree, error) {
	return internal.LoadCommitMultiTree(path, opts.toInternalOpts(), logger)
}
