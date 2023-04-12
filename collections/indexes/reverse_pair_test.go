package indexes

import (
	"testing"

	"cosmossdk.io/collections"
	"github.com/stretchr/testify/require"
)

type (
	Address = string
	Denom   = string
	Amount  = uint64
)

// our balance index, allows us to efficiently create an index between the key that maps
// balances which is a collections.Pair[Address, Denom] and the Denom.
type balanceIndex struct {
	Denom *ReversePair[Address, Denom, Amount]
}

func (b balanceIndex) IndexesList() []collections.Index[collections.Pair[Address, Denom], Amount] {
	return []collections.Index[collections.Pair[Address, Denom], Amount]{b.Denom}
}

func TestReversePair(t *testing.T) {
	sk, ctx := deps()
	sb := collections.NewSchemaBuilder(sk)
	// we create an indexed map that maps balances, which are saved as
	// key: Pair[Address, Denom]
	// value: Amount
	keyCodec := collections.PairKeyCodec(collections.StringKey, collections.StringKey)

	indexedMap := collections.NewIndexedMap(
		sb,
		collections.NewPrefix("balances"), "balances",
		keyCodec,
		collections.Uint64Value,
		balanceIndex{
			Denom: NewReversePair[Amount](sb, collections.NewPrefix("denom_index"), "denom_index", keyCodec),
		},
	)

	err := indexedMap.Set(ctx, collections.Join("address1", "atom"), 100)
	require.NoError(t, err)

	err = indexedMap.Set(ctx, collections.Join("address1", "osmo"), 200)
	require.NoError(t, err)

	err = indexedMap.Set(ctx, collections.Join("address2", "osmo"), 300)
	require.NoError(t, err)

	// assert if we iterate over osmo we find address1 and address2
	iter, err := indexedMap.Indexes.Denom.MatchExact(ctx, "osmo")
	require.NoError(t, err)
	defer iter.Close()

	pks, err := iter.PrimaryKeys()
	require.NoError(t, err)
	require.Equal(t, "address1", pks[0].K1())
	require.Equal(t, "address2", pks[1].K1())
}
