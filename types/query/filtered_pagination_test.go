package query_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/bank/types"
)

func TestFilteredPaginations(t *testing.T) {
	app, ctx, appCodec := setupTest()

	var balances sdk.Coins

	for i := 0; i < numBalances; i++ {
		denom := fmt.Sprintf("foo%ddenom", i)
		balances = append(balances, sdk.NewInt64Coin(denom, 100))
	}

	balances = append(balances, sdk.NewInt64Coin("test1", 250))
	balances = append(balances, sdk.NewInt64Coin("test2", 250))
	balances = append(balances, sdk.NewInt64Coin("test3", 250))
	balances = append(balances, sdk.NewInt64Coin("test4", 250))

	addr1 := sdk.AccAddress([]byte("addr1"))
	acc1 := app.AccountKeeper.NewAccountWithAddress(ctx, addr1)
	app.AccountKeeper.SetAccount(ctx, acc1)
	require.NoError(t, app.BankKeeper.SetBalances(ctx, addr1, balances))

	// .Log("verify empty page request results a max of defaultLimit records and counts total records")

	pageReq := &query.PageRequest{Key: nil, Limit: 2, CountTotal: true}

	store := ctx.KVStore(app.GetKey(authtypes.StoreKey))
	balancesStore := prefix.NewStore(store, types.BalancesPrefix)
	accountStore := prefix.NewStore(balancesStore, addr1.Bytes())

	var balResult sdk.Coins
	res, err := query.FilteredPaginate(accountStore, pageReq, func(key []byte, value []byte, accumulate bool) (bool, error) {
		var bal sdk.Coin
		err := appCodec.UnmarshalBinaryBare(value, &bal)
		if err != nil {
			return false, err
		}

		if bal.Amount.Int64() > int64(100) {
			if accumulate {
				balResult = append(balResult, bal)
			}

			return true, nil
		}

		return false, nil
	})

	require.NoError(t, err)
	require.NotNil(t, res)
	require.Equal(t, 2, len(balResult))
}
