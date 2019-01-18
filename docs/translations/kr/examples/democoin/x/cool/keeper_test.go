package cool

import (
	"testing"

	"github.com/stretchr/testify/require"

	abci "github.com/tendermint/tendermint/abci/types"
	dbm "github.com/tendermint/tendermint/libs/db"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store"
	sdk "github.com/cosmos/cosmos-sdk/types"
	auth "github.com/cosmos/cosmos-sdk/x/auth"
	bank "github.com/cosmos/cosmos-sdk/x/bank"
	"github.com/cosmos/cosmos-sdk/x/params"
)

type testInput struct {
	cdc    *codec.Codec
	ctx    sdk.Context
	capKey *sdk.KVStoreKey
	bk     bank.BaseKeeper
}

func setupTestInput() testInput {
	db := dbm.NewMemDB()

	cdc := codec.New()
	auth.RegisterBaseAccount(cdc)

	capKey := sdk.NewKVStoreKey("capkey")
	keyParams := sdk.NewKVStoreKey("params")
	tkeyParams := sdk.NewTransientStoreKey("transient_params")

	ms := store.NewCommitMultiStore(db)
	ms.MountStoreWithDB(capKey, sdk.StoreTypeIAVL, db)
	ms.MountStoreWithDB(keyParams, sdk.StoreTypeIAVL, db)
	ms.MountStoreWithDB(tkeyParams, sdk.StoreTypeTransient, db)
	ms.LoadLatestVersion()

	pk := params.NewKeeper(cdc, keyParams, tkeyParams)
	ak := auth.NewAccountKeeper(cdc, capKey, pk.Subspace(auth.DefaultParamspace), auth.ProtoBaseAccount)
	bk := bank.NewBaseKeeper(ak)
	ctx := sdk.NewContext(ms, abci.Header{}, false, nil)

	return testInput{cdc: cdc, ctx: ctx, capKey: capKey, bk: bk}
}

func TestCoolKeeper(t *testing.T) {
	input := setupTestInput()
	keeper := NewKeeper(input.capKey, input.bk, DefaultCodespace)
	ctx := input.ctx

	err := InitGenesis(ctx, keeper, Genesis{"icy"})
	require.Nil(t, err)

	genesis := ExportGenesis(ctx, keeper)
	require.Nil(t, err)
	require.Equal(t, genesis, Genesis{"icy"})

	res := keeper.GetTrend(ctx)
	require.Equal(t, res, "icy")

	keeper.setTrend(ctx, "fiery")
	res = keeper.GetTrend(ctx)
	require.Equal(t, res, "fiery")
}
