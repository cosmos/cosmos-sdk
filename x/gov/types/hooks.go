package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var _ GovHooks = MultiGovHooks{}

// MultiGovHooks combine multiple governance hooks, all hook functions are run in array sequence
type MultiGovHooks []GovHooks

// NewMultiGovHooks creates a new MultiGovHooks instance
func NewMultiGovHooks(hooks ...GovHooks) MultiGovHooks {
	return hooks
}

// AfterProposalSubmission implements GovHooks.AfterProposalSubmission. It iterates over all the
// registered hooks and calls AfterProposalSubmission
func (h MultiGovHooks) AfterProposalSubmission(ctx sdk.Context, proposalID uint64) error {
	for i := range h {
		return h[i].AfterProposalSubmission(ctx, proposalID)
	}
	return nil
}

// AfterProposalDeposit implements GovHooks.AfterProposalDeposit. It iterates over all the
// registered hooks and calls AfterProposalDeposit
func (h MultiGovHooks) AfterProposalDeposit(ctx sdk.Context, proposalID uint64, depositorAddr sdk.AccAddress) error {
	for i := range h {
		return h[i].AfterProposalDeposit(ctx, proposalID, depositorAddr)
	}
	return nil
}

// AfterProposalVote implements GovHooks.AfterProposalVote. It iterates over all the
// registered hooks and calls AfterProposalVote
func (h MultiGovHooks) AfterProposalVote(ctx sdk.Context, proposalID uint64, voterAddr sdk.AccAddress) error {
	for i := range h {
		return h[i].AfterProposalVote(ctx, proposalID, voterAddr)
	}
	return nil
}

// AfterProposalVotingPeriodEnded implements GovHooks.AfterProposalVotingPeriodEnded. It iterates over all the
// registered hooks and calls AfterProposalFailedMinDeposit
func (h MultiGovHooks) AfterProposalFailedMinDeposit(ctx sdk.Context, proposalID uint64) error {
	for i := range h {
		return h[i].AfterProposalFailedMinDeposit(ctx, proposalID)
	}
	return nil
}

// AfterProposalVotingPeriodEnded implements GovHooks.AfterProposalVotingPeriodEnded. It iterates over all the
// registered hooks and calls AfterProposalVotingPeriodEnded
func (h MultiGovHooks) AfterProposalVotingPeriodEnded(ctx sdk.Context, proposalID uint64) error {
	for i := range h {
		return h[i].AfterProposalVotingPeriodEnded(ctx, proposalID)
	}
	return nil
}
