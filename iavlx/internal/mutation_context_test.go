package internal

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMutationContext_AddOrphan(t *testing.T) {
	tests := []struct {
		name         string
		currentVer   uint32
		nodePtr      *NodePointer
		expectOrphan bool
	}{
		{
			name:         "node with valid checkpoint becomes orphan",
			currentVer:   5,
			nodePtr:      &NodePointer{id: NewNodeID(true, 3, 10)},
			expectOrphan: true,
		},
		{
			name:         "node with same-version checkpoint becomes orphan",
			currentVer:   5,
			nodePtr:      &NodePointer{id: NewNodeID(true, 5, 10)},
			expectOrphan: true,
		},
		{
			name:       "node with checkpoint 0 not orphan",
			currentVer: 5,
			nodePtr:    &NodePointer{id: NewNodeID(true, 0, 10)},
		},
		{
			name:       "nil pointer not orphan",
			currentVer: 5,
			nodePtr:    nil,
		},
		{
			name:       "empty node ID not orphan",
			currentVer: 5,
			nodePtr:    &NodePointer{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := NewMutationContext(tt.currentVer, tt.currentVer)
			added := ctx.addOrphan(tt.nodePtr)
			require.Equal(t, tt.expectOrphan, added, "addOrphan return value mismatch")
			if tt.expectOrphan {
				require.Len(t, ctx.orphans, 1, "expected exactly one orphan")
				require.Same(t, tt.nodePtr, ctx.orphans[0], "orphaned pointer mismatch")
			} else {
				require.Len(t, ctx.orphans, 0, "expected no orphans")
			}
		})
	}
}

func TestMutationContext_MutateNode(t *testing.T) {
	tests := []struct {
		name       string
		currentVer uint32
		cowVersion uint32
		node       *MemNode
		expectSame bool
	}{
		{
			name:       "node version below cowVersion gets copied and orphaned",
			currentVer: 5,
			cowVersion: 5,
			node:       newTestBranchNode(3, 1, newTestLeafNode(1, 1, "A"), newTestLeafNode(2, 2, "B")),
			expectSame: false,
		},
		{
			name:       "node version at cowVersion is reused in place",
			currentVer: 5,
			cowVersion: 3,
			node:       newTestBranchNode(3, 1, newTestLeafNode(1, 1, "A"), newTestLeafNode(2, 2, "B")),
			expectSame: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := NewMutationContext(tt.currentVer, tt.cowVersion)
			nodePtr := NewNodePointer(tt.node)
			nodePtr.id = tt.node.nodeId
			mutatedNode, err := ctx.mutateBranch(tt.node, nodePtr)
			require.NoError(t, err)
			if tt.expectSame {
				require.Same(t, tt.node, mutatedNode, "expected same node instance")
				require.Len(t, ctx.orphans, 0, "expected no orphans")
			} else {
				require.NotSame(t, tt.node, mutatedNode, "expected different node instance")
				require.Equal(t, tt.currentVer, mutatedNode.version, "mutated node layer mismatch")
				require.Len(t, ctx.orphans, 1, "expected one orphan")
				require.Equal(t, tt.node.ID(), ctx.orphans[0].id, "orphaned NodeID mismatch")
			}
		})
	}
}
