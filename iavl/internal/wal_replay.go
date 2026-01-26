package internal

import (
	"context"
	"fmt"
	"os"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

func ReplayWAL(ctx context.Context, root *NodePointer, walFile *os.File, rootVersion, targetVersion uint32) (*NodePointer, error) {
	_, span := tracer.Start(ctx, "ReplayWAL",
		trace.WithAttributes(
			attribute.String("walFile", walFile.Name()),
			attribute.Int64("from", int64(rootVersion)),
			attribute.Int64("to", int64(targetVersion))),
	)
	defer span.End()

	// TODO we can have mutation contexts which don't collect orphans and that don't copy transient nodes for faster replay
	stagedVersion := rootVersion + 1
	for entry, err := range ReadWAL(walFile) {
		if err != nil {
			return nil, fmt.Errorf("failed to read WAL: %w", err)
		}

		if entry.Version < uint64(stagedVersion) {
			continue
		}

		switch entry.Op {
		case WALOpCommit:
			if entry.Version != uint64(stagedVersion) {
				return nil, fmt.Errorf("WAL commit version %d does not match expected version %d", entry.Version, stagedVersion)
			}
			stagedVersion++
		case WALOpSet:
			ctx := NewMutationContext(rootVersion + 1)
			leafNode := ctx.NewLeafNode(entry.Key.SafeCopy(), entry.Value.SafeCopy())
			root, _, err = SetRecursive(root, leafNode, ctx)
			if err != nil {
				return nil, fmt.Errorf("failed to apply WAL set at version %d: %w", entry.Version, err)
			}
		case WALOpDelete:
			ctx := NewMutationContext(rootVersion + 1)
			_, root, _, err = RemoveRecursive(root, entry.Key.SafeCopy(), ctx)
			if err != nil {
				return nil, fmt.Errorf("failed to apply WAL delete at version %d: %w", entry.Version, err)
			}
		}
	}
	if stagedVersion-1 != targetVersion {
		return nil, fmt.Errorf("WAL replay ended at version %d, expected target version %d", stagedVersion-1, targetVersion)
	}
	return root, nil
}
