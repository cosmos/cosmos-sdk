package simulation_test

import (
	"encoding/binary"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	"github.com/cosmos/cosmos-sdk/simapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/kv"
	"github.com/cosmos/cosmos-sdk/x/gov/simulation"
	"github.com/cosmos/cosmos-sdk/x/gov/types"
)

var (
	delPk1   = ed25519.GenPrivKey().PubKey()
	delAddr1 = sdk.AccAddress(delPk1.Address())
)

func TestDecodeStore(t *testing.T) {
	cdc := simapp.MakeTestEncodingConfig().Marshaler
	dec := simulation.NewDecodeStore(cdc)

	endTime := time.Now().UTC()
	content := types.ContentFromProposalType("test", "test", types.ProposalTypeText)
	proposalA, err := types.NewProposal(content, 1, endTime, endTime.Add(24*time.Hour))
	require.NoError(t, err)
	proposalB, err := types.NewProposal(content, 2, endTime, endTime.Add(24*time.Hour))
	require.NoError(t, err)

	proposalIDBz := make([]byte, 8)
	binary.LittleEndian.PutUint64(proposalIDBz, 1)
	deposit := types.NewDeposit(1, delAddr1, sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdk.OneInt())))
	vote := types.NewVote(1, delAddr1, types.NewNonSplitVoteOption(types.OptionYes))

	proposalBzA, err := cdc.Marshal(&proposalA)
	require.NoError(t, err)
	proposalBzB, err := cdc.Marshal(&proposalB)
	require.NoError(t, err)

	tests := []struct {
		name        string
		kvA, kvB    kv.Pair
		expectedLog string
		wantPanic   bool
	}{
		{
			"proposals",
			kv.Pair{Key: types.ProposalKey(1), Value: proposalBzA},
			kv.Pair{Key: types.ProposalKey(2), Value: proposalBzB},
			fmt.Sprintf("%v\n%v", proposalA, proposalB), false,
		},
		{
			"proposal IDs",
			kv.Pair{Key: types.InactiveProposalQueueKey(1, endTime), Value: proposalIDBz},
			kv.Pair{Key: types.InactiveProposalQueueKey(1, endTime), Value: proposalIDBz},
			"proposalIDA: 1\nProposalIDB: 1", false,
		},
		{
			"deposits",
			kv.Pair{Key: types.DepositKey(1, delAddr1), Value: cdc.MustMarshal(&deposit)},
			kv.Pair{Key: types.DepositKey(1, delAddr1), Value: cdc.MustMarshal(&deposit)},
			fmt.Sprintf("%v\n%v", deposit, deposit), false,
		},
		{
			"votes",
			kv.Pair{Key: types.VoteKey(1, delAddr1), Value: cdc.MustMarshal(&vote)},
			kv.Pair{Key: types.VoteKey(1, delAddr1), Value: cdc.MustMarshal(&vote)},
			fmt.Sprintf("%v\n%v", vote, vote), false,
		},
		{
			"other",
			kv.Pair{Key: []byte{0x99}, Value: []byte{0x99}},
			kv.Pair{Key: []byte{0x99}, Value: []byte{0x99}},
			"", true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			if tt.wantPanic {
				require.Panics(t, func() { dec(tt.kvA, tt.kvB) }, tt.name)
			} else {
				require.Equal(t, tt.expectedLog, dec(tt.kvA, tt.kvB), tt.name)
			}
		})
	}
}
