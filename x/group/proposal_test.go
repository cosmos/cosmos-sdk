package group_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"cosmossdk.io/depinject"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/x/group"
	"github.com/cosmos/cosmos-sdk/x/group/testutil"
)

// TestGogoUnmarshalProposal tests some weird behavior in gogoproto
// unmarshalling.
// This test serves as a showcase that we need to be careful when unmarshalling
// multiple times into the same reference.
func TestGogoUnmarshalProposal(t *testing.T) {
	var cdc codec.Codec
	err := depinject.Inject(testutil.AppConfig, &cdc)
	require.NoError(t, err)

	p1 := group.Proposal{Proposers: []string{"foo"}}
	p2 := group.Proposal{Proposers: []string{"bar"}}

	p1Bz, err := cdc.Marshal(&p1)
	require.NoError(t, err)
	p2Bz, err := cdc.Marshal(&p2)
	require.NoError(t, err)

	var p group.Proposal
	err = cdc.Unmarshal(p1Bz, &p)
	require.NoError(t, err)

	var i group.Proposal
	err = cdc.Unmarshal(p2Bz, &i)
	require.NoError(t, err)

	require.Len(t, p.Proposers, 1)
	require.Len(t, i.Proposers, 1)
}
