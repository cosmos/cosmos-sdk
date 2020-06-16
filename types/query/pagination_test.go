package query_test

import (
	"github.com/cosmos/cosmos-sdk/simapp"
	"github.com/cosmos/cosmos-sdk/store"
	sdk "github.com/cosmos/cosmos-sdk/types"
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
