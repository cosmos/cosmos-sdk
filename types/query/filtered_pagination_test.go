package query_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/KiraCore/cosmos-sdk/codec"
	"github.com/KiraCore/cosmos-sdk/store/prefix"
	sdk "github.com/KiraCore/cosmos-sdk/types"
	"github.com/KiraCore/cosmos-sdk/types/query"
	authtypes "github.com/KiraCore/cosmos-sdk/x/auth/types"
	"github.com/KiraCore/cosmos-sdk/x/bank/types"
)

var addr1 = sdk.AccAddress([]byte("addr1"))

func TestFilteredPaginations(t *testing.T) {
	app, ctx, appCodec := setupTest()

	var balances sdk.Coins
	for i := 0; i < numBalances; i++ {
		denom := fmt.Sprintf("foo%ddenom", i)
		balances = append(balances, sdk.NewInt64Coin(denom, 100))
	}

	for i := 0; i < 4; i++ {
		denom := fmt.Sprintf("test%ddenom", i)
		balances = append(balances, sdk.NewInt64Coin(denom, 250))
	}

	addr1 := sdk.AccAddress([]byte("addr1"))
	acc1 := app.AccountKeeper.NewAccountWithAddress(ctx, addr1)
	app.AccountKeeper.SetAccount(ctx, acc1)
	require.NoError(t, app.BankKeeper.SetBalances(ctx, addr1, balances))
	store := ctx.KVStore(app.GetKey(authtypes.StoreKey))

	// verify pagination with limit > total values
	pageReq := &query.PageRequest{Key: nil, Limit: 5, CountTotal: true}
	balances, res, err := execFilterPaginate(store, pageReq, appCodec)
	require.NoError(t, err)
	require.NotNil(t, res)
	require.Equal(t, 4, len(balances))

	t.Log("verify empty request")
	balances, res, err = execFilterPaginate(store, nil, appCodec)
	require.NoError(t, err)
	require.NotNil(t, res)
	require.Equal(t, 4, len(balances))
	require.Equal(t, uint64(4), res.Total)
	require.Nil(t, res.NextKey)

	t.Log("verify nextKey is returned if there are more results")
	pageReq = &query.PageRequest{Key: nil, Limit: 2, CountTotal: true}
	balances, res, err = execFilterPaginate(store, pageReq, appCodec)
	require.NoError(t, err)
	require.NotNil(t, res)
	require.Equal(t, 2, len(balances))
	require.NotNil(t, res.NextKey)
	require.Equal(t, string(res.NextKey), fmt.Sprintf("test2denom"))
	require.Equal(t, uint64(4), res.Total)

	t.Log("verify both key and offset can't be given")
	pageReq = &query.PageRequest{Key: res.NextKey, Limit: 1, Offset: 2, CountTotal: true}
	_, _, err = execFilterPaginate(store, pageReq, appCodec)
	require.Error(t, err)

	t.Log("use nextKey for query")
	pageReq = &query.PageRequest{Key: res.NextKey, Limit: 2, CountTotal: true}
	balances, res, err = execFilterPaginate(store, pageReq, appCodec)
	require.NoError(t, err)
	require.NotNil(t, res)
	require.Equal(t, 2, len(balances))
	require.Nil(t, res.NextKey)

	t.Log("verify default limit")
	pageReq = &query.PageRequest{Key: nil, Limit: 0}
	balances, res, err = execFilterPaginate(store, pageReq, appCodec)
	require.NoError(t, err)
	require.NotNil(t, res)
	require.Equal(t, 4, len(balances))
	require.Equal(t, uint64(4), res.Total)

	t.Log("verify with offset")
	pageReq = &query.PageRequest{Offset: 2, Limit: 2}
	balances, res, err = execFilterPaginate(store, pageReq, appCodec)
	require.NoError(t, err)
	require.NotNil(t, res)
	require.LessOrEqual(t, len(balances), 2)
}

func ExampleFilteredPaginate() {
	app, ctx, appCodec := setupTest()

	var balances sdk.Coins
	for i := 0; i < numBalances; i++ {
		denom := fmt.Sprintf("foo%ddenom", i)
		balances = append(balances, sdk.NewInt64Coin(denom, 100))
	}

	for i := 0; i < 5; i++ {
		denom := fmt.Sprintf("test%ddenom", i)
		balances = append(balances, sdk.NewInt64Coin(denom, 250))
	}
	addr1 := sdk.AccAddress([]byte("addr1"))
	acc1 := app.AccountKeeper.NewAccountWithAddress(ctx, addr1)
	app.AccountKeeper.SetAccount(ctx, acc1)
	err := app.BankKeeper.SetBalances(ctx, addr1, balances)
	if err != nil { // should return no error
		fmt.Println(err)
	}

	pageReq := &query.PageRequest{Key: nil, Limit: 1, CountTotal: true}
	store := ctx.KVStore(app.GetKey(authtypes.StoreKey))
	balancesStore := prefix.NewStore(store, types.BalancesPrefix)
	accountStore := prefix.NewStore(balancesStore, addr1.Bytes())

	var balResult sdk.Coins
	pageRes, err := query.FilteredPaginate(accountStore, pageReq, func(key []byte, value []byte, accumulate bool) (bool, error) {
		var bal sdk.Coin
		err := appCodec.UnmarshalBinaryBare(value, &bal)
		if err != nil {
			return false, err
		}

		// filter balances with amount greater than 100
		if bal.Amount.Int64() > int64(100) {
			if accumulate {
				balResult = append(balResult, bal)
			}

			return true, nil
		}

		return false, nil
	})

	if err != nil { // should return no error
		fmt.Println(err)
	}
	fmt.Println(&types.QueryAllBalancesResponse{Balances: balResult, Pagination: pageRes})
	// Output:
	// balances:<denom:"test0denom" amount:"250" > pagination:<next_key:"test1denom" total:5 >
}

func execFilterPaginate(store sdk.KVStore, pageReq *query.PageRequest, appCodec codec.Marshaler) (balances sdk.Coins, res *query.PageResponse, err error) {
	balancesStore := prefix.NewStore(store, types.BalancesPrefix)
	accountStore := prefix.NewStore(balancesStore, addr1.Bytes())

	var balResult sdk.Coins
	res, err = query.FilteredPaginate(accountStore, pageReq, func(key []byte, value []byte, accumulate bool) (bool, error) {
		var bal sdk.Coin
		err := appCodec.UnmarshalBinaryBare(value, &bal)
		if err != nil {
			return false, err
		}

		// filter balances with amount greater than 100
		if bal.Amount.Int64() > int64(100) {
			if accumulate {
				balResult = append(balResult, bal)
			}

			return true, nil
		}

		return false, nil
	})

	return balResult, res, err
}
