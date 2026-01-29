package internal

import (
	"context"
	"fmt"
	"os"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// ReplayWAL replays WAL entries from walFile starting from root at rootVersion up to targetVersion.
// If targetVersion is 0, it replays to the end of the WAL.
// It returns the new root node pointer at the target version, the actual version reached, and any error encountered.
func ReplayWAL(ctx context.Context, root *NodePointer, walFile *os.File, rootVersion, targetVersion uint32) (*NodePointer, uint32, error) {
	_, span := tracer.Start(ctx, "ReplayWAL",
		trace.WithAttributes(
			attribute.String("walFile", walFile.Name()),
			attribute.Int64("from", int64(rootVersion)),
			attribute.Int64("to", int64(targetVersion))),
	)
	defer span.End()

	if targetVersion != 0 && rootVersion == targetVersion {
		// early exit if no replay is needed
		return root, 0, nil
	}

	// TODO we can have mutation contexts which don't collect orphans and that don't copy transient nodes for faster replay
	stagedVersion := rootVersion + 1
	for entry, err := range ReadWAL(walFile) {
		if err != nil {
			return nil, 0, fmt.Errorf("failed to read WAL: %w", err)
		}

		if entry.Version < uint64(stagedVersion) {
			continue
		}

		switch entry.Op {
		case WALOpCommit:
			if entry.Version != uint64(stagedVersion) {
				return nil, 0, fmt.Errorf("WAL commit version %d does not match expected version %d", entry.Version, stagedVersion)
			}
			if entry.Version == uint64(targetVersion) {
				// reached target version
				return root, targetVersion, nil
			}
			stagedVersion++
		case WALOpSet:
			ctx := NewMutationContext(stagedVersion, 0)
			leafNode := ctx.NewLeafNode(entry.Key.SafeCopy(), entry.Value.SafeCopy())
			root, _, err = SetRecursive(root, leafNode, ctx)
			if err != nil {
				return nil, 0, fmt.Errorf("failed to apply WAL set at version %d: %w", entry.Version, err)
			}
		case WALOpDelete:
			ctx := NewMutationContext(stagedVersion, 0)
			_, root, _, err = RemoveRecursive(root, entry.Key.SafeCopy(), ctx)
			if err != nil {
				return nil, 0, fmt.Errorf("failed to apply WAL delete at version %d: %w", entry.Version, err)
			}
		}
	}
	if targetVersion == 0 {
		// special case: replay to the end of the WAL
		return root, stagedVersion - 1, nil
	}
	return nil, 0, fmt.Errorf("WAL replay reached end of file before target version %d", targetVersion)
}
