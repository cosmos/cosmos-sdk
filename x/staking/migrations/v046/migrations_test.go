package v046_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/staking/exported"
	"github.com/cosmos/cosmos-sdk/x/staking/types"
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

//func TestMigrate(t *testing.T) {
//	encCfg := moduletestutil.MakeTestEncodingConfig(distribution.AppModuleBasic{})
//	cdc := encCfg.Codec
//
//	storeKey := sdk.NewKVStoreKey(v046.ModuleName)
//	tKey := sdk.NewTransientStoreKey("transient_test")
//	ctx := testutil.DefaultContext(storeKey, tKey)
//	store := ctx.KVStore(storeKey)
//
//	legacySubspace := newMockSubspace(types.DefaultParams())
//	require.NoError(t, v046.MigrateStore(ctx, storeKey, cdc, legacySubspace))
//
//	var res types.Params
//	bz := store.Get(v046.ParametersKey)
//	require.NoError(t, cdc.Unmarshal(bz, &res))
//	require.Equal(t, legacySubspace.ps, res)
//}
