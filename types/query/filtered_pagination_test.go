package query_test

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/simapp"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/address"
	"github.com/cosmos/cosmos-sdk/types/query"
	"github.com/cosmos/cosmos-sdk/x/bank/types"
)

var addr1 = sdk.AccAddress([]byte("addr1"))

func (s *paginationTestSuite) TestFilteredPaginations() {
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

	balances = balances.Sort()
	addr1 := sdk.AccAddress([]byte("addr1"))
	acc1 := app.AccountKeeper.NewAccountWithAddress(ctx, addr1)
	app.AccountKeeper.SetAccount(ctx, acc1)
	s.Require().NoError(simapp.FundAccount(app.BankKeeper, ctx, addr1, balances))
	store := ctx.KVStore(app.GetKey(types.StoreKey))

	// verify pagination with limit > total values
	pageReq := &query.PageRequest{Key: nil, Limit: 5, CountTotal: true}
	balances, res, err := execFilterPaginate(store, pageReq, appCodec)
	s.Require().NoError(err)
	s.Require().NotNil(res)
	s.Require().Equal(4, len(balances))

	s.T().Log("verify empty request")
	balances, res, err = execFilterPaginate(store, nil, appCodec)
	s.Require().NoError(err)
	s.Require().NotNil(res)
	s.Require().Equal(4, len(balances))
	s.Require().Equal(uint64(4), res.Total)
	s.Require().Nil(res.NextKey)

	s.T().Log("verify nextKey is returned if there are more results")
	pageReq = &query.PageRequest{Key: nil, Limit: 2, CountTotal: true}
	balances, res, err = execFilterPaginate(store, pageReq, appCodec)
	s.Require().NoError(err)
	s.Require().NotNil(res)
	s.Require().Equal(2, len(balances))
	s.Require().NotNil(res.NextKey)
	s.Require().Equal(string(res.NextKey), fmt.Sprintf("test2denom"))
	s.Require().Equal(uint64(4), res.Total)

	s.T().Log("verify both key and offset can't be given")
	pageReq = &query.PageRequest{Key: res.NextKey, Limit: 1, Offset: 2, CountTotal: true}
	_, _, err = execFilterPaginate(store, pageReq, appCodec)
	s.Require().Error(err)

	s.T().Log("use nextKey for query")
	pageReq = &query.PageRequest{Key: res.NextKey, Limit: 2, CountTotal: true}
	balances, res, err = execFilterPaginate(store, pageReq, appCodec)
	s.Require().NoError(err)
	s.Require().NotNil(res)
	s.Require().Equal(2, len(balances))
	s.Require().Nil(res.NextKey)

	s.T().Log("verify default limit")
	pageReq = &query.PageRequest{Key: nil, Limit: 0}
	balances, res, err = execFilterPaginate(store, pageReq, appCodec)
	s.Require().NoError(err)
	s.Require().NotNil(res)
	s.Require().Equal(4, len(balances))
	s.Require().Equal(uint64(4), res.Total)

	s.T().Log("verify with offset")
	pageReq = &query.PageRequest{Offset: 2, Limit: 2}
	balances, res, err = execFilterPaginate(store, pageReq, appCodec)
	s.Require().NoError(err)
	s.Require().NotNil(res)
	s.Require().LessOrEqual(len(balances), 2)
}

func (s *paginationTestSuite) TestReverseFilteredPaginations() {
	app, ctx, appCodec := setupTest()

	var balances sdk.Coins
	for i := 0; i < numBalances; i++ {
		denom := fmt.Sprintf("foo%ddenom", i)
		balances = append(balances, sdk.NewInt64Coin(denom, 100))
	}

	for i := 0; i < 10; i++ {
		denom := fmt.Sprintf("test%ddenom", i)
		balances = append(balances, sdk.NewInt64Coin(denom, 250))
	}

	balances = balances.Sort()
	addr1 := sdk.AccAddress([]byte("addr1"))
	acc1 := app.AccountKeeper.NewAccountWithAddress(ctx, addr1)
	app.AccountKeeper.SetAccount(ctx, acc1)
	s.Require().NoError(simapp.FundAccount(app.BankKeeper, ctx, addr1, balances))
	store := ctx.KVStore(app.GetKey(types.StoreKey))

	// verify pagination with limit > total values
	pageReq := &query.PageRequest{Key: nil, Limit: 5, CountTotal: true, Reverse: true}
	balns, res, err := execFilterPaginate(store, pageReq, appCodec)
	s.Require().NoError(err)
	s.Require().NotNil(res)
	s.Require().Equal(5, len(balns))

	s.T().Log("verify empty request")
	balns, res, err = execFilterPaginate(store, nil, appCodec)
	s.Require().NoError(err)
	s.Require().NotNil(res)
	s.Require().Equal(10, len(balns))
	s.Require().Equal(uint64(10), res.Total)
	s.Require().Nil(res.NextKey)

	s.T().Log("verify default limit")
	pageReq = &query.PageRequest{Reverse: true}
	balns, res, err = execFilterPaginate(store, pageReq, appCodec)
	s.Require().NoError(err)
	s.Require().NotNil(res)
	s.Require().Equal(10, len(balns))
	s.Require().Equal(uint64(10), res.Total)

	s.T().Log("verify nextKey is returned if there are more results")
	pageReq = &query.PageRequest{Limit: 2, CountTotal: true, Reverse: true}
	balns, res, err = execFilterPaginate(store, pageReq, appCodec)
	s.Require().NoError(err)
	s.Require().NotNil(res)
	s.Require().Equal(2, len(balns))
	s.Require().NotNil(res.NextKey)
	s.Require().Equal(string(res.NextKey), fmt.Sprintf("test7denom"))
	s.Require().Equal(uint64(10), res.Total)

	s.T().Log("verify both key and offset can't be given")
	pageReq = &query.PageRequest{Key: res.NextKey, Limit: 1, Offset: 2, Reverse: true}
	_, _, err = execFilterPaginate(store, pageReq, appCodec)
	s.Require().Error(err)

	s.T().Log("use nextKey for query and reverse true")
	pageReq = &query.PageRequest{Key: res.NextKey, Limit: 2, Reverse: true}
	balns, res, err = execFilterPaginate(store, pageReq, appCodec)
	s.Require().NoError(err)
	s.Require().NotNil(res)
	s.Require().Equal(2, len(balns))
	s.Require().NotNil(res.NextKey)
	s.Require().Equal(string(res.NextKey), fmt.Sprintf("test5denom"))

	s.T().Log("verify last page records, nextKey for query and reverse true")
	pageReq = &query.PageRequest{Key: res.NextKey, Reverse: true}
	balns, res, err = execFilterPaginate(store, pageReq, appCodec)
	s.Require().NoError(err)
	s.Require().NotNil(res)
	s.Require().Equal(6, len(balns))
	s.Require().Nil(res.NextKey)

	s.T().Log("verify Reverse pagination returns valid result")
	s.Require().Equal(balances[235:241].String(), balns.Sort().String())
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

	balances = balances.Sort()
	addr1 := sdk.AccAddress([]byte("addr1"))
	acc1 := app.AccountKeeper.NewAccountWithAddress(ctx, addr1)
	app.AccountKeeper.SetAccount(ctx, acc1)
	err := simapp.FundAccount(app.BankKeeper, ctx, addr1, balances)
	if err != nil { // should return no error
		fmt.Println(err)
	}

	pageReq := &query.PageRequest{Key: nil, Limit: 1, CountTotal: true}
	store := ctx.KVStore(app.GetKey(types.StoreKey))
	balancesStore := prefix.NewStore(store, types.BalancesPrefix)
	accountStore := prefix.NewStore(balancesStore, address.MustLengthPrefix(addr1))

	var balResult sdk.Coins
	pageRes, err := query.FilteredPaginate(accountStore, pageReq, func(key, value []byte, accumulate bool) (bool, error) {
		var bal sdk.Coin
		err := appCodec.Unmarshal(value, &bal)
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

func execFilterPaginate(store sdk.KVStore, pageReq *query.PageRequest, appCodec codec.Codec) (balances sdk.Coins, res *query.PageResponse, err error) {
	balancesStore := prefix.NewStore(store, types.BalancesPrefix)
	accountStore := prefix.NewStore(balancesStore, address.MustLengthPrefix(addr1))

	var balResult sdk.Coins
	res, err = query.FilteredPaginate(accountStore, pageReq, func(key, value []byte, accumulate bool) (bool, error) {
		var bal sdk.Coin
		err := appCodec.Unmarshal(value, &bal)
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
