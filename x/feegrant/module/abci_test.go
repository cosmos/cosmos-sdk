package module_test

import (
	"testing"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/simapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/feegrant"
	"github.com/cosmos/cosmos-sdk/x/feegrant/module"
	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
)

func TestFeegrantPruning(t *testing.T) {
	app := simapp.Setup(t, false)

	ctx := app.BaseApp.NewContext(false, tmproto.Header{})
	addrs := simapp.AddTestAddrs(app, ctx, 4, sdk.NewInt(1000))
	granter1 := addrs[0]
	granter2 := addrs[1]
	granter3 := addrs[2]
	grantee := addrs[3]
	spendLimit := sdk.NewCoins(sdk.NewCoin("stake", sdk.NewInt(1000)))
	now := ctx.BlockTime()
	oneDay := now.AddDate(0, 0, 1)

	header := tmproto.Header{Height: app.LastBlockHeight() + 1}
	app.BeginBlock(abci.RequestBeginBlock{Header: header})

	app.FeeGrantKeeper.GrantAllowance(
		ctx,
		granter1,
		grantee,
		&feegrant.BasicAllowance{
			Expiration: &now,
		},
	)
	app.FeeGrantKeeper.GrantAllowance(
		ctx,
		granter2,
		grantee,
		&feegrant.BasicAllowance{
			SpendLimit: spendLimit,
		},
	)
	app.FeeGrantKeeper.GrantAllowance(
		ctx,
		granter3,
		grantee,
		&feegrant.BasicAllowance{
			Expiration: &oneDay,
		},
	)

	queryHelper := baseapp.NewQueryServerTestHelper(ctx, app.InterfaceRegistry())
	feegrant.RegisterQueryServer(queryHelper, app.FeeGrantKeeper)
	queryClient := feegrant.NewQueryClient(queryHelper)

	module.EndBlocker(ctx, app.FeeGrantKeeper)

	res, err := queryClient.Allowances(ctx.Context(), &feegrant.QueryAllowancesRequest{
		Grantee: grantee.String(),
	})
	require.NoError(t, err)
	require.NotNil(t, res)
	require.Len(t, res.Allowances, 3)

	ctx = ctx.WithBlockTime(now.AddDate(0, 0, 2))
	module.EndBlocker(ctx, app.FeeGrantKeeper)

	res, err = queryClient.Allowances(ctx.Context(), &feegrant.QueryAllowancesRequest{
		Grantee: grantee.String(),
	})
	require.NoError(t, err)
	require.NotNil(t, res)
	require.Len(t, res.Allowances, 1)
}
