package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	govv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	"github.com/cosmos/cosmos-sdk/x/sanction"
)

// Define a Mock Gov Keeper that records calls to GetProposal and allows
// definition of what it returns.

type MockGovKeeper struct {
	GetProposalCalls   []uint64
	GetProposalReturns map[uint64]govv1.Proposal
}

var _ sanction.GovKeeper = &MockGovKeeper{}

func NewMockGovKeeper() *MockGovKeeper {
	return &MockGovKeeper{
		GetProposalCalls:   nil,
		GetProposalReturns: make(map[uint64]govv1.Proposal),
	}
}

func (k *MockGovKeeper) GetProposal(_ sdk.Context, proposalID uint64) (govv1.Proposal, bool) {
	k.GetProposalCalls = append(k.GetProposalCalls, proposalID)
	prop, ok := k.GetProposalReturns[proposalID]
	return prop, ok
}

func (k *MockGovKeeper) GetDepositParams(_ sdk.Context) govv1.DepositParams {
	return govv1.DefaultParams().DepositParams
}

func (k *MockGovKeeper) GetVotingParams(_ sdk.Context) govv1.VotingParams {
	return govv1.DefaultParams().VotingParams
}

func (k *MockGovKeeper) GetProposalID(_ sdk.Context) (uint64, error) {
	return 1, nil
}
