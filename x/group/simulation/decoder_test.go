package simulation_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"cosmossdk.io/x/group"
	"cosmossdk.io/x/group/internal/orm"
	"cosmossdk.io/x/group/keeper"
	"cosmossdk.io/x/group/module"
	"cosmossdk.io/x/group/simulation"

	codectestutil "github.com/cosmos/cosmos-sdk/codec/testutil"
	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	"github.com/cosmos/cosmos-sdk/types/kv"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
)

func TestDecodeStore(t *testing.T) {
	encodingConfig := moduletestutil.MakeTestEncodingConfig(codectestutil.CodecOptions{}, module.AppModule{})
	cdc := encodingConfig.Codec
	dec := simulation.NewDecodeStore(cdc)

	g := group.GroupInfo{Id: 1}
	groupBz, err := cdc.Marshal(&g)
	require.NoError(t, err)

	_, _, addr := testdata.KeyTestPubAddr()
	member := group.GroupMember{GroupId: 1, Member: &group.Member{
		Address: addr.String(),
	}}
	memberBz, err := cdc.Marshal(&member)
	require.NoError(t, err)

	_, _, accAddr := testdata.KeyTestPubAddr()
	acc := group.GroupPolicyInfo{Address: accAddr.String()}
	accBz, err := cdc.Marshal(&acc)
	require.NoError(t, err)

	proposal := group.Proposal{Id: 1}
	proposalBz, err := cdc.Marshal(&proposal)
	require.NoError(t, err)

	vote := group.Vote{Voter: addr.String(), ProposalId: 1}
	voteBz, err := cdc.Marshal(&vote)
	require.NoError(t, err)

	kvPairs := kv.Pairs{
		Pairs: []kv.Pair{
			{Key: append([]byte{keeper.GroupTablePrefix}, orm.PrimaryKey(&g)...), Value: groupBz},
			{Key: append([]byte{keeper.GroupMemberTablePrefix}, orm.PrimaryKey(&member)...), Value: memberBz},
			{Key: append([]byte{keeper.GroupPolicyTablePrefix}, orm.PrimaryKey(&acc)...), Value: accBz},
			{Key: append([]byte{keeper.ProposalTablePrefix}, orm.PrimaryKey(&proposal)...), Value: proposalBz},
			{Key: append([]byte{keeper.VoteTablePrefix}, orm.PrimaryKey(&vote)...), Value: voteBz},
			{Key: []byte{0x99}, Value: []byte{0x99}},
		},
	}

	tests := []struct {
		name        string
		expectErr   bool
		expectedLog string
	}{
		{"Group", false, fmt.Sprintf("%v\n%v", g, g)},
		{"GroupMember", false, fmt.Sprintf("%v\n%v", member, member)},
		{"GroupPolicy", false, fmt.Sprintf("%v\n%v", acc, acc)},
		{"Proposal", false, fmt.Sprintf("%v\n%v", proposal, proposal)},
		{"Vote", false, fmt.Sprintf("%v\n%v", vote, vote)},
		{"other", true, ""},
	}

	for i, tt := range tests {
		i, tt := i, tt
		t.Run(tt.name, func(t *testing.T) {
			if tt.expectErr {
				require.Panics(t, func() { dec(kvPairs.Pairs[i], kvPairs.Pairs[i]) }, tt.name)
			} else {
				require.Equal(t, tt.expectedLog, dec(kvPairs.Pairs[i], kvPairs.Pairs[i]), tt.name)
			}
		})
	}
}
