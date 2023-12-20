package keeper

import (
	"context"
	"fmt"

	"cosmossdk.io/collections"
	"cosmossdk.io/errors"
	"cosmossdk.io/x/gov/types"
	v1 "cosmossdk.io/x/gov/types/v1"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// AddVote adds a vote on a specific proposal
func (keeper Keeper) AddVote(ctx context.Context, proposalID uint64, voterAddr sdk.AccAddress, options v1.WeightedVoteOptions, metadata string) error {
	// Check if proposal is in voting period.
	inVotingPeriod, err := keeper.VotingPeriodProposals.Has(ctx, proposalID)
	if err != nil {
		return err
	}

	if !inVotingPeriod {
		return errors.Wrapf(types.ErrInactiveProposal, "%d", proposalID)
	}

	if err := keeper.assertMetadataLength(metadata); err != nil {
		return err
	}

	// get proposal
	proposal, err := keeper.Proposals.Get(ctx, proposalID)
	if err != nil {
		return err
	}

	for _, option := range options {
		switch proposal.ProposalType {
		case v1.ProposalType_PROPOSAL_TYPE_OPTIMISTIC:
			if option.Option != v1.OptionNo && option.Option != v1.OptionSpam {
				return errors.Wrap(types.ErrInvalidVote, "optimistic proposals can only be rejected")
			}
		default:
			if !v1.ValidWeightedVoteOption(*option) {
				return errors.Wrap(types.ErrInvalidVote, option.String())
			}
		}
	}

	vote := v1.NewVote(proposalID, voterAddr, options, metadata)
	err = keeper.Votes.Set(ctx, collections.Join(proposalID, voterAddr), vote)
	if err != nil {
		return err
	}

	// called after a vote on a proposal is cast
	err = keeper.Hooks().AfterProposalVote(ctx, proposalID, voterAddr)
	if err != nil {
		return err
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	sdkCtx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeProposalVote,
			sdk.NewAttribute(types.AttributeKeyVoter, voterAddr.String()),
			sdk.NewAttribute(types.AttributeKeyOption, options.String()),
			sdk.NewAttribute(types.AttributeKeyProposalID, fmt.Sprintf("%d", proposalID)),
		),
	)

	return nil
}

// deleteVotes deletes all the votes from a given proposalID.
func (keeper Keeper) deleteVotes(ctx context.Context, proposalID uint64) error {
	rng := collections.NewPrefixedPairRange[uint64, sdk.AccAddress](proposalID)
	err := keeper.Votes.Clear(ctx, rng)
	if err != nil {
		return err
	}

	return nil
}
