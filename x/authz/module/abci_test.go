package authz_test

import (
	"testing"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/simapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/authz"
	authzmodule "github.com/cosmos/cosmos-sdk/x/authz/module"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/proto/tendermint/types"
)

func TestExpiredGrantsQueue(t *testing.T) {
	app := simapp.Setup(t, false)
	ctx := app.BaseApp.NewContext(false, types.Header{})
	addrs := simapp.AddTestAddrsIncremental(app, ctx, 4, sdk.NewInt(30000000))
	granter := addrs[0]
	grantee1 := addrs[1]
	grantee2 := addrs[2]
	grantee3 := addrs[3]
	expiration := ctx.BlockTime().AddDate(0, 1, 0)
	smallCoins := sdk.NewCoins(sdk.NewInt64Coin("stake", 10))

	app.AuthzKeeper.SaveGrant(ctx, grantee1, granter, banktypes.NewSendAuthorization(smallCoins), expiration)
	app.AuthzKeeper.SaveGrant(ctx, grantee2, granter, banktypes.NewSendAuthorization(smallCoins), expiration)
	app.AuthzKeeper.SaveGrant(ctx, grantee3, granter, banktypes.NewSendAuthorization(smallCoins), expiration.AddDate(1, 0, 0))

	queryHelper := baseapp.NewQueryServerTestHelper(ctx, app.InterfaceRegistry())
	authz.RegisterQueryServer(queryHelper, app.AuthzKeeper)
	queryClient := authz.NewQueryClient(queryHelper)

	authzmodule.BeginBlocker(ctx, app.AuthzKeeper)

	res, err := queryClient.GranterGrants(ctx.Context(), &authz.QueryGranterGrantsRequest{
		Granter: granter.String(),
	})
	require.NoError(t, err)
	require.NotNil(t, res)
	require.Len(t, res.Grants, 3)

	ctx = ctx.WithBlockTime(expiration.AddDate(0, 2, 0))
	authzmodule.BeginBlocker(ctx, app.AuthzKeeper)
	res, err = queryClient.GranterGrants(ctx.Context(), &authz.QueryGranterGrantsRequest{
		Granter: granter.String(),
	})
	require.NoError(t, err)
	require.NotNil(t, res)
	require.Len(t, res.Grants, 1)
}
