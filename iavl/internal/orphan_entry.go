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

type OrphanEntry struct {
	OrphanedVersion uint32
	NodeID          NodeID
}
