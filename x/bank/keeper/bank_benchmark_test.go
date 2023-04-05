package keeper_test

import (
	"cosmossdk.io/collections"
	storetypes "cosmossdk.io/store/types"
	"fmt"
	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	cmttime "github.com/cometbft/cometbft/types/time"
	"github.com/cosmos/cosmos-sdk/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/bank/keeper"
	banktestutil "github.com/cosmos/cosmos-sdk/x/bank/testutil"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	"github.com/golang/mock/gomock"
	"testing"
	"time"
)

func BenchmarkSetBalances(b *testing.B) {
	key := storetypes.NewKVStoreKey(banktypes.StoreKey)
	testCtx := testutil.DefaultContextWithDB(b, key, storetypes.NewTransientStoreKey("transient_test"))
	ctx := testCtx.Ctx.WithBlockHeader(cmtproto.Header{Time: cmttime.Now()})
	encCfg := moduletestutil.MakeTestEncodingConfig()

	// gomock initializations
	ctrl := gomock.NewController(b)
	authKeeper := banktestutil.NewMockAccountKeeper(ctrl)

	bankKeeper := keeper.NewBaseKeeper(
		encCfg.Codec,
		key,
		authKeeper,
		nil,
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)

	// 1000 accounts, 10 coins each
	const accountsN = 100000
	const balancesPerAccount = 10
	type balance struct {
		addr sdk.AccAddress
		coin sdk.Coin
	}
	balances := make([]balance, 0, accountsN*balancesPerAccount)
	for i := 0; i < accountsN; i++ {
		addr := sdk.Uint64ToBigEndian(uint64(i))
		for j := 0; j < balancesPerAccount; j++ {
			balances = append(balances, balance{
				addr: addr,
				coin: sdk.NewInt64Coin(fmt.Sprintf("denom%d", i), int64(j)),
			})
		}
	}

	now := time.Now()
	for _, balance := range balances {
		_ = bankKeeper.Balances.Set(ctx, collections.Join(balance.addr, balance.coin.Denom), balance.coin.Amount)
	}
	b.Log("time elapsed ", time.Now().Sub(now))
}
