package group_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/simapp"
	"github.com/cosmos/cosmos-sdk/x/group"
)

// TestGogoUnmarshalProposal tests some weird behavior in gogoproto
// unmarshalling.
// This test serves as a showcase that we need to be careful when unmarshalling
// multiple times into the same reference.
func TestGogoUnmarshalProposal(t *testing.T) {
	cdc := simapp.MakeTestEncodingConfig().Codec
	p1 := group.Proposal{Proposers: []string{"foo"}}
	p2 := group.Proposal{Proposers: []string{"bar"}}

	p1Bz, err := cdc.Marshal(&p1)
	require.NoError(t, err)
	p2Bz, err := cdc.Marshal(&p2)
	require.NoError(t, err)

	var p group.Proposal
	err = cdc.Unmarshal(p1Bz, &p)
	require.NoError(t, err)
	err = cdc.Unmarshal(p2Bz, &p)
	require.NoError(t, err)

	// One would expect that unmarshalling into the same `&p` reference would
	// clear the previous `p` value. But it seems that (at least for array
	// fields), the values are not replaced, but concatenated, which
	// is not an intuitive behavior.
	require.Len(t, p.Proposers, 2)
}
