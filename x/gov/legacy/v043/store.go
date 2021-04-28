package v043

import (
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/address"
	v040gov "github.com/cosmos/cosmos-sdk/x/gov/legacy/v040"
	"github.com/cosmos/cosmos-sdk/x/gov/types"
)

const proposalIDLen = 8

// migratePrefixProposalAddress is a helper function that migrates all keys of format:
// <prefix_bytes><proposal_id (8 bytes)><address_bytes>
// into format:
// <prefix_bytes><proposal_id (8 bytes)><address_len (1 byte)><address_bytes>
func migratePrefixProposalAddress(store sdk.KVStore, prefixBz []byte) {
	oldStore := prefix.NewStore(store, prefixBz)

	oldStoreIter := oldStore.Iterator(nil, nil)
	defer oldStoreIter.Close()

	for ; oldStoreIter.Valid(); oldStoreIter.Next() {
		proposalID := oldStoreIter.Key()[:proposalIDLen]
		addr := oldStoreIter.Key()[proposalIDLen:]
		newStoreKey := append(append(prefixBz, proposalID...), address.MustLengthPrefix(addr)...)

		// Set new key on store. Values don't change.
		store.Set(newStoreKey, oldStoreIter.Value())
		oldStore.Delete(oldStoreIter.Key())
	}
}

// migrateStoreWeightedVotes migrates a legacy vote to an ADR-037 weighted vote.
func migrateVote(oldVote v040gov.Vote) types.Vote {
	return types.Vote{
		ProposalId: oldVote.ProposalId,
		Voter:      oldVote.Voter,
		Options:    []types.WeightedVoteOption{{Option: oldVote.Option, Weight: sdk.NewDec(1)}},
	}
}

// migrateStoreWeightedVotes migrates in-place all legacy votes to ADR-037 weighted votes.
func migrateStoreWeightedVotes(store sdk.KVStore, cdc codec.BinaryCodec) error {
	iterator := sdk.KVStorePrefixIterator(store, v040gov.VotesKeyPrefix)

	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		var oldVote v040gov.Vote
		err := cdc.UnmarshalBinaryBare(iterator.Value(), &oldVote)
		if err != nil {
			return err
		}

		newVote := migrateVote(oldVote)
		bz, err := cdc.MarshalBinaryBare(&newVote)
		if err != nil {
			return err
		}

		store.Set(iterator.Key(), bz)
	}

	return nil
}

// MigrateStore performs in-place store migrations from v0.40 to v0.42. The
// migration includes:
//
// - Change addresses to be length-prefixed.
func MigrateStore(ctx sdk.Context, storeKey sdk.StoreKey, cdc codec.BinaryCodec) error {
	store := ctx.KVStore(storeKey)
	migratePrefixProposalAddress(store, v040gov.DepositsKeyPrefix)
	migratePrefixProposalAddress(store, v040gov.VotesKeyPrefix)
	return migrateStoreWeightedVotes(store, cdc)
}
