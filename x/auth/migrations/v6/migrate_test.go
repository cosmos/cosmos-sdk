package v6

import (
	"testing"

	"github.com/stretchr/testify/require"

	"cosmossdk.io/collections"
	"cosmossdk.io/collections/colltest"
)

func TestMigrate(t *testing.T) {
	kv, ctx := colltest.MockStore()
	sb := collections.NewSchemaBuilder(kv)
	seq := collections.NewSequence(sb, collections.NewPrefix(2), "seq")

	wantValue := uint64(100)
	err := seq.Set(ctx, wantValue)
	require.NoError(t, err)

	err = Migrate(ctx, kv, seq)
	require.NoError(t, err)

	// check that after migration the sequence is what we want it to be
	_, err = (collections.Item[uint64])(seq).Get(ctx)
	require.ErrorIs(t, err, collections.ErrNotFound)

	// case the global account number was not set
	ctx = kv.NewStoreContext() // this resets the store to zero
	err = Migrate(ctx, kv, seq)
	require.NoError(t, err)
	_, err = (collections.Item[uint64])(seq).Get(ctx)
	require.ErrorIs(t, err, collections.ErrNotFound)
}
