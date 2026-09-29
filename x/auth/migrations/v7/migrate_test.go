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

	// Now set default params expected before migration
	paramsUnderTest := types.DefaultParams()
	paramsUnderTest.SigVerifyCostMlDsa65 = 0
	err := params.Set(ctx, paramsUnderTest)
	require.NoError(t, err)

	err = Migrate(ctx, params)
	require.NoError(t, err)

	// check that after migration the params object has the default value for SigVerifyCostMlDsa65
	seenParams, err := params.Get(ctx)
	require.NoError(t, err)
	require.Equal(t, types.DefaultSigVerifyCostMlDsa65, seenParams.SigVerifyCostMlDsa65)
}
