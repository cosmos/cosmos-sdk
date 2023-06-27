package keeper

import (
	"context"
	"fmt"

	"cosmossdk.io/collections"
	"cosmossdk.io/errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/gov/types"
	v1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
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

	err = keeper.assertMetadataLength(metadata)
	if err != nil {
		return err
	}

	for _, option := range options {
		if !v1.ValidWeightedVoteOption(*option) {
			return errors.Wrap(types.ErrInvalidVote, option.String())
		}
	}

	vote := v1.NewVote(proposalID, voterAddr, options, metadata)
	err = keeper.Votes.Set(ctx, collections.Join(proposalID, voterAddr), vote)
	if err != nil {
		return err
	}

	// called after a vote on a proposal is cast
	keeper.Hooks().AfterProposalVote(ctx, proposalID, voterAddr)

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	sdkCtx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeProposalVote,
			sdk.NewAttribute(types.AttributeKeyOption, options.String()),
			sdk.NewAttribute(types.AttributeKeyProposalID, fmt.Sprintf("%d", proposalID)),
		),
	)

	return nil
}

// deleteVotes deletes the all votes from a given proposalID.
func (keeper Keeper) deleteVotes(ctx context.Context, proposalID uint64) error {
	// TODO(tip): fix https://github.com/cosmos/cosmos-sdk/issues/16162
	return nil
}
