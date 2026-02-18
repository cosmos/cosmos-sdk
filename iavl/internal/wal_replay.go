package internal

import (
	"context"
	"fmt"
	"os"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

func ReplayWALForStartup(ctx context.Context, root *NodePointer, walFile *os.File, rootVersion, expectedVersion uint32) (*NodePointer, uint32, error) {
	_, span := tracer.Start(ctx, "ReplayWALForStartup",
		trace.WithAttributes(
			attribute.String("walFile", walFile.Name()),
			attribute.Int64("from", int64(rootVersion)),
			attribute.Int64("to", int64(expectedVersion))),
	)
	defer span.End()

	var rollbackOffset int
	for entry, err := range ReadWAL(walFile) {
		if err != nil {
			return nil, 0, fmt.Errorf("failed to read WAL: %w", err)
		}

		if entry.Version <= uint64(rootVersion) {
			continue
		}
		if expectedVersion != 0 {
			if entry.Version == uint64(expectedVersion)+1 {
				// we will need to rollback these entries but this isn't an error quite yet
				if rollbackOffset == 0 {
					rollbackOffset = entry.Offset
				}
				continue
			}
			if entry.Version > uint64(expectedVersion)+1 {
				// this means we've gone more than 1 version beyond the expected version
				// this is an unrecoverable error (some unexpected data corruption)
				return nil, 0, fmt.Errorf("WAL commit version %d is more than 1 version beyond expected version %d, WAL is corrupted", entry.Version, expectedVersion)
			}
		}

		root, err = applyWalEntry(entry, root, rootVersion)
		if err != nil {
			return nil, 0, err
		}

		if entry.Op == WALOpCommit {
			rootVersion++
		}
	}

	if rollbackOffset > 0 {
		logger.WarnContext(ctx, "WAL contains entries beyond expected version, rolling back to expected version", "walFile", walFile.Name(), "expectedVersion", expectedVersion, "rollbackOffset", rollbackOffset)
		// must rollback if we saw extra entries past the expected version
		err := RollbackFileToOffset(walFile, int64(rollbackOffset))
		if err != nil {
			return nil, 0, fmt.Errorf("failed to rollback WAL file to offset %d: %w", rollbackOffset, err)
		}
	}

	// finished replaying WAL
	return root, rootVersion, nil
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
