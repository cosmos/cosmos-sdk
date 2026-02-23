package internal

import (
	"context"
	"fmt"
	"os"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"

	"cosmossdk.io/log/v2"
)

// ReplayWALForStartup replays WAL entries from walFile starting from root at rootVersion up to expectedVersion.
// It returns the new root node pointer at the expected version if possible, the actual version reached, a bool indicating whether a rollback was needed, and any error encountered.
// If autoRepair is true, it will attempt to roll back the WAL file to the last good offset if it encounters entries beyond the expected version,
// which can happen if there was a crash during a commit that caused partial writes to the WAL.
func ReplayWALForStartup(ctx context.Context, root *NodePointer, walFile *os.File, rootVersion, expectedVersion uint32, logger log.Logger, autoRepair bool) (*NodePointer, uint32, bool, error) {
	_, span := tracer.Start(ctx, "ReplayWALForStartup",
		trace.WithAttributes(
			attribute.String("walFile", walFile.Name()),
			attribute.Int64("from", int64(rootVersion)),
			attribute.Int64("to", int64(expectedVersion))),
	)
	defer span.End()

	var lastGoodOffset int // defaults to zero or start of the file
	var needRollback bool
	for entry, err := range ReadWAL(walFile) {
		if err != nil {
			if expectedVersion != 0 && rootVersion == expectedVersion {
				// if we are at the expected version already and we have some corrupted data from a partial write
				// we stop reading further and just roll back to the last good offset
				// which should be the end of the last commit entry, since we got to the right commit
				needRollback = true
				break
			}
			return nil, 0, false, fmt.Errorf("failed to read WAL: %w", err)
		}

		if entry.Version <= uint64(rootVersion) {
			if entry.Op == WALOpCommit {
				// this is also a good rollback offset, since it's the end of a commit, so capture it
				lastGoodOffset = entry.EndOffset
			}
			continue
		}
		if expectedVersion != 0 {
			if entry.Version == uint64(expectedVersion)+1 {
				// we will need to rollback these entries but this isn't an error quite yet
				needRollback = true
				continue
			}
			if entry.Version > uint64(expectedVersion)+1 {
				// this means we've gone more than 1 version beyond the expected version
				// this is an unrecoverable error (some unexpected data corruption)
				return nil, 0, false, fmt.Errorf("WAL commit version %d is more than 1 version beyond expected version %d, WAL is corrupted", entry.Version, expectedVersion)
			}
		}

		root, err = applyWalEntry(entry, root, rootVersion)
		if err != nil {
			return nil, 0, false, err
		}

		if entry.Op == WALOpCommit {
			rootVersion++
			// then end of a commit is a good rollback offset, so capture it
			lastGoodOffset = entry.EndOffset
		}
	}

	if needRollback {
		if !autoRepair {
			return nil, 0, false, fmt.Errorf("WAL contains entries beyond expected version %d, auto repair disabled", expectedVersion)
		}

		logger.WarnContext(ctx, "WAL contains entries beyond expected version, rolling back to expected version", "walFile", walFile.Name(), "expectedVersion", expectedVersion, "rollbackOffset", lastGoodOffset)
		// must rollback if we saw extra entries past the expected version
		err := RollbackFileToOffset(walFile, int64(lastGoodOffset))
		if err != nil {
			return nil, 0, false, fmt.Errorf("failed to rollback WAL file to offset %d: %w", lastGoodOffset, err)
		}
	}

	// finished replaying WAL
	return root, rootVersion, needRollback, nil
}

// ReplayWALForQuery replays WAL entries from walFile starting from root at rootVersion up to targetVersion.
// It returns the new root node pointer at the target version if possible, the actual version reached, and any error encountered.
// WAL replay will still succeed even if we can't reach the target version, returning the highest version reached,
// so that replay can continue with the next WAL segment.
func ReplayWALForQuery(ctx context.Context, root *NodePointer, walFile *os.File, rootVersion, targetVersion uint32) (*NodePointer, uint32, error) {
	_, span := tracer.Start(ctx, "ReplayWALForQuery",
		trace.WithAttributes(
			attribute.String("walFile", walFile.Name()),
			attribute.Int64("from", int64(rootVersion)),
			attribute.Int64("to", int64(targetVersion))),
	)
	defer span.End()

	for entry, err := range ReadWAL(walFile) {
		if err != nil {
			return nil, 0, fmt.Errorf("failed to read WAL: %w", err)
		}

		if entry.Version <= uint64(rootVersion) {
			continue
		}

		root, err = applyWalEntry(entry, root, rootVersion)
		if err != nil {
			return nil, 0, err
		}

		if entry.Op == WALOpCommit {
			if entry.Version == uint64(targetVersion) {
				// reached target version
				return root, targetVersion, nil
			}
			rootVersion++
		}
	}

	// finished replaying WAL
	return root, rootVersion, nil
}

func applyWalEntry(entry WALEntry, root *NodePointer, version uint32) (newRoot *NodePointer, err error) {
	stagedVersion := version + 1
	switch entry.Op {
	case WALOpCommit:
		if entry.Version != uint64(stagedVersion) {
			if entry.Version < uint64(stagedVersion) {
				return nil, fmt.Errorf("version %d is no longer available (WAL starts at version %d); it may have been pruned", stagedVersion, entry.Version)
			}
			return nil, fmt.Errorf("WAL commit version %d does not match expected staged version %d, WAL is corrupted", entry.Version, stagedVersion)
		}
		return root, nil
	case WALOpSet:
		ctx := NewMutationContext(stagedVersion, 0) // all nodes can be mutated
		leafNode := ctx.NewLeafNode(entry.Key.SafeCopy(), entry.Value.SafeCopy())
		newRoot, _, err = SetRecursive(root, leafNode, ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to apply WAL set at version %d: %w", entry.Version, err)
		}
		return newRoot, nil
	case WALOpDelete:
		ctx := NewMutationContext(stagedVersion, 0) // all nodes can be mutated
		_, newRoot, _, err = RemoveRecursive(root, entry.Key.SafeCopy(), ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to apply WAL delete at version %d: %w", entry.Version, err)
		}
		return newRoot, nil
	default:
		return nil, fmt.Errorf("invalid WAL entry operation %d at version %d", entry.Op, entry.Version)
	}
}
