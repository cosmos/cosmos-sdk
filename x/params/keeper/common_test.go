package keeper_test

import (
	"github.com/pointnetwork/cosmos-point-sdk/codec"
	"github.com/pointnetwork/cosmos-point-sdk/simapp"
	storetypes "github.com/pointnetwork/cosmos-point-sdk/store/types"
	"github.com/pointnetwork/cosmos-point-sdk/testutil"
	sdk "github.com/pointnetwork/cosmos-point-sdk/types"
	paramskeeper "github.com/pointnetwork/cosmos-point-sdk/x/params/keeper"
)

func testComponents() (*codec.LegacyAmino, sdk.Context, storetypes.StoreKey, storetypes.StoreKey, paramskeeper.Keeper) {
	marshaler := simapp.MakeTestEncodingConfig().Codec
	legacyAmino := createTestCodec()
	mkey := sdk.NewKVStoreKey("test")
	tkey := sdk.NewTransientStoreKey("transient_test")
	ctx := testutil.DefaultContext(mkey, tkey)
	keeper := paramskeeper.NewKeeper(marshaler, legacyAmino, mkey, tkey)

	return legacyAmino, ctx, mkey, tkey, keeper
}

type invalid struct{}

type s struct {
	I int
}

func createTestCodec() *codec.LegacyAmino {
	cdc := codec.NewLegacyAmino()
	sdk.RegisterLegacyAminoCodec(cdc)
	cdc.RegisterConcrete(s{}, "test/s", nil)
	cdc.RegisterConcrete(invalid{}, "test/invalid", nil)
	return cdc
}
