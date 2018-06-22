package gov

import (
	"testing"

	"github.com/stretchr/testify/assert"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/mock"
)

var (
	coinsPos         = sdk.Coins{sdk.NewCoin("steak", 1000)}
	coinsZero        = sdk.Coins{}
	coinsNeg         = sdk.Coins{sdk.NewCoin("steak", -10000)}
	coinsPosNotAtoms = sdk.Coins{sdk.NewCoin("foo", 10000)}
	coinsMulti       = sdk.Coins{sdk.NewCoin("foo", 10000), sdk.NewCoin("steak", 1000)}
)

// test ValidateBasic for MsgCreateValidator
func TestMsgSubmitProposal(t *testing.T) {
	_, addrs, _, _ := mock.CreateGenAccounts(1, sdk.Coins{})
	tests := []struct {
		title, description string
		proposalType       byte
		proposerAddr       sdk.Address
		initialDeposit     sdk.Coins
		expectPass         bool
	}{
		{"Test Proposal", "the purpose of this proposal is to test", ProposalTypeText, addrs[0], coinsPos, true},
		{"", "the purpose of this proposal is to test", ProposalTypeText, addrs[0], coinsPos, false},
		{"Test Proposal", "", ProposalTypeText, addrs[0], coinsPos, false},
		{"Test Proposal", "the purpose of this proposal is to test", ProposalTypeParameterChange, addrs[0], coinsPos, true},
		{"Test Proposal", "the purpose of this proposal is to test", ProposalTypeSoftwareUpgrade, addrs[0], coinsPos, true},
		{"Test Proposal", "the purpose of this proposal is to test", 0x05, addrs[0], coinsPos, false},
		{"Test Proposal", "the purpose of this proposal is to test", ProposalTypeText, sdk.Address{}, coinsPos, false},
		{"Test Proposal", "the purpose of this proposal is to test", ProposalTypeText, addrs[0], coinsZero, true},
		{"Test Proposal", "the purpose of this proposal is to test", ProposalTypeText, addrs[0], coinsNeg, false},
		{"Test Proposal", "the purpose of this proposal is to test", ProposalTypeText, addrs[0], coinsMulti, true},
	}

	for i, tc := range tests {
		msg := NewMsgSubmitProposal(tc.title, tc.description, tc.proposalType, tc.proposerAddr, tc.initialDeposit)
		if tc.expectPass {
			assert.Nil(t, msg.ValidateBasic(), "test: %v", i)
		} else {
			assert.NotNil(t, msg.ValidateBasic(), "test: %v", i)
		}
	}
}

// test ValidateBasic for MsgDeposit
func TestMsgDeposit(t *testing.T) {
	_, addrs, _, _ := mock.CreateGenAccounts(1, sdk.Coins{})
	tests := []struct {
		proposalID    int64
		depositerAddr sdk.Address
		depositAmount sdk.Coins
		expectPass    bool
	}{
		{0, addrs[0], coinsPos, true},
		{-1, addrs[0], coinsPos, false},
		{1, sdk.Address{}, coinsPos, false},
		{1, addrs[0], coinsZero, true},
		{1, addrs[0], coinsNeg, false},
		{1, addrs[0], coinsMulti, true},
	}

	for i, tc := range tests {
		msg := NewMsgDeposit(tc.depositerAddr, tc.proposalID, tc.depositAmount)
		if tc.expectPass {
			assert.Nil(t, msg.ValidateBasic(), "test: %v", i)
		} else {
			assert.NotNil(t, msg.ValidateBasic(), "test: %v", i)
		}
	}
}

// test ValidateBasic for MsgDeposit
func TestMsgVote(t *testing.T) {
	_, addrs, _, _ := mock.CreateGenAccounts(1, sdk.Coins{})
	tests := []struct {
		proposalID int64
		voterAddr  sdk.Address
		option     VoteOption
		expectPass bool
	}{
		{0, addrs[0], OptionYes, true},
		{-1, addrs[0], OptionYes, false},
		{0, sdk.Address{}, OptionYes, false},
		{0, addrs[0], OptionNo, true},
		{0, addrs[0], OptionNoWithVeto, true},
		{0, addrs[0], OptionAbstain, true},
		{0, addrs[0], VoteOption(0x13), false},
	}

	for i, tc := range tests {
		msg := NewMsgVote(tc.voterAddr, tc.proposalID, tc.option)
		if tc.expectPass {
			assert.Nil(t, msg.ValidateBasic(), "test: %v", i)
		} else {
			assert.NotNil(t, msg.ValidateBasic(), "test: %v", i)
		}
	}
}
