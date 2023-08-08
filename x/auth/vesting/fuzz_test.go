package vesting

import (
	"encoding/json"
	"testing"
	"time"

	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	cmttime "github.com/cometbft/cometbft/types/time"
	"github.com/golang/mock/gomock"
	fuzz "github.com/google/gofuzz"

	"cosmossdk.io/log"
	storetypes "cosmossdk.io/store/types"

	"github.com/cosmos/cosmos-sdk/codec/address"
	"github.com/cosmos/cosmos-sdk/runtime"
	"github.com/cosmos/cosmos-sdk/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	vestingtypes "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"
	"github.com/cosmos/cosmos-sdk/x/bank/keeper"
	banktestutil "github.com/cosmos/cosmos-sdk/x/bank/testutil"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
)

var (
	fromAddr = sdk.AccAddress([]byte("from1________________"))
	to2Addr  = sdk.AccAddress([]byte("to2__________________"))
	to3Addr  = sdk.AccAddress([]byte("to3__________________"))
	fooCoin  = sdk.NewInt64Coin("foo", 100)
	accAddrs = []sdk.AccAddress{
		sdk.AccAddress([]byte("addr1_______________")),
		sdk.AccAddress([]byte("addr2_______________")),
		sdk.AccAddress([]byte("addr3_______________")),
		sdk.AccAddress([]byte("addr4_______________")),
		sdk.AccAddress([]byte("addr5_______________")),
	}
)

func FuzzMsgServerCreateVestingAccount(f *testing.F) {
	if testing.Short() {
		f.Skip("Skipping in -short mode")
	}

	// 1. Add some seeds.
	seeds := []*vestingtypes.MsgCreateVestingAccount{
		vestingtypes.NewMsgCreateVestingAccount(
			fromAddr,
			to2Addr,
			sdk.Coins{fooCoin},
			time.Now().Unix(),
			true,
		),
		vestingtypes.NewMsgCreateVestingAccount(
			fromAddr,
			to3Addr,
			sdk.Coins{fooCoin},
			time.Now().Unix(),
			false,
		),
	}

	gf := fuzz.New()
	for _, seed := range seeds {
		for i := 0; i <= 1e4; i++ {
			blob, err := json.Marshal(seed)
			if err != nil {
				f.Fatal(err)
			}
			f.Add(blob)

			// 1.5. Now mutate that seed a couple of times for the next round.
			gf.Fuzz(seed)
		}
	}

	key := storetypes.NewKVStoreKey(authtypes.StoreKey)
	storeService := runtime.NewKVStoreService(key)

	maccPerms := map[string][]string{}

	encCfg := moduletestutil.MakeTestEncodingConfig()
	accountKeeper := authkeeper.NewAccountKeeper(
		encCfg.Codec,
		storeService,
		authtypes.ProtoBaseAccount,
		maccPerms,
		address.NewBech32Codec("cosmos"),
		"cosmos",
		authtypes.NewModuleAddress("gov").String(),
	)

	vestingtypes.RegisterInterfaces(encCfg.InterfaceRegistry)
	authtypes.RegisterInterfaces(encCfg.InterfaceRegistry)

	// 2. Now run the fuzzers.
	f.Fuzz(func(t *testing.T, in []byte) {
		va := new(vestingtypes.MsgCreateVestingAccount)
		if err := json.Unmarshal(in, va); err != nil {
			// Skip over malformed inputs that can JSON unmarshal.
			return
		}

		ctrl := gomock.NewController(t)
		authKeeper := banktestutil.NewMockAccountKeeper(ctrl)
		authKeeper.EXPECT().AddressCodec().Return(address.NewBech32Codec("cosmos")).AnyTimes()
		bankKeeper := keeper.NewBaseKeeper(
			encCfg.Codec,
			storeService,
			authKeeper,
			map[string]bool{accAddrs[4].String(): true},
			authtypes.NewModuleAddress(banktypes.GovModuleName).String(),
			log.NewNopLogger(),
		)

		msgServer := NewMsgServerImpl(accountKeeper, bankKeeper)
		testCtx := testutil.DefaultContextWithDB(t, key, storetypes.NewTransientStoreKey("transient_test"))
		ctx := testCtx.Ctx.WithBlockHeader(cmtproto.Header{Time: cmttime.Now()})

		_, _ = msgServer.CreateVestingAccount(ctx, va)
	})
}
