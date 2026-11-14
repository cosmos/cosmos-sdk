package v7

import (
	"testing"

	"github.com/stretchr/testify/require"

	"cosmossdk.io/collections"
	"cosmossdk.io/collections/colltest"

	"github.com/cosmos/cosmos-sdk/x/auth/types"
)

func TestMigrate(t *testing.T) {
	kv, ctx := colltest.MockStore()
	sb := collections.NewSchemaBuilder(kv)
	params := collections.NewItem(sb, collections.NewPrefix(0), "params", colltest.MockValueCodec[types.Params]())

	// First test with invalid params i.e. none
	require.ErrorIs(t, Migrate(ctx, params), collections.ErrNotFound)

	// Params as stored by a v6 chain: only fields 1-5 exist. Fields added in
	// later versions are zero, so do not seed from DefaultParams().
	paramsUnderTest := types.Params{
		MaxMemoCharacters:      types.DefaultMaxMemoCharacters,
		TxSigLimit:             types.DefaultTxSigLimit,
		TxSizeCostPerByte:      types.DefaultTxSizeCostPerByte,
		SigVerifyCostED25519:   types.DefaultSigVerifyCostED25519,
		SigVerifyCostSecp256k1: types.DefaultSigVerifyCostSecp256k1,
	}
	err := params.Set(ctx, paramsUnderTest)
	require.NoError(t, err)

	err = Migrate(ctx, params)
	require.NoError(t, err)

	// check that after migration the params object has the default value for SigVerifyCostMlDsa65
	seenParams, err := params.Get(ctx)
	require.NoError(t, err)
	require.Equal(t, types.DefaultSigVerifyCostMlDsa65, seenParams.SigVerifyCostMlDsa65)

	// The other fields are unchanged.
	want := paramsUnderTest
	want.SigVerifyCostMlDsa65 = types.DefaultSigVerifyCostMlDsa65
	require.Equal(t, want, seenParams)
}
