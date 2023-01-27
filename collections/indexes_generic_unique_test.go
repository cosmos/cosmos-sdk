package collections

import (
	"testing"

	"github.com/stretchr/testify/require"
)

type nftBalance struct {
	nftIDs []uint64
}

func TestGenericUniqueIndex(t *testing.T) {
	// we create the same testing context as with GenericMultiIndex. We have a mapping:
	// Address => NFT balance.
	// An NFT balance is represented as a slice of IDs, those IDs are unique, meaning that
	// they can be held only by one address.
	sk, ctx := deps()
	sb := NewSchemaBuilder(sk)
	ui := NewGenericUniqueIndex(
		sb, NewPrefix("nft_to_owner_index"), "ntf_to_owner_index", Uint64Key, StringKey,
		func(pk string, value nftBalance) ([]IndexReference[uint64, string], error) {
			// the referencing keys are all the NFT unique ids.
			refs := make([]IndexReference[uint64, string], len(value.nftIDs))
			// for each NFT contained in the balance we create an index reference
			// between the NFT unique ID and the owner of the balance.
			for i, id := range value.nftIDs {
				refs[i] = NewIndexReference(id, pk)
			}
			return refs, nil
		},
	)

	// let's create the relationships
	err := ui.Reference(ctx, "cosmosAddr1", nftBalance{nftIDs: []uint64{0, 1}}, nil)
	require.NoError(t, err)

	// assert relations were created
	iter, err := ui.Iterate(ctx, nil)
	require.NoError(t, err)
	defer iter.Close()

	kv, err := iter.KeyValues()
	require.NoError(t, err)
	require.Len(t, kv, 2)
	require.Equal(t, kv[0].Key, uint64(0))
	require.Equal(t, kv[0].Value, "cosmosAddr1")
	require.Equal(t, kv[1].Key, uint64(1))
	require.Equal(t, kv[1].Value, "cosmosAddr1")

	// assert only one address can own a unique NFT
	err = ui.Reference(ctx, "cosmosAddr2", nftBalance{nftIDs: []uint64{0}}, nil) // nft with ID 0 is already owned by cosmosAddr1
	require.ErrorIs(t, err, ErrConflict)

	// during modifications references are updated, we update the index in
	// such a way that cosmosAddr1 loses ownership of nft with id 0.
	err = ui.Reference(ctx, "cosmosAddr1",
		nftBalance{nftIDs: []uint64{1}},     // this is the update nft balance, which contains only id 1
		&nftBalance{nftIDs: []uint64{0, 1}}, // this is the old nft balance, which contains both 0 and 1
	)
	require.NoError(t, err)

	// the updated balance does not contain nft with id 0
	_, err = ui.Get(ctx, 0)
	require.ErrorIs(t, err, ErrNotFound)

	// unreferencing clears all the indexes
	err = ui.Unreference(ctx, "cosmosAddr1", nftBalance{nftIDs: []uint64{1}})
	require.NoError(t, err)
	_, err = ui.Get(ctx, 1)
	require.ErrorIs(t, err, ErrNotFound)
}
