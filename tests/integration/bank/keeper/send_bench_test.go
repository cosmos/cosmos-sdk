package keeper_test

import (
	"context"
	"testing"

	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"

	"cosmossdk.io/log/v2"

	addresscodec "github.com/cosmos/cosmos-sdk/codec/address"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	"github.com/cosmos/cosmos-sdk/runtime"
	storetypes "github.com/cosmos/cosmos-sdk/store/v2/types"
	"github.com/cosmos/cosmos-sdk/testutil/integration"
	sdk "github.com/cosmos/cosmos-sdk/types"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
	"github.com/cosmos/cosmos-sdk/x/auth"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	authsims "github.com/cosmos/cosmos-sdk/x/auth/simulation"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/bank"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	banktestutil "github.com/cosmos/cosmos-sdk/x/bank/testutil"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
)

// countingAccountKeeper wraps a real, store-backed AccountKeeper and counts how
// often HasAccount is invoked while still delegating to the genuine lookup
// (Accounts.Has -> KV store read). This lets the benchmark measure the real
// cost of the avoided probe rather than the near-zero cost of a gomock stub,
// which is what the previous mock-based benchmark could not demonstrate.
type countingAccountKeeper struct {
	authkeeper.AccountKeeper
	hasAccountCalls *int64
}

func (c countingAccountKeeper) HasAccount(ctx context.Context, addr sdk.AccAddress) bool {
	*c.hasAccountCalls++
	return c.AccountKeeper.HasAccount(ctx, addr)
}

// BenchmarkSendCoins_ExistingRecipient exercises the hot path the #24228
// optimization targets: the recipient already holds a non-zero balance in the
// transferred denom, so addCoins reports hadPriorBalance=true and SendCoins can
// skip the redundant auth HasAccount probe (an account with a balance provably
// already exists).
//
// Unlike the prior x/bank/keeper benchmark, this wires REAL auth + bank keepers
// over a real multistore, so HasAccount performs an actual store lookup. The
// reported hasaccount/op metric makes the avoided work unambiguous: on the
// parent commit it is ~1.00 (the probe fires every call), on this branch it is
// 0.00.
//
// Compare with benchstat:
//
//	go test -run NONE -bench BenchmarkSendCoins_ExistingRecipient \
//	    -benchmem -count=10 ./tests/integration/bank/keeper/ > new.txt
//	git checkout origin/main -- x/bank/keeper/send.go x/bank/keeper/keeper.go x/bank/keeper/virtual.go
//	go test -run NONE -bench BenchmarkSendCoins_ExistingRecipient \
//	    -benchmem -count=10 ./tests/integration/bank/keeper/ > old.txt
//	git checkout HEAD -- x/bank/keeper/send.go x/bank/keeper/keeper.go x/bank/keeper/virtual.go
//	benchstat old.txt new.txt
func BenchmarkSendCoins_ExistingRecipient(b *testing.B) {
	keys := storetypes.NewKVStoreKeys(authtypes.StoreKey, banktypes.StoreKey)
	cdc := moduletestutil.MakeTestEncodingConfig(auth.AppModuleBasic{}, bank.AppModuleBasic{}).Codec
	logger := log.NewNopLogger()

	cms := integration.CreateMultiStore(keys, logger)
	ctx := sdk.NewContext(cms, cmtproto.Header{Height: 1, AppHash: []byte("bench")}, false, logger)

	authority := authtypes.NewModuleAddress("gov")
	maccPerms := map[string][]string{minttypes.ModuleName: {authtypes.Minter}}

	accountKeeper := authkeeper.NewAccountKeeper(
		cdc,
		runtime.NewKVStoreService(keys[authtypes.StoreKey]),
		authtypes.ProtoBaseAccount,
		maccPerms,
		addresscodec.NewBech32Codec(sdk.Bech32MainPrefix),
		sdk.Bech32MainPrefix,
		authority.String(),
	)

	var hasAccountCalls int64
	ak := countingAccountKeeper{AccountKeeper: accountKeeper, hasAccountCalls: &hasAccountCalls}

	bankKeeper := bankkeeper.NewBaseKeeper(
		cdc,
		runtime.NewKVStoreService(keys[banktypes.StoreKey]),
		ak,
		map[string]bool{},
		authority.String(),
		log.NewNopLogger(),
	)

	authModule := auth.NewAppModule(cdc, accountKeeper, authsims.RandomGenesisAccounts)
	bankModule := bank.NewAppModule(cdc, bankKeeper, accountKeeper)
	authModule.InitGenesis(ctx, cdc, authModule.DefaultGenesis(cdc))
	bankModule.InitGenesis(ctx, cdc, bankModule.DefaultGenesis(cdc))

	const denom = "stake"
	sender := sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())
	recipient := sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())

	if err := banktestutil.FundAccount(ctx, bankKeeper, sender, sdk.NewCoins(sdk.NewInt64Coin(denom, 1_000_000_000_000))); err != nil {
		b.Fatalf("fund sender: %v", err)
	}
	// Seed the recipient with a non-zero balance so the account already exists;
	// every subsequent send hits the optimized path.
	if err := banktestutil.FundAccount(ctx, bankKeeper, recipient, sdk.NewCoins(sdk.NewInt64Coin(denom, 1))); err != nil {
		b.Fatalf("fund recipient: %v", err)
	}

	amt := sdk.NewCoins(sdk.NewInt64Coin(denom, 1))

	hasAccountCalls = 0
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		if err := bankKeeper.SendCoins(ctx, sender, recipient, amt); err != nil {
			b.Fatalf("send: %v", err)
		}
	}
	b.StopTimer()

	b.ReportMetric(float64(hasAccountCalls)/float64(b.N), "hasaccount/op")
}
