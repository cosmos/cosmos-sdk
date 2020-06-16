package query

import (
	"github.com/cosmos/cosmos-sdk/simapp"
	"github.com/cosmos/cosmos-sdk/store"
	sdk "github.com/cosmos/cosmos-sdk/types"
	auth "github.com/cosmos/cosmos-sdk/x/auth/types"
	staking "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"
	dbm "github.com/tendermint/tm-db"
	"testing"
)

func TestPagination(t *testing.T) {
	app, ctx := SetupTest(t)
	// TODO Add pagenation tests
}

func SetupTest(t *testing.T) (*simapp.SimApp, sdk.Context) {
	keyAcc := sdk.NewKVStoreKey(auth.StoreKey)
	keyStaking := sdk.NewKVStoreKey(staking.StoreKey)

	db := dbm.NewMemDB()
	ms := store.NewCommitMultiStore(db)

	ms.MountStoreWithDB(keyAcc, sdk.StoreTypeIAVL, db)
	ms.MountStoreWithDB(keyStaking, sdk.StoreTypeIAVL, db)

	err := ms.LoadLatestVersion()
	require.Nil(t, err)
	// setup
	app := simapp.Setup(false)
	ctx := app.BaseApp.NewContext(false, abci.Header{}) // make block height non-zero to ensure account numbers part of signBytes
	ctx = ctx.WithBlockHeight(1)
	_, _ = simapp.MakeCodecs()

	return app, ctx
}
