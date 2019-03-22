package gov

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/gov/proposal"
	"github.com/cosmos/cosmos-sdk/x/mock"
)

var (
	coinsPos         = sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 1000))
	coinsZero        = sdk.NewCoins()
	coinsPosNotAtoms = sdk.NewCoins(sdk.NewInt64Coin("foo", 10000))
	coinsMulti       = sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 1000), sdk.NewInt64Coin("foo", 10000))
)

func init() {
	coinsMulti.Sort()
}

// test ValidateBasic for MsgCreateValidator
func TestMsgSubmitProposal(t *testing.T) {
	_, addrs, _, _ := mock.CreateGenAccounts(1, sdk.NewCoins())
	tests := []struct {
		title, description string
		proposalType       string
		proposerAddr       sdk.AccAddress
		initialDeposit     sdk.Coins
		expectPass         bool
	}{
		{"Test Proposal", "the purpose of this proposal is to test", ProposalTypeText, addrs[0], coinsPos, true},
		{"", "the purpose of this proposal is to test", ProposalTypeText, addrs[0], coinsPos, false},
		{"Test Proposal", "", ProposalTypeText, addrs[0], coinsPos, false},
		{"Test Proposal", "the purpose of this proposal is to test", ProposalTypeSoftwareUpgrade, addrs[0], coinsPos, true},
		{"Test Proposal", "the purpose of this proposal is to test", "ProposalTypeInvalid", addrs[0], coinsPos, false},
		{"Test Proposal", "the purpose of this proposal is to test", ProposalTypeText, sdk.AccAddress{}, coinsPos, false},
		{"Test Proposal", "the purpose of this proposal is to test", ProposalTypeText, addrs[0], coinsZero, true},
		{"Test Proposal", "the purpose of this proposal is to test", ProposalTypeText, addrs[0], coinsMulti, true},
		{strings.Repeat("#", proposal.MaxTitleLength*2), "the purpose of this proposal is to test", ProposalTypeText, addrs[0], coinsMulti, false},
		{"Test Proposal", strings.Repeat("#", proposal.MaxDescriptionLength*2), ProposalTypeText, addrs[0], coinsMulti, false},
	}

	for i, tc := range tests {
		msg := NewMsgSubmitProposal(tc.title, tc.description, tc.proposalType, tc.proposerAddr, tc.initialDeposit)
		if tc.expectPass {
			require.NoError(t, msg.ValidateBasic(), "test: %v", i)
		} else {
			require.Error(t, msg.ValidateBasic(), "test: %v", i)
		}
	}
}

func TestMsgDepositGetSignBytes(t *testing.T) {
	addr := sdk.AccAddress("addr1")
	msg := NewMsgDeposit(addr, 0, coinsPos)
	res := msg.GetSignBytes()

	expected := `{"type":"cosmos-sdk/MsgDeposit","value":{"amount":[{"amount":"1000","denom":"stake"}],"depositor":"cosmos1v9jxgu33kfsgr5","proposal_id":"0"}}`
	require.Equal(t, expected, string(res))
}

// test ValidateBasic for MsgDeposit
func TestMsgDeposit(t *testing.T) {
	_, addrs, _, _ := mock.CreateGenAccounts(1, sdk.NewCoins())
	tests := []struct {
		proposalID    uint64
		depositorAddr sdk.AccAddress
		depositAmount sdk.Coins
		expectPass    bool
	}{
		{startingProposalID - 1, addrs[0], coinsPos, false},
		{startingProposalID, addrs[0], coinsPos, true},
		{startingProposalID + 1, sdk.AccAddress{}, coinsPos, false},
		{startingProposalID + 1, addrs[0], coinsZero, true},
		{startingProposalID + 1, addrs[0], coinsMulti, true},
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

// test ValidateBasic for MsgDeposit
func TestMsgVote(t *testing.T) {
	_, addrs, _, _ := mock.CreateGenAccounts(1, sdk.NewCoins())
	tests := []struct {
		proposalID uint64
		voterAddr  sdk.AccAddress
		option     VoteOption
		expectPass bool
	}{
		{startingProposalID, addrs[0], OptionYes, true},
		{startingProposalID - 1, addrs[0], OptionYes, false},
		{startingProposalID, sdk.AccAddress{}, OptionYes, false},
		{startingProposalID, addrs[0], OptionNo, true},
		{startingProposalID, addrs[0], OptionNoWithVeto, true},
		{startingProposalID, addrs[0], OptionAbstain, true},
		{startingProposalID, addrs[0], VoteOption(0x13), false},
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
