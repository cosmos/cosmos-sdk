package gov

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/gov/types"
)

// AddVote Adds a vote on a specific proposal
func (keeper Keeper) AddVote(ctx sdk.Context, proposalID uint64, voterAddr sdk.AccAddress, option VoteOption) sdk.Error {
	proposal, ok := keeper.GetProposal(ctx, proposalID)
	if !ok {
		return ErrUnknownProposal(keeper.codespace, proposalID)
	}
	if proposal.Status != StatusVotingPeriod {
		return ErrInactiveProposal(keeper.codespace, proposalID)
	}

	if !ValidVoteOption(option) {
		return ErrInvalidVote(keeper.codespace, option)
	}

	vote := NewVote(proposalID, voterAddr, option)
	keeper.setVote(ctx, proposalID, voterAddr, vote)

	return nil
}

// GetVotes returns all the votes from the store
func (keeper Keeper) GetVotes(ctx sdk.Context) (votes Votes) {
	store := ctx.KVStore(keeper.storeKey)
	iterator := sdk.KVStorePrefixIterator(store, types.VotesKeyPrefix)

	keeper.IterateVotes(ctx, iterator, func(vote Vote) bool {
		votes = append(votes, vote)
		return false
	})
	return
}

// GetProposalVotes returns all the votes from a proposal
func (keeper Keeper) GetProposalVotes(ctx sdk.Context, proposalID uint64) (votes Votes) {
	store := ctx.KVStore(keeper.storeKey)
	iterator := sdk.KVStorePrefixIterator(store, types.KeyProposalVotes(proposalID))

	keeper.IterateVotes(ctx, iterator, func(vote Vote) bool {
		votes = append(votes, vote)
		return false
	})
	return
}

// GetVote gets the vote from an address on a specific proposal
func (keeper Keeper) GetVote(ctx sdk.Context, proposalID uint64, voterAddr sdk.AccAddress) (vote Vote, found bool) {
	store := ctx.KVStore(keeper.storeKey)
	bz := store.Get(types.KeyProposalVote(proposalID, voterAddr))
	if bz == nil {
		return vote, false
	}

	keeper.cdc.MustUnmarshalBinaryLengthPrefixed(bz, &vote)
	return vote, true
}

func (keeper Keeper) setVote(ctx sdk.Context, proposalID uint64, voterAddr sdk.AccAddress, vote Vote) {
	store := ctx.KVStore(keeper.storeKey)
	bz := keeper.cdc.MustMarshalBinaryLengthPrefixed(vote)
	store.Set(types.KeyProposalVote(proposalID, voterAddr), bz)
}

// GetProposalVotesIterator gets all the votes on a specific proposal as an sdk.Iterator
func (keeper Keeper) GetProposalVotesIterator(ctx sdk.Context, proposalID uint64) sdk.Iterator {
	store := ctx.KVStore(keeper.storeKey)
	return sdk.KVStorePrefixIterator(store, types.KeyProposalVotes(proposalID))
}

func (keeper Keeper) deleteVote(ctx sdk.Context, proposalID uint64, voterAddr sdk.AccAddress) {
	store := ctx.KVStore(keeper.storeKey)
	store.Delete(types.KeyProposalVote(proposalID, voterAddr))
}
