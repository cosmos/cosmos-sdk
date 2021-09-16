package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/gov/types"
)

// Implements GovHooks interface
var _ types.GovHooks = Keeper{}

// AfterProposalSubmission - call hook if registered
func (keeper Keeper) AfterProposalSubmission(ctx sdk.Context, proposalID uint64) error {
	if keeper.hooks != nil {
		return keeper.hooks.AfterProposalSubmission(ctx, proposalID)
	}
	return nil
}

// AfterProposalDeposit - call hook if registered
func (keeper Keeper) AfterProposalDeposit(ctx sdk.Context, proposalID uint64, depositorAddr sdk.AccAddress) error {
	if keeper.hooks != nil {
		return keeper.hooks.AfterProposalDeposit(ctx, proposalID, depositorAddr)
	}
	return nil
}

// AfterProposalVote - call hook if registered
func (keeper Keeper) AfterProposalVote(ctx sdk.Context, proposalID uint64, voterAddr sdk.AccAddress) error {
	if keeper.hooks != nil {
		return keeper.hooks.AfterProposalVote(ctx, proposalID, voterAddr)
	}
	return nil
}

// AfterProposalFailedMinDeposit - call hook if registered
func (keeper Keeper) AfterProposalFailedMinDeposit(ctx sdk.Context, proposalID uint64) error {
	if keeper.hooks != nil {
		return keeper.hooks.AfterProposalFailedMinDeposit(ctx, proposalID)
	}
	return nil
}

// AfterProposalVotingPeriodEnded - call hook if registered
func (keeper Keeper) AfterProposalVotingPeriodEnded(ctx sdk.Context, proposalID uint64) error {
	if keeper.hooks != nil {
		return keeper.hooks.AfterProposalVotingPeriodEnded(ctx, proposalID)
	}
	return nil
}
