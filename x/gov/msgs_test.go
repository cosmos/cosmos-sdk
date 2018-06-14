package gov

import (
	"testing"

	"github.com/stretchr/testify/assert"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

var (
	coinsPos         = sdk.Coins{sdk.Coin{"steak", 1000}}
	coinsZero        = sdk.Coins{}
	coinsNeg         = sdk.Coins{sdk.Coin{"steak", -10000}}
	coinsPosNotAtoms = sdk.Coins{sdk.Coin{"foo", 10000}}
	coinsMulti       = sdk.Coins{sdk.Coin{"foo", 10000}, sdk.Coin{"steak", 1000}}
)

// test ValidateBasic for MsgCreateValidator
func TestMsgSubmitProposal(t *testing.T) {
	tests := []struct {
		title, description, proposalType string
		proposerAddr                     sdk.Address
		initialDeposit                   sdk.Coins
		expectPass                       bool
	}{
		{"Test Proposal", "the purpose of this proposal is to test", "Text", addrs[0], coinsPos, true},
		{"", "the purpose of this proposal is to test", "Text", addrs[0], coinsPos, false},
		{"Test Proposal", "", "Text", addrs[0], coinsPos, false},
		{"Test Proposal", "the purpose of this proposal is to test", "ParameterChange", addrs[0], coinsPos, true},
		{"Test Proposal", "the purpose of this proposal is to test", "SoftwareUpgrade", addrs[0], coinsPos, true},
		{"Test Proposal", "the purpose of this proposal is to test", "Other", addrs[0], coinsPos, false},
		{"Test Proposal", "the purpose of this proposal is to test", "Text", sdk.Address{}, coinsPos, false},
		{"Test Proposal", "the purpose of this proposal is to test", "Text", addrs[0], coinsZero, true},
		{"Test Proposal", "the purpose of this proposal is to test", "Text", addrs[0], coinsNeg, false},
		{"Test Proposal", "the purpose of this proposal is to test", "Text", addrs[0], coinsMulti, true},
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
	tests := []struct {
		proposalID int64
		voterAddr  sdk.Address
		option     string
		expectPass bool
	}{
		{0, addrs[0], "Yes", true},
		{-1, addrs[0], "Yes", false},
		{0, sdk.Address{}, "Yes", false},
		{0, addrs[0], "No", true},
		{0, addrs[0], "NoWithVeto", true},
		{0, addrs[0], "Abstain", true},
		{0, addrs[0], "Meow", false},
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
