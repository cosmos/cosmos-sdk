package keeper

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	abci "github.com/tendermint/tendermint/abci/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/bank/internal/types"
)

func TestBalances(t *testing.T) {
	input := setupTestInput()
	req := abci.RequestQuery{
		Path: fmt.Sprintf("custom/bank/%s", QueryBalance),
		Data: []byte{},
	}

	querier := NewQuerier(input.k)

	res, err := querier(input.ctx, []string{"balances"}, req)
	require.NotNil(t, err)
	require.Nil(t, res)

	_, _, addr := authtypes.KeyTestPubAddr()
	req.Data = input.cdc.MustMarshalJSON(types.NewQueryBalanceParams(addr))
	res, err = querier(input.ctx, []string{"balances"}, req)
	require.Nil(t, err) // the account does not exist, no error returned anyway
	require.NotNil(t, res)

	var coins sdk.Coins
	require.NoError(t, input.cdc.UnmarshalJSON(res, &coins))
	require.True(t, coins.IsZero())

	acc := input.ak.NewAccountWithAddress(input.ctx, addr)
	acc.SetCoins(sdk.NewCoins(sdk.NewInt64Coin("foo", 10)))
	input.ak.SetAccount(input.ctx, acc)
	res, err = querier(input.ctx, []string{"balances"}, req)
	require.Nil(t, err)
	require.NotNil(t, res)
	require.NoError(t, input.cdc.UnmarshalJSON(res, &coins))
	require.True(t, coins.AmountOf("foo").Equal(sdk.NewInt(10)))
}

func TestQuerierRouteNotFound(t *testing.T) {
	input := setupTestInput()
	req := abci.RequestQuery{
		Path: "custom/bank/notfound",
		Data: []byte{},
	}

	querier := NewQuerier(input.k)
	_, err := querier(input.ctx, []string{"notfound"}, req)
	require.Error(t, err)
}
