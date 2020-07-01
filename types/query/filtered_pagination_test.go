package query_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/bank/types"
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

	// verify empty request
	balances, res, err = execFilterPaginate(store, nil, appCodec)

	require.NoError(t, err)
	require.NotNil(t, res)
	require.Equal(t, 4, len(balances))
	require.Equal(t, uint64(4), res.Total)

	// verify next key is returned
	pageReq = &query.PageRequest{Key: nil, Limit: 2, CountTotal: true}
	balances, res, err = execFilterPaginate(store, pageReq, appCodec)

	require.NoError(t, err)
	require.NotNil(t, res)
	require.Equal(t, 2, len(balances))
	require.NotNil(t, res.NextKey)
	require.Equal(t, uint64(4), res.Total)

	// use next key for query
	pageReq = &query.PageRequest{Key: res.NextKey, Limit: 2, CountTotal: true}
	balances, res, err = execFilterPaginate(store, pageReq, appCodec)

	require.NoError(t, err)
	require.NotNil(t, res)
	require.Equal(t, 2, len(balances))
	require.NotNil(t, res.NextKey)

	// verify both key and offset can't be given
	pageReq = &query.PageRequest{Key: res.NextKey, Limit: 1, Offset: 2, CountTotal: true}
	balances, res, err = execFilterPaginate(store, pageReq, appCodec)
	require.Error(t, err)

	// verify default limit
	pageReq = &query.PageRequest{Key: nil, Limit: 0}
	balances, res, err = execFilterPaginate(store, pageReq, appCodec)

	require.NoError(t, err)
	require.NotNil(t, res)
	require.Equal(t, 4, len(balances))
	require.Equal(t, uint64(4), res.Total)

	// verify offset
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
	res, err := query.FilteredPaginate(accountStore, pageReq, func(key []byte, value []byte, accumulate bool) (bool, error) {
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
	fmt.Println(&types.QueryAllBalancesResponse{Balances: balResult, Res: res})
	// Output:
	// balances:<denom:"test0denom" amount:"250" > res:<next_key:"test0denom" total:5 >
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
