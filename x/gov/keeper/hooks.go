package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/gov/types"
)

// Implements GovHooks interface
var _ types.GovHooks = Keeper{}

// AfterProposalSubmission - call hook if registered
func (keeper Keeper) AfterProposalSubmission(ctx sdk.Context, proposalID uint64) {
	if keeper.hooks != nil {
		keeper.hooks.AfterProposalSubmission(ctx, proposalID)
	}
}

// AfterProposalDeposit - call hook if registered
func (keeper Keeper) AfterProposalDeposit(ctx sdk.Context, proposalID uint64, depositAddr sdk.AccAddress) {
	if keeper.hooks != nil {
		keeper.hooks.AfterProposalDeposit(ctx, proposalID, depositAddr)
	}
}

// AfterProposalVote - call hook if registered
func (keeper Keeper) AfterProposalVote(ctx sdk.Context, proposalID uint64, voterAddr sdk.AccAddress) {
	if keeper.hooks != nil {
		keeper.hooks.AfterProposalVote(ctx, proposalID, voterAddr)
	}
}

// AfterProposalInactive - call hook if registered
func (keeper Keeper) AfterProposalInactive(ctx sdk.Context, proposalID uint64) {
	if keeper.hooks != nil {
		keeper.hooks.AfterProposalInactive(ctx, proposalID)
	}
}

// AfterProposalActive - call hook if registered
func (keeper Keeper) AfterProposalActive(ctx sdk.Context, proposalID uint64) {
	if keeper.hooks != nil {
		keeper.hooks.AfterProposalActive(ctx, proposalID)
	}
}
