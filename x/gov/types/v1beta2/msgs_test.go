package v1beta2_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
	"github.com/cosmos/cosmos-sdk/x/gov/types/v1beta2"
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
	msg := v1beta2.NewMsgDeposit(addr, 0, coinsPos)
	res := msg.GetSignBytes()

	expected := `{"type":"cosmos-sdk/v1beta2/MsgDeposit","value":{"amount":[{"amount":"1000","denom":"stake"}],"depositor":"cosmos1v9jxgu33kfsgr5","proposal_id":"0"}}`
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
		msg := v1beta2.NewMsgDeposit(tc.depositorAddr, tc.proposalID, tc.depositAmount)
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
		option     v1beta2.VoteOption
		expectPass bool
	}{
		{0, addrs[0], v1beta2.OptionYes, true},
		{0, sdk.AccAddress{}, v1beta2.OptionYes, false},
		{0, addrs[0], v1beta2.OptionNo, true},
		{0, addrs[0], v1beta2.OptionNoWithVeto, true},
		{0, addrs[0], v1beta2.OptionAbstain, true},
		{0, addrs[0], v1beta2.VoteOption(0x13), false},
	}

	for i, tc := range tests {
		msg := v1beta2.NewMsgVote(tc.voterAddr, tc.proposalID, tc.option)
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
		options    v1beta2.WeightedVoteOptions
		expectPass bool
	}{
		{0, addrs[0], v1beta2.NewNonSplitVoteOption(v1beta2.OptionYes), true},
		{0, sdk.AccAddress{}, v1beta2.NewNonSplitVoteOption(v1beta2.OptionYes), false},
		{0, addrs[0], v1beta2.NewNonSplitVoteOption(v1beta2.OptionNo), true},
		{0, addrs[0], v1beta2.NewNonSplitVoteOption(v1beta2.OptionNoWithVeto), true},
		{0, addrs[0], v1beta2.NewNonSplitVoteOption(v1beta2.OptionAbstain), true},
		{0, addrs[0], v1beta2.WeightedVoteOptions{ // weight sum > 1
			v1beta2.NewWeightedVoteOption(v1beta2.OptionYes, sdk.NewDec(1)),
			v1beta2.NewWeightedVoteOption(v1beta2.OptionAbstain, sdk.NewDec(1)),
		}, false},
		{0, addrs[0], v1beta2.WeightedVoteOptions{ // duplicate option
			v1beta2.NewWeightedVoteOption(v1beta2.OptionYes, sdk.NewDecWithPrec(5, 1)),
			v1beta2.NewWeightedVoteOption(v1beta2.OptionYes, sdk.NewDecWithPrec(5, 1)),
		}, false},
		{0, addrs[0], v1beta2.WeightedVoteOptions{ // zero weight
			v1beta2.NewWeightedVoteOption(v1beta2.OptionYes, sdk.NewDec(0)),
		}, false},
		{0, addrs[0], v1beta2.WeightedVoteOptions{ // negative weight
			v1beta2.NewWeightedVoteOption(v1beta2.OptionYes, sdk.NewDec(-1)),
		}, false},
		{0, addrs[0], v1beta2.WeightedVoteOptions{}, false},
		{0, addrs[0], v1beta2.NewNonSplitVoteOption(v1beta2.VoteOption(0x13)), false},
		{0, addrs[0], v1beta2.WeightedVoteOptions{ // weight sum <1
			v1beta2.NewWeightedVoteOption(v1beta2.OptionYes, sdk.NewDecWithPrec(5, 1)),
		}, false},
	}

	for i, tc := range tests {
		msg := v1beta2.NewMsgVoteWeighted(tc.voterAddr, tc.proposalID, tc.options)
		if tc.expectPass {
			require.Nil(t, msg.ValidateBasic(), "test: %v", i)
		} else {
			require.NotNil(t, msg.ValidateBasic(), "test: %v", i)
		}
	}
}

func TestMsgSubmitProposal_ValidateBasic(t *testing.T) {
	metadata := []byte{42}
	// Valid msg
	msg1, err := v1beta2.NewLegacyContent(v1beta1.NewTextProposal("Title", "description"), addrs[0].String())
	require.NoError(t, err)
	// Invalid msg
	msg2, err := v1beta2.NewLegacyContent(v1beta1.NewTextProposal("Title", "description"), "foo")
	require.NoError(t, err)

	tests := []struct {
		name           string
		proposer       string
		initialDeposit sdk.Coins
		messages       []sdk.Msg
		metadata       []byte
		expErr         bool
	}{
		{"invalid addr", "", coinsPos, []sdk.Msg{msg1}, metadata, true},
		{"empty msgs and metadata", addrs[0].String(), coinsPos, nil, nil, true},
		{"invalid msg", addrs[0].String(), coinsPos, []sdk.Msg{msg1, msg2}, metadata, true},
		{"valid with no Msg", addrs[0].String(), coinsPos, nil, metadata, false},
		{"valid with no metadata", addrs[0].String(), coinsPos, []sdk.Msg{msg1}, nil, false},
		{"valid with everything", addrs[0].String(), coinsPos, []sdk.Msg{msg1}, metadata, false},
	}

	for _, tc := range tests {
		msg, err := v1beta2.NewMsgSubmitProposal(tc.messages, tc.initialDeposit, tc.proposer, tc.metadata)
		require.NoError(t, err)
		if tc.expErr {
			require.Error(t, msg.ValidateBasic(), "test: %s", tc.name)
		} else {
			require.NoError(t, msg.ValidateBasic(), "test: %s", tc.name)
		}
	}
}

// this tests that Amino JSON MsgSubmitProposal.GetSignBytes() still works with Content as Any using the ModuleCdc
func TestMsgSubmitProposal_GetSignBytes(t *testing.T) {
	proposal := []sdk.Msg{v1beta2.NewMsgVote(addrs[0], 1, v1beta2.OptionYes)}
	msg, err := v1beta2.NewMsgSubmitProposal(proposal, sdk.NewCoins(), sdk.AccAddress{}.String(), nil)
	require.NoError(t, err)
	var bz []byte
	require.NotPanics(t, func() {
		bz = msg.GetSignBytes()
	})
	require.Equal(t, "{\"type\":\"cosmos-sdk/v1beta2/MsgSubmitProposal\",\"value\":{\"initial_deposit\":[],\"messages\":[{\"type\":\"cosmos-sdk/v1beta2/MsgVote\",\"value\":{\"option\":1,\"proposal_id\":\"1\",\"voter\":\"cosmos1w3jhxap3gempvr\"}}]}}",
		string(bz))
}
