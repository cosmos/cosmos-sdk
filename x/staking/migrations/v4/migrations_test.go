package v4_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	storetypes "cosmossdk.io/store/types"

	"github.com/cosmos/cosmos-sdk/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"github.com/cosmos/cosmos-sdk/x/staking"
	v4 "github.com/cosmos/cosmos-sdk/x/staking/migrations/v4"
	"github.com/cosmos/cosmos-sdk/x/staking/types"
)

type mockSubspace struct {
	ps types.Params
}

func newMockSubspace(ps types.Params) mockSubspace {
	return mockSubspace{ps: ps}
}

func (ms mockSubspace) GetParamSet(ctx sdk.Context, ps paramtypes.ParamSet) {
	*ps.(*types.Params) = ms.ps
}

// no-op required for type coercion
func (ms mockSubspace) Set(ctx sdk.Context, key []byte, value interface{}) {
}

func TestMigrate(t *testing.T) {
	cdc := moduletestutil.MakeTestEncodingConfig(staking.AppModuleBasic{}).Codec

	storeKey := storetypes.NewKVStoreKey(v4.ModuleName)
	tKey := storetypes.NewTransientStoreKey("transient_test")
	ctx := testutil.DefaultContext(storeKey, tKey)
	store := ctx.KVStore(storeKey)

	legacySubspace := newMockSubspace(types.DefaultParams())

	testCases := []struct {
		name        string
		doMigration bool
	}{
		{
			name:        "without state migration",
			doMigration: false,
		},
		{
			name:        "with state migration",
			doMigration: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.doMigration {
				require.NoError(t, v4.MigrateStore(ctx, store, cdc, legacySubspace))
			}

			if tc.doMigration {
				var res types.Params
				bz := store.Get(v4.ParamsKey)
				require.NoError(t, cdc.Unmarshal(bz, &res))
				require.Equal(t, legacySubspace.ps, res)
				require.Equal(t, types.DefaultMinCommissionRate, legacySubspace.ps.MinCommissionRate)
			} else {
				require.Equal(t, true, true)
			}
		})
	}
}
