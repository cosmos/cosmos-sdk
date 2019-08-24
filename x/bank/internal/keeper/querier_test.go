package keeper_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	abci "github.com/tendermint/tendermint/abci/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	keep "github.com/cosmos/cosmos-sdk/x/bank/internal/keeper"
	"github.com/cosmos/cosmos-sdk/x/bank/internal/types"
)

func TestBalances(t *testing.T) {
	app, ctx := createTestApp(false)
	req := abci.RequestQuery{
		Path: fmt.Sprintf("custom/bank/%s", keep.QueryBalance),
		Data: []byte{},
	}

	querier := keep.NewQuerier(app.BankKeeper)

	res, err := querier(ctx, []string{"balances"}, req)
	require.NotNil(t, err)
	require.Nil(t, res)

	_, _, addr := authtypes.KeyTestPubAddr()
	req.Data = app.Codec().MustMarshalJSON(types.NewQueryBalanceParams(addr))
	res, err = querier(ctx, []string{"balances"}, req)
	require.Nil(t, err) // the account does not exist, no error returned anyway
	require.NotNil(t, res)

	var coins sdk.Coins
	require.NoError(t, app.Codec().UnmarshalJSON(res, &coins))
	require.True(t, coins.IsZero())

	acc := app.AccountKeeper.NewAccountWithAddress(ctx, addr)
	acc.SetCoins(sdk.NewCoins(sdk.NewInt64Coin("foo", 10)))
	app.AccountKeeper.SetAccount(ctx, acc)
	res, err = querier(ctx, []string{"balances"}, req)
	require.Nil(t, err)
	require.NotNil(t, res)
	require.NoError(t, app.Codec().UnmarshalJSON(res, &coins))
	require.True(t, coins.AmountOf("foo").Equal(sdk.NewInt(10)))
}

func TestQuerierRouteNotFound(t *testing.T) {
	app, ctx := createTestApp(false)
	req := abci.RequestQuery{
		Path: "custom/bank/notfound",
		Data: []byte{},
	}

	querier := keep.NewQuerier(app.BankKeeper)
	_, err := querier(ctx, []string{"notfound"}, req)
	require.Error(t, err)
}
