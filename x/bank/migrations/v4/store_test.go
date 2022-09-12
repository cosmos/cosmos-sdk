package v4_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/store/prefix"
	"github.com/cosmos/cosmos-sdk/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
	"github.com/cosmos/cosmos-sdk/x/bank"
	"github.com/cosmos/cosmos-sdk/x/bank/exported"
	v4 "github.com/cosmos/cosmos-sdk/x/bank/migrations/v4"
	"github.com/cosmos/cosmos-sdk/x/bank/types"
)

type mockSubspace struct {
	ps types.Params
}

func newMockSubspace(ps types.Params) mockSubspace {
	return mockSubspace{ps: ps}
}

func (ms mockSubspace) GetParamSet(ctx sdk.Context, ps exported.ParamSet) {
	*ps.(*types.Params) = ms.ps
}

func TestMigrateParams(t *testing.T) {
	encCfg := moduletestutil.MakeTestEncodingConfig(bank.AppModuleBasic{})
	cdc := encCfg.Codec

	storeKey := sdk.NewKVStoreKey(v4.ModuleName)
	tKey := sdk.NewTransientStoreKey("transient_test")
	ctx := testutil.DefaultContext(storeKey, tKey)
	store := ctx.KVStore(storeKey)

	legacySubspace := newMockSubspace(types.DefaultParams())
	require.NoError(t, v4.MigrateStore(ctx, storeKey, legacySubspace, cdc))

	var res types.Params
	bz := store.Get(v4.ParamsKey)
	require.NoError(t, cdc.Unmarshal(bz, &res))
	require.Equal(t, legacySubspace.ps, res)
}

func TestVerifyDenom(t *testing.T) {
	encCfg := moduletestutil.MakeTestEncodingConfig(bank.AppModuleBasic{})
	cdc := encCfg.Codec

	storeKey := sdk.NewKVStoreKey(v4.ModuleName)
	tKey := sdk.NewTransientStoreKey("transient_test")
	ctx := testutil.DefaultContext(storeKey, tKey)
	store := ctx.KVStore(storeKey)
	denomStore := prefix.NewStore(store, v4.DenomMetadataPrefix)

	testcases := []struct {
		denom  string
		expErr bool
	}{
		{"1token", true},
		{"token1", false},
	}

	for _, tc := range testcases {
		t.Run(tc.denom, func(t *testing.T) {
			metadata := types.Metadata{Base: tc.denom, Name: tc.denom, Symbol: tc.denom, Display: tc.denom, DenomUnits: []*types.DenomUnit{{Denom: tc.denom}}}
			denomStore.Set([]byte(tc.denom), cdc.MustMarshal(&metadata))

			err := v4.MigrateStore(ctx, storeKey, newMockSubspace(types.DefaultParams()), cdc)
			if tc.expErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}

			// Cleanup
			denomStore.Delete([]byte(tc.denom))
		})
	}
}
