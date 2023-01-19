package collections

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestGenericUniqueIndex(t *testing.T) {
	// we simulate we have a balance
	// which has multiple coins.
	// we want to maintain a relationship
	// between the owner of the balance
	// and the denoms of the balance.
	// the secondary key is a string (denom), and is the one referencing.
	// the primary key (the key being referenced) is a string ID (eg: stringified acc address).
	// the difference between this and the multi index is the fact that
	// the balance is exclusive, meaning that only one person can own a specific denom.
	sk, ctx := deps()
	sb := NewSchemaBuilder(sk)
	mi := NewGenericUniqueIndex(
		sb, NewPrefix("denoms"), "denom_to_owner", StringKey, StringKey,
		func(pk string, value balance) ([]IndexReference[string, string], error) {
			// the references are all the denoms
			refs := make([]IndexReference[string, string], len(value.coins))
			// this is saying, create a relationship between all the denoms
			// and the owner of the balance.
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

	// assert relations were created
	iter, err := mi.Iterate(ctx, nil)
	require.NoError(t, err)
	defer iter.Close()

	kv, err := iter.KeyValues()
	require.NoError(t, err)
	require.Len(t, kv, 2)
	require.Equal(t, kv[0].Key, "atom")
	require.Equal(t, kv[0].Value, "cosmosAddr1")
	require.Equal(t, kv[1].Key, "osmo")
	require.Equal(t, kv[1].Value, "cosmosAddr1")

	// assert only one address can own a denom (uniqueness)
	err = mi.Reference(ctx, "cosmosAddr2", balance{coins: []coin{{"atom", 5000}}}, nil)
	require.ErrorIs(t, err, ErrConflict)

	// during modifications references are updated
	err = mi.Reference(ctx, "cosmosAddr1",
		balance{coins: []coin{{"atom", 5000}}}, // new
		&balance{coins: []coin{
			{"atom", 1000}, {"osmo", 5000}, // old
		}})

	// reference to osmo must not exist
	_, err = mi.Get(ctx, "osmo")
	require.ErrorIs(t, err, ErrNotFound)

	// unreferencing clears all
	err = mi.Unreference(ctx, "cosmosAddr1", balance{coins: []coin{{"atom", 5000}}})
	require.NoError(t, err)
	_, err = mi.Get(ctx, "atom")
	require.ErrorIs(t, err, ErrNotFound)
}
