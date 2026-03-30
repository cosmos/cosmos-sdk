package internal

import (
	"fmt"
	"unsafe"
)

const SizeOrphanEntry = 12

func init() {
	// Verify the size of OrphanEntry is what we expect it to be at runtime.
	if unsafe.Sizeof(OrphanEntry{}) != SizeOrphanEntry {
		panic(fmt.Sprintf("invalid OrphanEntry size: got %d, want %d", unsafe.Sizeof(OrphanEntry{}), SizeOrphanEntry))
	}
}

// OrphanEntry records that a node was orphaned (replaced or deleted from the tree) at a specific version.
// During compaction, the OrphanRewriter reads these entries to decide which nodes can be pruned:
// if a node was orphaned before the retain version, it's no longer needed for any queryable version
// and can be deleted from the data files.
type OrphanEntry struct {
	// OrphanedVersion is the version at which this node was removed from the tree.
	OrphanedVersion uint32
	// NodeID identifies the node that was orphaned. The checkpoint in the NodeID tells us
	// which changeset's data files contain this node.
	NodeID NodeID
}
