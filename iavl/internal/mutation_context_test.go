package internal

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMutationContext_AddOrphan(t *testing.T) {
	tests := []struct {
		name         string
		currentVer   uint32
		nodeId       NodeID
		expectOrphan bool
	}{
		{
			name:         "older version becomes orphan",
			currentVer:   5,
			nodeId:       NewNodeID(true, 3, 10),
			expectOrphan: true,
		},
		{
			name:       "same version not orphan",
			currentVer: 5,
			nodeId:     NewNodeID(true, 5, 10),
		},
		{
			name:       "empty node ID not orphan",
			currentVer: 5,
			nodeId:     NodeID{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := &mutationContext{version: tt.currentVer}
			added := ctx.addOrphan(tt.nodeId)
			require.Equal(t, tt.expectOrphan, added, "addOrphan return value mismatch")
			if tt.expectOrphan {
				require.Len(t, ctx.orphans, 1, "expected exactly one orphan")
				require.Equal(t, tt.nodeId, ctx.orphans[0], "orphaned NodeID mismatch")
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
		node       Node
		expectSame bool
	}{
		{
			name:       "mutate existing node creates new node and adds orphan",
			currentVer: 5,
			node:       newTestBranchNode(3, 1, newTestLeafNode(1, 1, "A"), newTestLeafNode(2, 2, "B")),
			expectSame: false,
		},
		{
			name:       "mutate uncommitted node returns same node",
			currentVer: 5,
			node:       newTestBranchNode(0, 0, newTestLeafNode(1, 1, "A"), newTestLeafNode(2, 2, "B")),
			expectSame: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := &mutationContext{version: tt.currentVer}
			mutatedNode, err := ctx.mutateBranch(tt.node)
			require.NoError(t, err)
			if tt.expectSame {
				require.Same(t, tt.node, mutatedNode, "expected same node instance")
				require.Len(t, ctx.orphans, 0, "expected no orphans")
			} else {
				require.NotSame(t, tt.node, mutatedNode, "expected different node instance")
				require.Equal(t, tt.currentVer, mutatedNode.version, "mutated node version mismatch")
				require.Len(t, ctx.orphans, 1, "expected one orphan")
				require.Equal(t, tt.node.ID(), ctx.orphans[0], "orphaned NodeID mismatch")
			}
		})
	}
}
