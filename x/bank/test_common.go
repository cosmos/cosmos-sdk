// nolint
package bank

import (
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/params"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/libs/log"
	dbm "github.com/tendermint/tm-db"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store"
	sdk "github.com/cosmos/cosmos-sdk/types"
	autypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/bank/internal/types"
	"github.com/cosmos/cosmos-sdk/x/params/subspace"
)

type testInput struct {
	cdc *codec.Codec
	ctx sdk.Context
	ak  types.AccountKeeper
	bk  BaseKeeper
}

func setupTestInput() testInput {
	db := dbm.NewMemDB()
	cdc := codec.New()

	authCapKey := sdk.NewKVStoreKey("authCapKey")
	keyParams := sdk.NewKVStoreKey("subspace")
	tkeyParams := sdk.NewTransientStoreKey("transient_subspace")

	ms := store.NewCommitMultiStore(db)
	ms.MountStoreWithDB(authCapKey, sdk.StoreTypeIAVL, db)
	ms.MountStoreWithDB(keyParams, sdk.StoreTypeIAVL, db)
	ms.MountStoreWithDB(tkeyParams, sdk.StoreTypeTransient, db)
	if err := ms.LoadLatestVersion(); err != nil {
		panic(err)
	}

	pk := params.NewKeeper(types.ModuleCdc, keyParams, tkeyParams, params.DefaultCodespace)
	ps := subspace.NewSubspace(cdc, keyParams, tkeyParams, types.DefaultParamspace)
	ak := auth.NewAccountKeeper(cdc, authCapKey, ps, autypes.ProtoBaseAccount)
	bk := NewBaseKeeper(ak, pk.Subspace(DefaultParamspace), DefaultCodespace, nil)

	ctx := sdk.NewContext(ms, abci.Header{ChainID: "test-chain-id"}, false, log.NewNopLogger())
	bk.SetSendEnabled(ctx, true)

	return testInput{cdc: cdc, ctx: ctx, ak: ak, bk: bk}
}
