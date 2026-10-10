package v8

import (
	"testing"

	"github.com/stretchr/testify/require"

	"cosmossdk.io/collections"
	"cosmossdk.io/collections/colltest"

	v7 "github.com/cosmos/cosmos-sdk/x/auth/migrations/v7"
	"github.com/cosmos/cosmos-sdk/x/auth/types"
)

func TestMigrate7to8(t *testing.T) {
	kv, ctx := colltest.MockStore()
	sb := collections.NewSchemaBuilder(kv)
	params := collections.NewItem(sb, collections.NewPrefix(0), "params", colltest.MockValueCodec[types.Params]())

	// No params stored yet.
	require.ErrorIs(t, Migrate(ctx, params), collections.ErrNotFound)

	// v7 params: non-default values for the existing fields and the new fields
	// unset. The enabled flag is set to true to check that the migration
	// forces it off.
	v7Params := types.Params{
		MaxMemoCharacters:      512,
		TxSigLimit:             9,
		TxSizeCostPerByte:      11,
		SigVerifyCostED25519:   600,
		SigVerifyCostSecp256k1: 1100,
		SigVerifyCostMlDsa65:   800,
		PubKeyChangeEnabled:    true,
		PubKeyChangeCost:       0,
	}
	require.NoError(t, params.Set(ctx, v7Params))

	require.NoError(t, Migrate(ctx, params))

	got, err := params.Get(ctx)
	require.NoError(t, err)
	require.False(t, got.PubKeyChangeEnabled)
	require.Equal(t, types.DefaultPubKeyChangeCost, got.PubKeyChangeCost)

	// The other fields are unchanged.
	want := v7Params
	want.PubKeyChangeEnabled = false
	want.PubKeyChangeCost = types.DefaultPubKeyChangeCost
	require.Equal(t, want, got)
}

// TestMigrate6to8 runs the 6->7 and 7->8 migrations back to back on params as
// stored by a v6 chain, as RunMigrations does for a chain skipping v7.
func TestMigrate6to8(t *testing.T) {
	kv, ctx := colltest.MockStore()
	sb := collections.NewSchemaBuilder(kv)
	params := collections.NewItem(sb, collections.NewPrefix(0), "params", colltest.MockValueCodec[types.Params]())

	// v6 params: only fields 1-5 exist.
	v6Params := types.Params{
		MaxMemoCharacters:      256,
		TxSigLimit:             7,
		TxSizeCostPerByte:      10,
		SigVerifyCostED25519:   590,
		SigVerifyCostSecp256k1: 1000,
	}
	require.NoError(t, params.Set(ctx, v6Params))

	require.NoError(t, v7.Migrate(ctx, params))
	require.NoError(t, Migrate(ctx, params))

	got, err := params.Get(ctx)
	require.NoError(t, err)

	want := v6Params
	want.SigVerifyCostMlDsa65 = types.DefaultSigVerifyCostMlDsa65
	want.PubKeyChangeEnabled = types.DefaultPubKeyChangeEnabled
	want.PubKeyChangeCost = types.DefaultPubKeyChangeCost
	require.Equal(t, want, got)
	require.NoError(t, got.Validate())
}
