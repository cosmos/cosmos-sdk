package iavl

import (
	"fmt"
	"testing"

	db "github.com/cosmos/cosmos-db"
	"github.com/stretchr/testify/require"
)

func TestProofOp(t *testing.T) {
	tree, err := NewMutableTreeWithOpts(db.NewMemDB(), 0, nil, false)
	require.NoError(t, err)
	keys := []byte{0x0a, 0x11, 0x2e, 0x32, 0x50, 0x72, 0x99, 0xa1, 0xe4, 0xf7} // 10 total.
	for _, ikey := range keys {
		key := []byte{ikey}
		tree.Set(key, key)
	}

	testcases := []struct {
		key           byte
		expectPresent bool
	}{
		{0x00, false},
		{0x0a, true},
		{0x0b, false},
		{0x11, true},
		{0x60, false},
		{0x72, true},
		{0x99, true},
		{0xaa, false},
		{0xe4, true},
		{0xf7, true},
		{0xff, false},
	}

	for _, tc := range testcases {
		tc := tc
		t.Run(fmt.Sprintf("%02x", tc.key), func(t *testing.T) {
			key := []byte{tc.key}
			if tc.expectPresent {
				proof, err := tree.GetMembershipProof(key)
				require.NoError(t, err)

				// Verify that proof is valid.
				res, err := tree.VerifyMembership(proof, key)
				require.NoError(t, err)
				require.True(t, res)
			} else {
				proof, err := tree.GetNonMembershipProof(key)
				require.NoError(t, err)

				// Verify that proof is valid.
				res, err := tree.VerifyNonMembership(proof, key)
				require.NoError(t, err)
				require.True(t, res)
			}
		})
	}
}
