package pow

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

func TestPowKeeperGetSet(t *testing.T) {
	input := setupTestInput()
	ctx := input.ctx

	config := NewConfig("pow", int64(1))
	keeper := NewKeeper(input.capKey, config, input.bk, DefaultCodespace)

	err := InitGenesis(ctx, keeper, Genesis{uint64(1), uint64(0)})
	require.Nil(t, err)

	genesis := ExportGenesis(ctx, keeper)
	require.Nil(t, err)
	require.Equal(t, genesis, Genesis{uint64(1), uint64(0)})

	res, err := keeper.GetLastDifficulty(ctx)
	require.Nil(t, err)
	require.Equal(t, res, uint64(1))

	keeper.SetLastDifficulty(ctx, 2)

	res, err = keeper.GetLastDifficulty(ctx)
	require.Nil(t, err)
	require.Equal(t, res, uint64(2))
}
