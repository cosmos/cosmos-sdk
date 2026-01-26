package internal

import (
	"fmt"
	"os"
)

func ReplayWAL(root *NodePointer, walFile *os.File, rootVersion, targetVersion uint32) (*NodePointer, error) {
	walReader, err := NewWALReader(walFile)
	if err != nil {
		return nil, err
	}

	// TODO we can have mutation contexts which don't collect orphans and that don't copy transient nodes for faster replay
	for {
		entryType, ok, err := walReader.Next()
		if err != nil {
			return nil, err
		}
		if !ok {
			if rootVersion != targetVersion {
				return nil, fmt.Errorf("WAL ended before reaching target version %d (current %d)", targetVersion, rootVersion)
			}
			return root, nil
		}

		if walReader.Version < uint64(rootVersion+1) {
			continue
		}

		switch entryType {
		case WALEntryStart:
			// skip start entry
		case WALEntryCommit:
			if walReader.Version != uint64(rootVersion+1) {
				return nil, fmt.Errorf("WAL commit version %d does not match expected version %d", walReader.Version, rootVersion+1)
			}
			rootVersion = uint32(walReader.Version)
		case WALEntrySet:
			ctx := NewMutationContext(rootVersion + 1)
			leafNode := ctx.NewLeafNode(walReader.Key.SafeCopy(), walReader.Value.SafeCopy())
			root, _, err = SetRecursive(root, leafNode, ctx)
			if err != nil {
				return nil, fmt.Errorf("failed to apply WAL set at version %d: %w", walReader.Version, err)
			}
		case WALEntryDelete:
			ctx := NewMutationContext(rootVersion + 1)
			_, root, _, err = RemoveRecursive(root, walReader.Key.SafeCopy(), ctx)
			if err != nil {
				return nil, fmt.Errorf("failed to apply WAL delete at version %d: %w", walReader.Version, err)
			}
		}
	}
}
