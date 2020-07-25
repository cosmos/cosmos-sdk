package query_test

import (
	gocontext "context"
	"fmt"
	"testing"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"

	"github.com/cosmos/cosmos-sdk/baseapp"

	"github.com/cosmos/cosmos-sdk/x/bank/types"

	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"
	dbm "github.com/tendermint/tm-db"

	"github.com/cosmos/cosmos-sdk/simapp"
	"github.com/cosmos/cosmos-sdk/store"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
)

const (
	holder          = "holder"
	multiPerm       = "multiple permissions account"
	randomPerm      = "random permission"
	numBalances     = 235
	defaultLimit    = 100
	overLimit       = 101
	underLimit      = 10
	lastPageRecords = 35
)

func TestPagination(t *testing.T) {
	app, ctx, _ := setupTest()
	queryHelper := baseapp.NewQueryServerTestHelper(ctx, app.InterfaceRegistry())
	types.RegisterQueryServer(queryHelper, app.BankKeeper)
	queryClient := types.NewQueryClient(queryHelper)

	var balances sdk.Coins

	for i := 0; i < numBalances; i++ {
		denom := fmt.Sprintf("foo%ddenom", i)
		balances = append(balances, sdk.NewInt64Coin(denom, 100))
	}

	addr1 := sdk.AccAddress([]byte("addr1"))
	acc1 := app.AccountKeeper.NewAccountWithAddress(ctx, addr1)
	app.AccountKeeper.SetAccount(ctx, acc1)
	require.NoError(t, app.BankKeeper.SetBalances(ctx, addr1, balances))

	t.Log("verify empty page request results a max of defaultLimit records and counts total records")
	pageReq := &query.PageRequest{}
	request := types.NewQueryAllBalancesRequest(addr1, pageReq)
	res, err := queryClient.AllBalances(gocontext.Background(), request)
	require.NoError(t, err)
	require.Equal(t, res.Pagination.Total, uint64(numBalances))
	require.NotNil(t, res.Pagination.NextKey)
	require.LessOrEqual(t, res.Balances.Len(), defaultLimit)

	t.Log("verify page request with limit > defaultLimit, returns less or equal to `limit` records")
	pageReq = &query.PageRequest{Limit: overLimit}
	request = types.NewQueryAllBalancesRequest(addr1, pageReq)
	res, err = queryClient.AllBalances(gocontext.Background(), request)
	require.NoError(t, err)
	require.Equal(t, res.Pagination.Total, uint64(0))
	require.NotNil(t, res.Pagination.NextKey)
	require.LessOrEqual(t, res.Balances.Len(), overLimit)

	t.Log("verify paginate with custom limit and countTotal true")
	pageReq = &query.PageRequest{Limit: underLimit, CountTotal: true}
	request = types.NewQueryAllBalancesRequest(addr1, pageReq)
	res, err = queryClient.AllBalances(gocontext.Background(), request)
	require.NoError(t, err)
	require.Equal(t, res.Balances.Len(), underLimit)
	require.NotNil(t, res.Pagination.NextKey)
	require.Equal(t, res.Pagination.Total, uint64(numBalances))

	t.Log("verify paginate with custom limit and countTotal false")
	pageReq = &query.PageRequest{Limit: defaultLimit, CountTotal: false}
	request = types.NewQueryAllBalancesRequest(addr1, pageReq)
	res, err = queryClient.AllBalances(gocontext.Background(), request)
	require.NoError(t, err)
	require.Equal(t, res.Balances.Len(), defaultLimit)
	require.NotNil(t, res.Pagination.NextKey)
	require.Equal(t, res.Pagination.Total, uint64(0))

	t.Log("verify paginate with custom limit, key and countTotal false")
	pageReq = &query.PageRequest{Key: res.Pagination.NextKey, Limit: defaultLimit, CountTotal: false}
	request = types.NewQueryAllBalancesRequest(addr1, pageReq)
	res, err = queryClient.AllBalances(gocontext.Background(), request)
	require.NoError(t, err)
	require.Equal(t, res.Balances.Len(), defaultLimit)
	require.NotNil(t, res.Pagination.NextKey)
	require.Equal(t, res.Pagination.Total, uint64(0))

	t.Log("verify paginate for last page, results in records less than max limit")
	pageReq = &query.PageRequest{Key: res.Pagination.NextKey, Limit: defaultLimit, CountTotal: false}
	request = types.NewQueryAllBalancesRequest(addr1, pageReq)
	res, err = queryClient.AllBalances(gocontext.Background(), request)
	require.NoError(t, err)
	require.LessOrEqual(t, res.Balances.Len(), defaultLimit)
	require.Equal(t, res.Balances.Len(), lastPageRecords)
	require.Nil(t, res.Pagination.NextKey)
	require.Equal(t, res.Pagination.Total, uint64(0))

	t.Log("verify paginate with offset and limit")
	pageReq = &query.PageRequest{Offset: 200, Limit: defaultLimit, CountTotal: false}
	request = types.NewQueryAllBalancesRequest(addr1, pageReq)
	res, err = queryClient.AllBalances(gocontext.Background(), request)
	require.NoError(t, err)
	require.LessOrEqual(t, res.Balances.Len(), defaultLimit)
	require.Equal(t, res.Balances.Len(), lastPageRecords)
	require.Nil(t, res.Pagination.NextKey)
	require.Equal(t, res.Pagination.Total, uint64(0))

	t.Log("verify paginate with offset and limit")
	pageReq = &query.PageRequest{Offset: 100, Limit: defaultLimit, CountTotal: false}
	request = types.NewQueryAllBalancesRequest(addr1, pageReq)
	res, err = queryClient.AllBalances(gocontext.Background(), request)
	require.NoError(t, err)
	require.LessOrEqual(t, res.Balances.Len(), defaultLimit)
	require.NotNil(t, res.Pagination.NextKey)
	require.Equal(t, res.Pagination.Total, uint64(0))

	t.Log("verify paginate with offset and key - error")
	pageReq = &query.PageRequest{Key: res.Pagination.NextKey, Offset: 100, Limit: defaultLimit, CountTotal: false}
	request = types.NewQueryAllBalancesRequest(addr1, pageReq)
	res, err = queryClient.AllBalances(gocontext.Background(), request)
	require.Error(t, err)
	require.Equal(t, err.Error(), "invalid request, either offset or key is expected, got both")

	t.Log("verify paginate with offset greater than total results")
	pageReq = &query.PageRequest{Offset: 300, Limit: defaultLimit, CountTotal: false}
	request = types.NewQueryAllBalancesRequest(addr1, pageReq)
	res, err = queryClient.AllBalances(gocontext.Background(), request)
	require.NoError(t, err)
	require.LessOrEqual(t, res.Balances.Len(), 0)
	require.Nil(t, res.Pagination.NextKey)
}

func ExamplePaginate() {
	app, ctx, _ := setupTest()

	var balances sdk.Coins

	for i := 0; i < 2; i++ {
		denom := fmt.Sprintf("foo%ddenom", i)
		balances = append(balances, sdk.NewInt64Coin(denom, 100))
	}

	addr1 := sdk.AccAddress([]byte("addr1"))
	acc1 := app.AccountKeeper.NewAccountWithAddress(ctx, addr1)
	app.AccountKeeper.SetAccount(ctx, acc1)
	err := app.BankKeeper.SetBalances(ctx, addr1, balances)
	if err != nil {
		fmt.Println(err)
	}
	// Paginate example
	pageReq := &query.PageRequest{Key: nil, Limit: 1, CountTotal: true}
	request := types.NewQueryAllBalancesRequest(addr1, pageReq)
	balResult := sdk.NewCoins()
	authStore := ctx.KVStore(app.GetKey(authtypes.StoreKey))
	balancesStore := prefix.NewStore(authStore, types.BalancesPrefix)
	accountStore := prefix.NewStore(balancesStore, addr1.Bytes())
	pageRes, err := query.Paginate(accountStore, request.Pagination, func(key []byte, value []byte) error {
		var tempRes sdk.Coin
		err := app.Codec().UnmarshalBinaryBare(value, &tempRes)
		if err != nil {
			return err
		}
		balResult = append(balResult, tempRes)
		return nil
	})
	if err != nil { // should return no error
		fmt.Println(err)
	}
	fmt.Println(&types.QueryAllBalancesResponse{Balances: balResult, Pagination: pageRes})
	// Output:
	// balances:<denom:"foo0denom" amount:"100" > pagination:<next_key:"foo1denom" total:2 >
}

func setupTest() (*simapp.SimApp, sdk.Context, codec.Marshaler) {
	app := simapp.Setup(false)
	ctx := app.BaseApp.NewContext(false, abci.Header{Height: 1})
	appCodec := app.AppCodec()

	db := dbm.NewMemDB()
	ms := store.NewCommitMultiStore(db)

	ms.LoadLatestVersion()

	maccPerms := simapp.GetMaccPerms()
	maccPerms[holder] = nil
	maccPerms[authtypes.Burner] = []string{authtypes.Burner}
	maccPerms[authtypes.Minter] = []string{authtypes.Minter}
	maccPerms[multiPerm] = []string{authtypes.Burner, authtypes.Minter, authtypes.Staking}
	maccPerms[randomPerm] = []string{"random"}
	app.AccountKeeper = authkeeper.NewAccountKeeper(
		appCodec, app.GetKey(authtypes.StoreKey), app.GetSubspace(authtypes.ModuleName),
		authtypes.ProtoBaseAccount, maccPerms,
	)
	app.BankKeeper = bankkeeper.NewBaseKeeper(
		appCodec, app.GetKey(authtypes.StoreKey), app.AccountKeeper,
		app.GetSubspace(types.ModuleName), make(map[string]bool),
	)

	return app, ctx, appCodec
}
