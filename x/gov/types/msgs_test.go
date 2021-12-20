package types

import (
	"testing"

	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

var (
	coinsPos   = sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 1000))
	coinsZero  = sdk.NewCoins()
	coinsMulti = sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 1000), sdk.NewInt64Coin("foo", 10000))
	addrs      = []sdk.AccAddress{
		sdk.AccAddress("test1"),
		sdk.AccAddress("test2"),
	}
)

func init() {
	coinsMulti.Sort()
}

func TestMsgDepositGetSignBytes(t *testing.T) {
	addr := sdk.AccAddress("addr1")
	msg := NewMsgDeposit(addr, 0, coinsPos)
	res := msg.GetSignBytes()

	expected := `{"type":"cosmos-sdk/MsgDeposit","value":{"amount":[{"amount":"1000","denom":"stake"}],"depositor":"cosmos1v9jxgu33kfsgr5"}}`
	require.Equal(t, expected, string(res))
}

// test ValidateBasic for MsgDeposit
func TestMsgDeposit(t *testing.T) {
	tests := []struct {
		proposalID    uint64
		depositorAddr sdk.AccAddress
		depositAmount sdk.Coins
		expectPass    bool
	}{
		{0, addrs[0], coinsPos, true},
		{1, sdk.AccAddress{}, coinsPos, false},
		{1, addrs[0], coinsZero, true},
		{1, addrs[0], coinsMulti, true},
	}

	for i, tc := range tests {
		msg := NewMsgDeposit(tc.depositorAddr, tc.proposalID, tc.depositAmount)
		if tc.expectPass {
			require.NoError(t, msg.ValidateBasic(), "test: %v", i)
		} else {
			require.Error(t, msg.ValidateBasic(), "test: %v", i)
		}
	}
}

// test ValidateBasic for MsgVote
func TestMsgVote(t *testing.T) {
	tests := []struct {
		proposalID uint64
		voterAddr  sdk.AccAddress
		option     VoteOption
		expectPass bool
	}{
		{0, addrs[0], OptionYes, true},
		{0, sdk.AccAddress{}, OptionYes, false},
		{0, addrs[0], OptionNo, true},
		{0, addrs[0], OptionNoWithVeto, true},
		{0, addrs[0], OptionAbstain, true},
		{0, addrs[0], VoteOption(0x13), false},
	}

	for i, tc := range tests {
		msg := NewMsgVote(tc.voterAddr, tc.proposalID, tc.option)
		if tc.expectPass {
			require.Nil(t, msg.ValidateBasic(), "test: %v", i)
		} else {
			require.NotNil(t, msg.ValidateBasic(), "test: %v", i)
		}
	}
}

// test ValidateBasic for MsgVoteWeighted
func TestMsgVoteWeighted(t *testing.T) {
	tests := []struct {
		proposalID uint64
		voterAddr  sdk.AccAddress
		options    WeightedVoteOptions
		expectPass bool
	}{
		{0, addrs[0], NewNonSplitVoteOption(OptionYes), true},
		{0, sdk.AccAddress{}, NewNonSplitVoteOption(OptionYes), false},
		{0, addrs[0], NewNonSplitVoteOption(OptionNo), true},
		{0, addrs[0], NewNonSplitVoteOption(OptionNoWithVeto), true},
		{0, addrs[0], NewNonSplitVoteOption(OptionAbstain), true},
		{0, addrs[0], WeightedVoteOptions{ // weight sum > 1
			NewWeightedVoteOption(OptionYes, sdk.NewDec(1)),
			NewWeightedVoteOption(OptionAbstain, sdk.NewDec(1)),
		}, false},
		{0, addrs[0], WeightedVoteOptions{ // duplicate option
			NewWeightedVoteOption(OptionYes, sdk.NewDecWithPrec(5, 1)),
			NewWeightedVoteOption(OptionYes, sdk.NewDecWithPrec(5, 1)),
		}, false},
		{0, addrs[0], WeightedVoteOptions{ // zero weight
			NewWeightedVoteOption(OptionYes, sdk.NewDec(0)),
		}, false},
		{0, addrs[0], WeightedVoteOptions{ // negative weight
			NewWeightedVoteOption(OptionYes, sdk.NewDec(-1)),
		}, false},
		{0, addrs[0], WeightedVoteOptions{}, false},
		{0, addrs[0], NewNonSplitVoteOption(VoteOption(0x13)), false},
		{0, addrs[0], WeightedVoteOptions{ // weight sum <1
			NewWeightedVoteOption(OptionYes, sdk.NewDecWithPrec(5, 1)),
		}, false},
	}

	for i, tc := range tests {
		msg := NewMsgVoteWeighted(tc.voterAddr, tc.proposalID, tc.options)
		if tc.expectPass {
			require.Nil(t, msg.ValidateBasic(), "test: %v", i)
		} else {
			require.NotNil(t, msg.ValidateBasic(), "test: %v", i)
		}
	}
}

// this tests that Amino JSON MsgSubmitProposal.GetSignBytes() still works with Content as Any using the ModuleCdc
func TestMsgSubmitProposal_GetSignBytes(t *testing.T) {
	proposal := []sdk.Msg{NewMsgVote(addrs[0], 1, OptionYes)}
	msg, err := NewMsgSubmitProposal(proposal, sdk.NewCoins(), sdk.AccAddress{})
	require.NoError(t, err)
	var bz []byte
	require.NotPanics(t, func() {
		bz = msg.GetSignBytes()
	})
	require.Equal(t, "{\"type\":\"cosmos-sdk/MsgSubmitProposal\",\"value\":{\"initial_deposit\":[],\"messages\":[{\"type\":\"cosmos-sdk/MsgVote\",\"value\":{\"option\":1,\"proposal_id\":\"1\",\"voter\":\"cosmos1w3jhxap3gempvr\"}}]}}",
		string(bz))
}
