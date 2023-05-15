package keeper

import (
	"context"
	"cosmossdk.io/collections"
	"fmt"

	"cosmossdk.io/errors"
	storetypes "cosmossdk.io/store/types"

	"github.com/cosmos/cosmos-sdk/runtime"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/gov/types"
	v1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
)

// AddVote adds a vote on a specific proposal
func (keeper Keeper) AddVote(ctx context.Context, proposalID uint64, voterAddr sdk.AccAddress, options v1.WeightedVoteOptions, metadata string) error {
	// Check if proposal is in voting period.
	store := keeper.storeService.OpenKVStore(ctx)
	inVotingPeriod, err := store.Has(types.VotingPeriodProposalKey(proposalID))
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

// GetAllVotes returns all the votes from the store
func (keeper Keeper) GetAllVotes(ctx context.Context) (votes v1.Votes, err error) {
	err = keeper.IterateAllVotes(ctx, func(vote v1.Vote) error {
		votes = append(votes, &vote)
		return nil
	})
	return
}

// GetVotes returns all the votes from a proposal
func (keeper Keeper) GetVotes(ctx context.Context, proposalID uint64) (votes v1.Votes, err error) {
	err = keeper.IterateVotes(ctx, proposalID, func(vote v1.Vote) error {
		votes = append(votes, &vote)
		return nil
	})
	return
}

// GetVote gets the vote from an address on a specific proposal
func (keeper Keeper) GetVote(ctx context.Context, proposalID uint64, voterAddr sdk.AccAddress) (vote v1.Vote, err error) {
	store := keeper.storeService.OpenKVStore(ctx)
	bz, err := store.Get(types.VoteKey(proposalID, voterAddr))
	if err != nil {
		return vote, err
	}

	if bz == nil {
		return vote, types.ErrVoteNotFound
	}

	err = keeper.cdc.Unmarshal(bz, &vote)
	if err != nil {
		return vote, err
	}

	return vote, nil
}

// IterateAllVotes iterates over all the stored votes and performs a callback function
func (keeper Keeper) IterateAllVotes(ctx context.Context, cb func(vote v1.Vote) error) error {
	store := keeper.storeService.OpenKVStore(ctx)
	iterator := storetypes.KVStorePrefixIterator(runtime.KVStoreAdapter(store), types.VotesKeyPrefix)

	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		var vote v1.Vote
		err := keeper.cdc.Unmarshal(iterator.Value(), &vote)
		if err != nil {
			return err
		}

		err = cb(vote)
		// exit early without error if cb returns ErrStopIterating
		if errors.IsOf(err, errors.ErrStopIterating) {
			return nil
		} else if err != nil {
			return err
		}
	}

	return nil
}

// IterateVotes iterates over all the proposals votes and performs a callback function
func (keeper Keeper) IterateVotes(ctx context.Context, proposalID uint64, cb func(vote v1.Vote) error) error {
	store := keeper.storeService.OpenKVStore(ctx)
	iterator := storetypes.KVStorePrefixIterator(runtime.KVStoreAdapter(store), types.VotesKey(proposalID))

	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		var vote v1.Vote
		err := keeper.cdc.Unmarshal(iterator.Value(), &vote)
		if err != nil {
			return err
		}

		err = cb(vote)
		// exit early without error if cb returns ErrStopIterating
		if errors.IsOf(err, errors.ErrStopIterating) {
			return nil
		} else if err != nil {
			return err
		}
	}

	return nil
}

// deleteVotes deletes the all votes from a given proposalID.
func (keeper Keeper) deleteVotes(ctx context.Context, proposalID uint64) error {
	store := keeper.storeService.OpenKVStore(ctx)
	return store.Delete(types.VotesKey(proposalID))
}

// deleteVote deletes a vote from a given proposalID and voter from the store
func (keeper Keeper) deleteVote(ctx context.Context, proposalID uint64, voterAddr sdk.AccAddress) error {
	store := keeper.storeService.OpenKVStore(ctx)
	return store.Delete(types.VoteKey(proposalID, voterAddr))
}
