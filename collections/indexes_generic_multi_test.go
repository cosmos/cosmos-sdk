package collections

import (
	"testing"

	"github.com/stretchr/testify/require"
)

type coin struct {
	denom  string // this will be used as indexing field.
	amount uint64
}

type balance struct {
	coins []coin
}

func TestGenericMultiIndex(t *testing.T) {
	// we are simulating a context in which we have the following mapping:
	//
	// address (represented as string) => balance (slice of coins).
	//
	// we want to create an index that creates a relationship between the coin
	// denom, which is part of the balance structure, and the address. This means
	// we know given a denom who are the addresses holding that denom.
	// From GenericMultiIndex point of view, the denom field of the array becomes
	// the referencing key which points to the address (string), which is the key
	// being referenced.
	sk, ctx := deps()
	sb := NewSchemaBuilder(sk)
	mi := NewGenericMultiIndex(
		sb, NewPrefix("denoms"), "denom_to_owner", StringKey, StringKey,
		func(pk string, value balance) ([]IndexReference[string, string], error) {
			// the referencing keys are all the denoms.
			refs := make([]IndexReference[string, string], len(value.coins))
			// the index reference being created, generates a relationship
			// between denom (the key that references) and pk (address, the key
			// that is being referenced).
			for i, coin := range value.coins {
				refs[i] = NewIndexReference(coin.denom, pk)
			}
			return refs, nil
		},
	)

	// let's create the relationships
	err := mi.Reference(ctx, "cosmosAddr1", balance{coins: []coin{
		{"atom", 1000}, {"osmo", 5000},
	}}, nil)
	require.NoError(t, err)

	// we must find relations between cosmosaddr1 and the denom atom and osmo
	iter, err := mi.Iterate(ctx, nil)
	require.NoError(t, err)

	keys, err := iter.Keys()
	require.NoError(t, err)
	require.Len(t, keys, 2)
	require.Equal(t, keys[0].K1(), "atom") // assert relationship with atom created
	require.Equal(t, keys[1].K1(), "osmo") // assert relationship with osmo created

	// if we update the reference to remove osmo as balance then we must not find it anymore
	err = mi.Reference(ctx, "cosmosAddr1", balance{coins: []coin{{"atom", 1000}}}, // this is the update which does not have osmo
		&balance{coins: []coin{{"atom", 1000}, {"osmo", 5000}}}, // this is the previous record
	)
	require.NoError(t, err)

	exists, err := mi.Has(ctx, "osmo", "cosmosAddr1") // osmo must not exist anymore
	require.NoError(t, err)
	require.False(t, exists)

	exists, err = mi.Has(ctx, "atom", "cosmosAddr1") // atom still exists
	require.NoError(t, err)
	require.True(t, exists)

	// if we unreference then no relationship is maintained anymore
	err = mi.Unreference(ctx, "cosmosAddr1", balance{coins: []coin{{"atom", 1000}}})
	require.NoError(t, err)

	exists, err = mi.Has(ctx, "atom", "cosmosAddr1") // atom is not part of the index anymore because cosmosAddr1 was removed.
	require.NoError(t, err)
	require.False(t, exists)
}
