package testnet_test

import (
	"testing"

	"github.com/cometbft/cometbft/node"
	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/testutil/testnet"
)

// Nil entries in a Nodes slice don't fail Stop or Wait.
func TestNodes_StopWaitNil(t *testing.T) {
	for _, tc := range []struct {
		Name  string
		Nodes []*node.Node
	}{
		{Name: "nil slice", Nodes: nil},
		{Name: "slice with nil elements", Nodes: []*node.Node{nil}},
	} {
		ns := testnet.Nodes(tc.Nodes)
		t.Run(tc.Name, func(t *testing.T) {
			require.NoError(t, ns.Stop())

			// Nothing special to assert about Wait().
			// It should return immediately, without panicking.
			ns.Wait()
		})
	}
}
