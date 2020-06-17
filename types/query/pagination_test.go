package query_test

import (
	"github.com/cosmos/cosmos-sdk/simapp"
	"github.com/cosmos/cosmos-sdk/store"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/bank"
	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"
	dbm "github.com/tendermint/tm-db"
	"testing"
)

const (
	holder     = "holder"
	multiPerm  = "multiple permissions account"
	randomPerm = "random permission"
)

func TestPagination(t *testing.T) {
	app, ctx := SetupTest(t)

	balances := sdk.NewCoins(sdk.NewInt64Coin("foo", 100), sdk.NewInt64Coin("bar", 50))

	addr1 := sdk.AccAddress([]byte("addr1"))
	acc1 := app.AccountKeeper.NewAccountWithAddress(ctx, addr1)
	app.AccountKeeper.SetAccount(ctx, acc1)
	require.NoError(t, app.BankKeeper.SetBalances(ctx, addr1, balances))

	acc1Balances := app.BankKeeper.GetAllBalances(ctx, addr1)
	require.Equal(t, balances, acc1Balances)

	// verify limit
	pageReq := &query.PageRequest{Key: nil, Limit: 2, CountTotal: true}
	balances, res, err := app.BankKeeper.QueryAllBalances(ctx, addr1, pageReq)
	require.NoError(t, err)
	require.NotNil(t, res)
	require.Equal(t, acc1Balances, balances)

	// verify paginate with multiple keys
	pageReq = &query.PageRequest{Key: nil, Limit: 1, CountTotal: true}
	balances, res, err = app.BankKeeper.QueryAllBalances(ctx, addr1, pageReq)
	require.NoError(t, err)
	require.NotNil(t, res)
	require.Equal(t, len(balances), 1)
	require.NotNil(t, res.NextKey)

	pageReq = &query.PageRequest{Key: res.NextKey, Limit: 1, CountTotal: true}
	balances, res, err = app.BankKeeper.QueryAllBalances(ctx, addr1, pageReq)
	require.NoError(t, err)
	require.NotNil(t, res)
	require.Equal(t, len(balances), 1)

	//verify offset usage over key
	pageReq = &query.PageRequest{Offset: 1, Limit: 1, CountTotal: true}
	balances, res, err = app.BankKeeper.QueryAllBalances(ctx, addr1, pageReq)
	require.NoError(t, err)
	require.NotNil(t, res)
	require.Equal(t,1,len(balances))
	require.Nil(t, res.NextKey)

	//verify offset usage over key and countTotal false
	pageReq = &query.PageRequest{Offset: 1, Limit: 1, CountTotal: false}
	balances, res, err = app.BankKeeper.QueryAllBalances(ctx, addr1, pageReq)
	require.NoError(t, err)
	require.NotNil(t, res)
	require.Equal(t,1,len(balances))
	require.NotNil(t, res.NextKey)

	//verify offset usage over key
	pageReq = &query.PageRequest{Offset: 1, Limit: 2, CountTotal: false}
	balances, res, err = app.BankKeeper.QueryAllBalances(ctx, addr1, pageReq)
	require.NoError(t, err)
	require.NotNil(t, res)
	require.Equal(t,2,len(balances))
	require.Nil(t, res.NextKey)

	//verify offset usage over key
	pageReq = &query.PageRequest{Offset: 1, Limit: 1, CountTotal: true}
	balances, res, err = app.BankKeeper.QueryAllBalances(ctx, addr1, pageReq)
	require.NoError(t, err)
	require.NotNil(t, res)
	require.Equal(t,1,len(balances))
	require.Nil(t, res.NextKey)

	t.Log("verify not in range offset")
	pageReq = &query.PageRequest{Offset: 3, Limit: 1, CountTotal: false}
	balances, res, err = app.BankKeeper.QueryAllBalances(ctx, addr1, pageReq)
	require.NoError(t, err)
	require.NotNil(t, res)
	require.Equal(t,0,len(balances))
	require.Nil(t, res.NextKey)

	// verify if total is not returned
	pageReq = &query.PageRequest{Key: nil, Limit: 2, CountTotal: false}
	balances, res, err = app.BankKeeper.QueryAllBalances(ctx, addr1, pageReq)
	require.NoError(t, err)
	require.NotNil(t, res)
	require.Equal(t, uint64(0), res.Total)
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
	maccPerms[auth.Burner] = []string{auth.Burner}
	maccPerms[auth.Minter] = []string{auth.Minter}
	maccPerms[multiPerm] = []string{auth.Burner, auth.Minter, auth.Staking}
	maccPerms[randomPerm] = []string{"random"}
	app.AccountKeeper = auth.NewAccountKeeper(
		appCodec, app.GetKey(auth.StoreKey), app.GetSubspace(auth.ModuleName),
		auth.ProtoBaseAccount, maccPerms,
	)
	app.BankKeeper = bank.NewBaseKeeper(
		appCodec, app.GetKey(auth.StoreKey), app.AccountKeeper,
		app.GetSubspace(bank.ModuleName), make(map[string]bool),
	)

	return app, ctx
}
