package v2_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
	v2 "github.com/cosmos/cosmos-sdk/x/crisis/migrations/v2"
	"github.com/cosmos/cosmos-sdk/x/crisis/types"
	"github.com/cosmos/cosmos-sdk/x/distribution"
)

type mockSubspace struct {
	constantFee sdk.Coin
}

func newMockSubspace(fee sdk.Coin) mockSubspace {
	return mockSubspace{constantFee: fee}
}

func (ms mockSubspace) Get(ctx sdk.Context, key []byte, ptr interface{}) {
	*ptr.(*sdk.Coin) = ms.constantFee
}

func TestMigrate(t *testing.T) {
	encCfg := moduletestutil.MakeTestEncodingConfig(distribution.AppModuleBasic{})
	cdc := encCfg.Codec

	storeKey := sdk.NewKVStoreKey(v2.ModuleName)
	tKey := sdk.NewTransientStoreKey("transient_test")
	ctx := testutil.DefaultContext(storeKey, tKey)
	store := ctx.KVStore(storeKey)

	legacySubspace := newMockSubspace(types.DefaultGenesisState().ConstantFee)
	require.NoError(t, v2.MigrateStore(ctx, storeKey, legacySubspace, cdc))

	var res sdk.Coin
	bz := store.Get(v2.ConstantFeeKey)
	require.NoError(t, cdc.Unmarshal(bz, &res))
	require.NotNil(t, res)
	require.Equal(t, legacySubspace.constantFee, res)
}
