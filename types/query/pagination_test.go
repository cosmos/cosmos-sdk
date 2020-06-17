package query_test

import (
	gocontext "context"
	"fmt"
	"testing"

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
	holder     = "holder"
	multiPerm  = "multiple permissions account"
	randomPerm = "random permission"
)

func TestPagination(t *testing.T) {
	app, ctx := SetupTest(t)
	queryHelper := baseapp.NewQueryServerTestHelper(ctx)
	types.RegisterQueryServer(queryHelper, app.BankKeeper)
	queryClient := types.NewQueryClient(queryHelper)

	var balances sdk.Coins
	const (
		numBalances     = 235
		maxLimit        = 100
		overLimit       = 101
		underLimit      = 10
		lastPageRecords = 35
	)

	for i := 0; i < numBalances; i++ {
		denom := fmt.Sprintf("foo%ddenom", i)
		balances = append(balances, sdk.NewInt64Coin(denom, 100))
	}

	addr1 := sdk.AccAddress([]byte("addr1"))
	acc1 := app.AccountKeeper.NewAccountWithAddress(ctx, addr1)
	app.AccountKeeper.SetAccount(ctx, acc1)
	require.NoError(t, app.BankKeeper.SetBalances(ctx, addr1, balances))

	t.Log("verify empty page request results a max of maxLimit records")
	pageReq := &query.PageRequest{}
	request := types.NewQueryAllBalancesRequest(addr1, pageReq)
	res, err := queryClient.AllBalances(gocontext.Background(), request)
	require.NoError(t, err)
	require.Equal(t, res.Res.Total, uint64(0))
	require.NotNil(t, res.Res.NextKey)
	require.LessOrEqual(t, res.Balances.Len(), maxLimit)

	t.Log("verify page request with limit > maxLimit, returns only `maxLimit` records")
	pageReq = &query.PageRequest{Limit: overLimit}
	request = types.NewQueryAllBalancesRequest(addr1, pageReq)
	res, err = queryClient.AllBalances(gocontext.Background(), request)
	require.NoError(t, err)
	require.Equal(t, res.Res.Total, uint64(0))
	require.NotNil(t, res.Res.NextKey)
	require.LessOrEqual(t, res.Balances.Len(), maxLimit)

	t.Log("verify paginate with custom limit and countTotal true")
	pageReq = &query.PageRequest{Limit: underLimit, CountTotal: true}
	request = types.NewQueryAllBalancesRequest(addr1, pageReq)
	res, err = queryClient.AllBalances(gocontext.Background(), request)
	require.NoError(t, err)
	require.Equal(t, res.Balances.Len(), underLimit)
	require.Nil(t, res.Res.NextKey)
	require.Equal(t, res.Res.Total, uint64(numBalances))

	t.Log("verify paginate with custom limit and countTotal false")
	pageReq = &query.PageRequest{Limit: maxLimit, CountTotal: false}
	request = types.NewQueryAllBalancesRequest(addr1, pageReq)
	res, err = queryClient.AllBalances(gocontext.Background(), request)
	require.NoError(t, err)
	require.Equal(t, res.Balances.Len(), maxLimit)
	require.NotNil(t, res.Res.NextKey)
	require.Equal(t, res.Res.Total, uint64(0))

	t.Log("verify paginate with custom limit, key and countTotal false")
	pageReq = &query.PageRequest{Key: res.Res.NextKey, Limit: maxLimit, CountTotal: false}
	request = types.NewQueryAllBalancesRequest(addr1, pageReq)
	res, err = queryClient.AllBalances(gocontext.Background(), request)
	require.NoError(t, err)
	require.Equal(t, res.Balances.Len(), maxLimit)
	require.NotNil(t, res.Res.NextKey)
	require.Equal(t, res.Res.Total, uint64(0))

	t.Log("verify paginate for last page, results in records less than max limit")
	pageReq = &query.PageRequest{Key: res.Res.NextKey, Limit: maxLimit, CountTotal: false}
	request = types.NewQueryAllBalancesRequest(addr1, pageReq)
	res, err = queryClient.AllBalances(gocontext.Background(), request)
	require.NoError(t, err)
	require.LessOrEqual(t, res.Balances.Len(), maxLimit)
	require.Equal(t, res.Balances.Len(), lastPageRecords)
	require.Nil(t, res.Res.NextKey)
	require.Equal(t, res.Res.Total, uint64(0))

	t.Log("verify paginate with offset and limit")
	pageReq = &query.PageRequest{Offset: 200, Limit: maxLimit, CountTotal: false}
	request = types.NewQueryAllBalancesRequest(addr1, pageReq)
	res, err = queryClient.AllBalances(gocontext.Background(), request)
	require.NoError(t, err)
	require.LessOrEqual(t, res.Balances.Len(), maxLimit)
	require.Equal(t, res.Balances.Len(), lastPageRecords)
	require.Nil(t, res.Res.NextKey)
	require.Equal(t, res.Res.Total, uint64(0))

	t.Log("verify paginate with offset and limit")
	pageReq = &query.PageRequest{Offset: 100, Limit: maxLimit, CountTotal: false}
	request = types.NewQueryAllBalancesRequest(addr1, pageReq)
	res, err = queryClient.AllBalances(gocontext.Background(), request)
	require.NoError(t, err)
	require.LessOrEqual(t, res.Balances.Len(), maxLimit)
	require.NotNil(t, res.Res.NextKey)
	require.Equal(t, res.Res.Total, uint64(0))

	t.Log("verify paginate with offset and key - error")
	pageReq = &query.PageRequest{Key: res.Res.NextKey, Offset: 100, Limit: maxLimit, CountTotal: false}
	request = types.NewQueryAllBalancesRequest(addr1, pageReq)
	res, err = queryClient.AllBalances(gocontext.Background(), request)
	require.Error(t, err)
	require.Equal(t, err.Error(), "invalid request, either offset or key is expected, got both")

	t.Log("verify paginate with offset greater than total results")
	pageReq = &query.PageRequest{Offset: 300, Limit: maxLimit, CountTotal: false}
	request = types.NewQueryAllBalancesRequest(addr1, pageReq)
	res, err = queryClient.AllBalances(gocontext.Background(), request)
	require.NoError(t, err)
	require.LessOrEqual(t, res.Balances.Len(), 0)
	require.Nil(t, res.Res.NextKey)
}

func SetupTest(t *testing.T) (*simapp.SimApp, sdk.Context) {
	app := simapp.Setup(false)
	ctx := app.BaseApp.NewContext(false, abci.Header{Height: 1})
	appCodec := app.AppCodec()

	db := dbm.NewMemDB()
	ms := store.NewCommitMultiStore(db)

	err := ms.LoadLatestVersion()
	require.Nil(t, err)

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

	return app, ctx
}
