package internal

import (
	"fmt"
	"unsafe"
)

const versionInfoSize = 40

func init() {
	if unsafe.Sizeof(VersionInfo{}) != versionInfoSize {
		panic(fmt.Sprintf("invalid VersionInfo size: got %d, want %d", unsafe.Sizeof(VersionInfo{}), versionInfoSize))
	}
}

// VersionInfo holds metadata about a specific version of changes in the IAVL tree.
type VersionInfo struct {
	// Leaves holds metadata about the leaf nodes in this version.
	Leaves NodeSetInfo
	// Branches holds metadata about the branch nodes in this version.
	Branches NodeSetInfo
	// RootID is the NodeID of the root node for this version.
	// The NodeID redundantly includes the version number as well as a sort of sanity-check.
	// If the tree is empty at this version, RootID will be the zero value.
	RootID NodeID
}

// NodeSetInfo holds metadata about a set of nodes (either leaves or branches) in a specific version.
type NodeSetInfo struct {
	// StartOffset is the starting 0-based offset of the node set for this version in the corresponding nodes file.
	// This is not a bytes offset but an index into the virtual array of nodes in the file.
	StartOffset uint32

	// Count is the number of nodes in this set for this version.
	Count uint32

	// StartIndex is the starting index of this node set in the changeset.
	// If this is an uncompacted changeset, this will always be 1 (indexes start at 1).
	// When this is a compacted changeset, this will help reduce the search space when looking for nodes.
	StartIndex uint32

	// EndIndex is the ending index of this node set in the changeset.
	// If this is an uncompacted changeset, this will always be equal to Count (indexes start at 1).
	// When this is a compacted changeset, this will help reduce the search space when looking for nodes.
	EndIndex uint32
}
