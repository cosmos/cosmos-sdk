package keeper_test

import (
	"testing"

	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	cmttime "github.com/cometbft/cometbft/types/time"
	"go.uber.org/mock/gomock"

	"cosmossdk.io/log/v2"

	"github.com/cosmos/cosmos-sdk/codec/address"
	"github.com/cosmos/cosmos-sdk/runtime"
	storetypes "github.com/cosmos/cosmos-sdk/store/v2/types"
	"github.com/cosmos/cosmos-sdk/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/bank/keeper"
	banktestutil "github.com/cosmos/cosmos-sdk/x/bank/testutil"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
)

// BenchmarkSendCoins_ExistingRecipient exercises BaseSendKeeper.SendCoins
// along the hot path that the #24228 optimization targets: the recipient
// already holds a balance in the input denom, so addCoins reports
// hadPriorBalance=true and the auth.HasAccount/SetAccount/
// NewAccountWithAddress probe is skipped.
//
// Intended usage is benchstat-style comparison: run this benchmark on the
// parent commit and on this branch. On the parent, the probe fires once
// per call and the mocked AccountKeeper records the extra gomock-mediated
// calls in allocs/op and ns/op. On this branch the probe is gone and the
// delta surfaces directly.
//
//	go test -run XXX -bench BenchmarkSendCoins_ExistingRecipient \
//	    -benchmem -count=10 ./x/bank/keeper/ > new.txt
//	git stash && git checkout HEAD~ -- x/bank/keeper x/bank/testutil
//	go test -run XXX -bench BenchmarkSendCoins_ExistingRecipient \
//	    -benchmem -count=10 ./x/bank/keeper/ > old.txt
//	benchstat old.txt new.txt
func BenchmarkSendCoins_ExistingRecipient(b *testing.B) {
	b.ReportAllocs()

	key := storetypes.NewKVStoreKey(banktypes.StoreKey)
	oKey := storetypes.NewObjectStoreKey(banktypes.ObjectStoreKey)
	testCtx := testutil.DefaultContextWithObjectStore(b, key, storetypes.NewTransientStoreKey("transient_bench"), oKey)
	ctx := testCtx.Ctx.WithBlockHeader(cmtproto.Header{Time: cmttime.Now()})

	encCfg := moduletestutil.MakeTestEncodingConfig()
	storeService := runtime.NewKVStoreService(key)

	ctrl := gomock.NewController(b)
	defer ctrl.Finish()

	authKeeper := banktestutil.NewMockAccountKeeper(ctrl)
	authKeeper.EXPECT().AddressCodec().Return(address.NewBech32Codec("cosmos")).AnyTimes()
	authKeeper.EXPECT().GetModuleAddress(gomock.Any()).Return(mintAcc.GetAddress()).AnyTimes()
	authKeeper.EXPECT().GetModuleAccount(gomock.Any(), gomock.Any()).Return(mintAcc).AnyTimes()
	authKeeper.EXPECT().GetAccount(gomock.Any(), gomock.Any()).Return(authtypes.NewBaseAccountWithAddress(accAddrs[0])).AnyTimes()
	authKeeper.EXPECT().NewAccountWithAddress(gomock.Any(), gomock.Any()).Return(authtypes.NewBaseAccountWithAddress(accAddrs[1])).AnyTimes()
	authKeeper.EXPECT().SetAccount(gomock.Any(), gomock.Any()).AnyTimes()
	authKeeper.EXPECT().HasAccount(gomock.Any(), gomock.Any()).Return(true).AnyTimes()

	bk := keeper.NewBaseKeeper(
		encCfg.Codec,
		storeService,
		authKeeper,
		nil,
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
		log.NewNopLogger(),
	)
	bk = bk.WithObjStoreKey(oKey)

	sender := accAddrs[0]
	recipient := accAddrs[1]
	denom := fooDenom

	if err := banktestutil.FundAccount(ctx, bk, sender, sdk.NewCoins(sdk.NewInt64Coin(denom, 1_000_000_000_000))); err != nil {
		b.Fatalf("fund sender: %v", err)
	}
	if err := banktestutil.FundAccount(ctx, bk, recipient, sdk.NewCoins(sdk.NewInt64Coin(denom, 1))); err != nil {
		b.Fatalf("fund recipient: %v", err)
	}

	amt := sdk.NewCoins(sdk.NewInt64Coin(denom, 1))

	b.ResetTimer()
	for b.Loop() {
		if err := bk.SendCoins(ctx, sender, recipient, amt); err != nil {
			b.Fatalf("send: %v", err)
		}
	}
}
