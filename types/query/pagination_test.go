package query_test

import (
	"github.com/cosmos/cosmos-sdk/simapp"
	"github.com/cosmos/cosmos-sdk/store"
	"github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/bank"
	paramkeeper "github.com/cosmos/cosmos-sdk/x/params/keeper"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"
	dbm "github.com/tendermint/tm-db"
	"testing"
)

func TestPagination(t *testing.T) {
	app, ctx, _ := SetupTest(t)

	balances := sdk.NewCoins(sdk.NewInt64Coin("foo", 100), sdk.NewInt64Coin("bar", 50))

	addr1 := sdk.AccAddress([]byte("addr1"))
	acc1 := app.AccountKeeper.NewAccountWithAddress(ctx, addr1)
	app.AccountKeeper.SetAccount(ctx, acc1)
	require.NoError(t, app.BankKeeper.SetBalances(ctx, addr1, balances))

	acc1Balances := app.BankKeeper.GetAllBalances(ctx, addr1)
	require.Equal(t, balances, acc1Balances)
}

func SetupTest(t *testing.T) (*simapp.SimApp, sdk.Context, types.CommitMultiStore) {
	app := simapp.Setup(false)
	ctx := app.BaseApp.NewContext(false, abci.Header{})

	keyAcc := sdk.NewKVStoreKey(auth.StoreKey)
	keyBank := sdk.NewKVStoreKey(bank.StoreKey)
	db := dbm.NewMemDB()
	ms := store.NewCommitMultiStore(db)

	err := ms.LoadLatestVersion()
	require.Nil(t, err)
	appCodec, _ := simapp.MakeCodecs()

	keyParams := sdk.NewKVStoreKey(paramtypes.StoreKey)
	tkeyParams := sdk.NewTransientStoreKey(paramtypes.TStoreKey)

	pk := paramkeeper.NewKeeper(appCodec, keyParams, tkeyParams)
	app.AccountKeeper = auth.NewAccountKeeper(
		appCodec, keyAcc, pk.Subspace(auth.DefaultParamspace), auth.ProtoBaseAccount, nil,
	)

	app.BankKeeper = bank.NewBaseKeeper(
		appCodec, keyBank, app.AccountKeeper, pk.Subspace(bank.DefaultParamspace), app.BlacklistedAccAddrs(),
	)

	return app, ctx, ms
}
