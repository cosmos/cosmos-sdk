package testutil_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	coretesting "cosmossdk.io/core/testing"

	"github.com/cosmos/cosmos-sdk/testutil"
)

func TestDiffCollectionsMigration(t *testing.T) {
	ctx := coretesting.Context()
	kvs := coretesting.KVStoreService(ctx, "test")

	// First try with some invalid hash
	err := testutil.DiffCollectionsMigration(
		ctx,
		kvs,
		5,
		func(i int64) {
			if err := kvs.OpenKVStore(ctx).Set([]byte{byte(i)}, []byte{byte(i)}); err != nil {
				panic(err)
			}
		},
		"abcdef0123456789",
	)
	require.Error(t, err)

	// Now reset and try with the correct hash
	err = testutil.DiffCollectionsMigration(
		ctx,
		kvs,
		5,
		func(i int64) {
			if err := kvs.OpenKVStore(ctx).Set([]byte{byte(i)}, []byte{byte(i)}); err != nil {
				panic(err)
			}
		},
		"79541ed9da9c16cb7a1d43d5a3d5f6ee31a873c85a6cb4334fb99e021ee0e556",
	)
	require.NoError(t, err)

	// Change the data a little and it will result in an error
	err = testutil.DiffCollectionsMigration(
		ctx,
		kvs,
		5,
		func(i int64) {
			if err := kvs.OpenKVStore(ctx).Set([]byte{byte(i)}, []byte{byte(i + 1)}); err != nil {
				panic(err)
			}
		},
		"79541ed9da9c16cb7a1d43d5a3d5f6ee31a873c85a6cb4334fb99e021ee0e556",
	)
	require.Error(t, err)
}
