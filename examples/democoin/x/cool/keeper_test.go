package cool

import (
	"testing"

	"github.com/stretchr/testify/require"

	abci "github.com/tendermint/tendermint/abci/types"
	dbm "github.com/tendermint/tendermint/libs/db"

	"github.com/cosmos/cosmos-sdk/store"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/wire"
	auth "github.com/cosmos/cosmos-sdk/x/auth"
	bank "github.com/cosmos/cosmos-sdk/x/bank"
)

func setupMultiStore() (sdk.MultiStore, *sdk.KVStoreKey) {
	db := dbm.NewMemDB()
	capKey := sdk.NewKVStoreKey("capkey")
	ms := store.NewCommitMultiStore(db)
	ms.MountStoreWithDB(capKey, sdk.StoreTypeIAVL, db)
	ms.LoadLatestVersion()
	return ms, capKey
}

func TestCoolKeeper(t *testing.T) {
	ms, capKey := setupMultiStore()
	cdc := wire.NewCodec()
	auth.RegisterBaseAccount(cdc)

	am := auth.NewAccountMapper(cdc, capKey, auth.ProtoBaseAccount)
	ctx := sdk.NewContext(ms, abci.Header{}, false, nil)
	ck := bank.NewKeeper(am)
	keeper := NewKeeper(capKey, ck, DefaultCodespace)

	err := InitGenesis(ctx, keeper, Genesis{"icy"})
	require.Nil(t, err)

	genesis := WriteGenesis(ctx, keeper)
	require.Nil(t, err)
	require.Equal(t, genesis, Genesis{"icy"})

	res := keeper.GetTrend(ctx)
	require.Equal(t, res, "icy")

	keeper.setTrend(ctx, "fiery")
	res = keeper.GetTrend(ctx)
	require.Equal(t, res, "fiery")
}
