package keeper

import (
	"context"
	stderrors "errors"
	"fmt"

	"cosmossdk.io/collections"
	"cosmossdk.io/errors"
	"cosmossdk.io/x/gov/types"
	v1 "cosmossdk.io/x/gov/types/v1"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// AddVote adds a vote on a specific proposal
func (k Keeper) AddVote(ctx context.Context, proposalID uint64, voterAddr sdk.AccAddress, options v1.WeightedVoteOptions, metadata string) error {
	// Check if proposal is in voting period.
	inVotingPeriod, err := k.VotingPeriodProposals.Has(ctx, proposalID)
	if err != nil {
		return err
	}

	if !inVotingPeriod {
		return errors.Wrapf(types.ErrInactiveProposal, "%d", proposalID)
	}

	if err := k.assertMetadataLength(metadata); err != nil {
		return err
	}

	// get proposal
	proposal, err := k.Proposals.Get(ctx, proposalID)
	if err != nil {
		return err
	}

	for _, option := range options {
		switch proposal.ProposalType {
		case v1.ProposalType_PROPOSAL_TYPE_OPTIMISTIC:
			if option.Option != v1.OptionNo && option.Option != v1.OptionSpam {
				return errors.Wrap(types.ErrInvalidVote, "optimistic proposals can only be rejected")
			}
		case v1.ProposalType_PROPOSAL_TYPE_MULTIPLE_CHOICE:
			proposalOptionsStr, err := k.ProposalVoteOptions.Get(ctx, proposalID)
			if err != nil {
				if stderrors.Is(err, collections.ErrNotFound) {
					return errors.Wrap(types.ErrInvalidProposal, "invalid multiple choice proposal, no options set")
				}

				return err
			}

			// verify votes only on existing votes
			if proposalOptionsStr.OptionOne == "" && option.Option == v1.OptionOne { // should never trigger option one is always mandatory
				return errors.Wrap(types.ErrInvalidVote, "invalid vote option")
			} else if proposalOptionsStr.OptionTwo == "" && option.Option == v1.OptionTwo { // should never trigger option two is always mandatory
				return errors.Wrap(types.ErrInvalidVote, "invalid vote option")
			} else if proposalOptionsStr.OptionThree == "" && option.Option == v1.OptionThree {
				return errors.Wrap(types.ErrInvalidVote, "invalid vote option")
			} else if proposalOptionsStr.OptionFour == "" && option.Option == v1.OptionFour {
				return errors.Wrap(types.ErrInvalidVote, "invalid vote option")
			}
		}

		if !v1.ValidWeightedVoteOption(*option) {
			return errors.Wrap(types.ErrInvalidVote, option.String())
		}
	}

	vote := v1.NewVote(proposalID, voterAddr, options, metadata)
	err = k.Votes.Set(ctx, collections.Join(proposalID, voterAddr), vote)
	if err != nil {
		return err
	}

	// called after a vote on a proposal is cast
	err = k.Hooks().AfterProposalVote(ctx, proposalID, voterAddr)
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
func (k Keeper) deleteVotes(ctx context.Context, proposalID uint64) error {
	rng := collections.NewPrefixedPairRange[uint64, sdk.AccAddress](proposalID)
	err := k.Votes.Clear(ctx, rng)
	if err != nil {
		return err
	}

	return nil
}
