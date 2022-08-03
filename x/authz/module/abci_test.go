package authz_test

import (
	"testing"
	"time"

	"github.com/cosmos/cosmos-sdk/baseapp"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/authz"
	"github.com/cosmos/cosmos-sdk/x/authz/keeper"
	authzmodule "github.com/cosmos/cosmos-sdk/x/authz/module"
	"github.com/cosmos/cosmos-sdk/x/authz/testutil"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/proto/tendermint/types"
)

func TestExpiredGrantsQueue(t *testing.T) {
	var interfaceRegistry codectypes.InterfaceRegistry
	var authzKeeper keeper.Keeper
	var bankKeeper bankkeeper.Keeper
	var stakingKeeper *stakingkeeper.Keeper

	app, err := simtestutil.Setup(
		testutil.AppConfig,
		&interfaceRegistry,
		&authzKeeper,
		&bankKeeper,
		&stakingKeeper,
	)
	require.NoError(t, err)

	ctx := app.BaseApp.NewContext(false, types.Header{})
	addrs := simtestutil.AddTestAddrsIncremental(bankKeeper, stakingKeeper, ctx, 5, sdk.NewInt(30000000))
	granter := addrs[0]
	grantee1 := addrs[1]
	grantee2 := addrs[2]
	grantee3 := addrs[3]
	grantee4 := addrs[4]
	expiration := ctx.BlockTime().AddDate(0, 1, 0)
	expiration2 := expiration.AddDate(1, 0, 0)
	smallCoins := sdk.NewCoins(sdk.NewInt64Coin("stake", 10))
	sendAuthz := banktypes.NewSendAuthorization(smallCoins, nil)

	save := func(grantee sdk.AccAddress, exp *time.Time) {
		err := authzKeeper.SaveGrant(ctx, grantee, granter, sendAuthz, exp)
		require.NoError(t, err, "Grant from %s", grantee.String())
	}
	save(grantee1, &expiration)
	save(grantee2, &expiration)
	save(grantee3, &expiration2)
	save(grantee4, nil)

	queryHelper := baseapp.NewQueryServerTestHelper(ctx, interfaceRegistry)
	authz.RegisterQueryServer(queryHelper, authzKeeper)
	queryClient := authz.NewQueryClient(queryHelper)

	checkGrants := func(ctx sdk.Context, expectedNum int) {
		authzmodule.BeginBlocker(ctx, authzKeeper)

		res, err := queryClient.GranterGrants(ctx.Context(), &authz.QueryGranterGrantsRequest{
			Granter: granter.String(),
		})
		require.NoError(t, err)
		require.NotNil(t, res)
		require.Equal(t, expectedNum, len(res.Grants))
	}

	checkGrants(ctx, 4)

	// expiration is exclusive!
	ctx = ctx.WithBlockTime(expiration)
	checkGrants(ctx, 4)

	ctx = ctx.WithBlockTime(expiration.AddDate(0, 0, 1))
	checkGrants(ctx, 2)

	ctx = ctx.WithBlockTime(expiration2.AddDate(0, 0, 1))
	checkGrants(ctx, 1)
}
