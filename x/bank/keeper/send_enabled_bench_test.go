package keeper_test

import (
	"testing"

	"go.uber.org/mock/gomock"

	"cosmossdk.io/log/v2"

	"github.com/cosmos/cosmos-sdk/codec/address"
	"github.com/cosmos/cosmos-sdk/runtime"
	storetypes "github.com/cosmos/cosmos-sdk/store/v2/types"
	"github.com/cosmos/cosmos-sdk/testutil"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/bank/keeper"
	banktestutil "github.com/cosmos/cosmos-sdk/x/bank/testutil"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
)

// BenchmarkIsSendEnabledDenom measures IsSendEnabledDenom, which runs on the
// SendCoins hot path for every denom in a multi-coin send. It previously paid
// both a Has and a Get for a single logical read; see #26448.
func BenchmarkIsSendEnabledDenom(b *testing.B) {
	key := storetypes.NewKVStoreKey(banktypes.StoreKey)
	oKey := storetypes.NewObjectStoreKey(banktypes.ObjectStoreKey)
	testCtx := testutil.DefaultContextWithObjectStore(b, key, storetypes.NewTransientStoreKey("transient_test"), oKey)
	ctx := testCtx.Ctx
	encCfg := moduletestutil.MakeTestEncodingConfig()

	ctrl := gomock.NewController(b)
	authKeeper := banktestutil.NewMockAccountKeeper(ctrl)
	authKeeper.EXPECT().AddressCodec().Return(address.NewBech32Codec("cosmos")).AnyTimes()

	bankKeeper := keeper.NewBaseKeeper(
		encCfg.Codec,
		runtime.NewKVStoreService(key),
		authKeeper,
		map[string]bool{},
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
		log.NewNopLogger(),
	).WithObjStoreKey(oKey)

	denom := "benchdenom"
	bankKeeper.SetSendEnabled(ctx, denom, true)

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		bankKeeper.IsSendEnabledDenom(ctx, denom)
	}
}
