package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/gov/types"
	"github.com/cosmos/cosmos-sdk/x/gov/types/v1beta2"
)

// AddVote adds a vote on a specific proposal
func (keeper Keeper) AddVote(ctx sdk.Context, proposalID uint64, voterAddr sdk.AccAddress, options v1beta2.WeightedVoteOptions) error {
	proposal, ok := keeper.GetProposal(ctx, proposalID)
	if !ok {
		return sdkerrors.Wrapf(types.ErrUnknownProposal, "%d", proposalID)
	}
	if proposal.Status != v1beta2.StatusVotingPeriod {
		return sdkerrors.Wrapf(types.ErrInactiveProposal, "%d", proposalID)
	}

	for _, option := range options {
		if !v1beta2.ValidWeightedVoteOption(*option) {
			return sdkerrors.Wrap(types.ErrInvalidVote, option.String())
		}
	}

	vote := v1beta2.NewVote(proposalID, voterAddr, options)
	keeper.SetVote(ctx, vote)

	// called after a vote on a proposal is cast
	keeper.AfterProposalVote(ctx, proposalID, voterAddr)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeProposalVote,
			sdk.NewAttribute(types.AttributeKeyOption, options.String()),
			sdk.NewAttribute(types.AttributeKeyProposalID, fmt.Sprintf("%d", proposalID)),
		),
	)

	return nil
}

// GetAllVotes returns all the votes from the store
func (keeper Keeper) GetAllVotes(ctx sdk.Context) (votes v1beta2.Votes) {
	keeper.IterateAllVotes(ctx, func(vote v1beta2.Vote) bool {
		votes = append(votes, &vote)
		return false
	})
	return
}

// GetVotes returns all the votes from a proposal
func (keeper Keeper) GetVotes(ctx sdk.Context, proposalID uint64) (votes v1beta2.Votes) {
	keeper.IterateVotes(ctx, proposalID, func(vote v1beta2.Vote) bool {
		votes = append(votes, &vote)
		return false
	})
	return
}

// GetVote gets the vote from an address on a specific proposal
func (keeper Keeper) GetVote(ctx sdk.Context, proposalID uint64, voterAddr sdk.AccAddress) (vote v1beta2.Vote, found bool) {
	store := ctx.KVStore(keeper.storeKey)
	bz := store.Get(types.VoteKey(proposalID, voterAddr))
	if bz == nil {
		return vote, false
	}

	keeper.cdc.MustUnmarshal(bz, &vote)

	return vote, true
}

// SetVote sets a Vote to the gov store
func (keeper Keeper) SetVote(ctx sdk.Context, vote v1beta2.Vote) {
	store := ctx.KVStore(keeper.storeKey)
	bz := keeper.cdc.MustMarshal(&vote)
	addr, err := sdk.AccAddressFromBech32(vote.Voter)
	if err != nil {
		panic(err)
	}
	store.Set(types.VoteKey(vote.ProposalId, addr), bz)
}

// IterateAllVotes iterates over the all the stored votes and performs a callback function
func (keeper Keeper) IterateAllVotes(ctx sdk.Context, cb func(vote v1beta2.Vote) (stop bool)) {
	store := ctx.KVStore(keeper.storeKey)
	iterator := sdk.KVStorePrefixIterator(store, types.VotesKeyPrefix)

	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		var vote v1beta2.Vote
		keeper.cdc.MustUnmarshal(iterator.Value(), &vote)

		if cb(vote) {
			break
		}
	}
}

// IterateVotes iterates over the all the proposals votes and performs a callback function
func (keeper Keeper) IterateVotes(ctx sdk.Context, proposalID uint64, cb func(vote v1beta2.Vote) (stop bool)) {
	store := ctx.KVStore(keeper.storeKey)
	iterator := sdk.KVStorePrefixIterator(store, types.VotesKey(proposalID))

	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		var vote v1beta2.Vote
		keeper.cdc.MustUnmarshal(iterator.Value(), &vote)

		if cb(vote) {
			break
		}
	}
}

// deleteVote deletes a vote from a given proposalID and voter from the store
func (keeper Keeper) deleteVote(ctx sdk.Context, proposalID uint64, voterAddr sdk.AccAddress) {
	store := ctx.KVStore(keeper.storeKey)
	store.Delete(types.VoteKey(proposalID, voterAddr))
}
