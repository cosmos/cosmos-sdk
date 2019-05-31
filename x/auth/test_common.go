// nolint
package auth

import (
	abci "github.com/tendermint/tendermint/abci/types"
	dbm "github.com/tendermint/tendermint/libs/db"
	"github.com/tendermint/tendermint/libs/log"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/params/subspace"
)

type testInput struct {
	cdc *codec.Codec
	ctx sdk.Context
	ak  AccountKeeper
	fck types.FeeCollectionKeeper
}

func setupTestInput() testInput {
	db := dbm.NewMemDB()

	cdc := codec.New()
	types.RegisterBaseAccount(cdc)

	authCapKey := sdk.NewKVStoreKey("authCapKey")
	fckCapKey := sdk.NewKVStoreKey("fckCapKey")
	keyParams := sdk.NewKVStoreKey("subspace")
	tkeyParams := sdk.NewTransientStoreKey("transient_subspace")

	ms := store.NewCommitMultiStore(db)
	ms.MountStoreWithDB(authCapKey, sdk.StoreTypeIAVL, db)
	ms.MountStoreWithDB(fckCapKey, sdk.StoreTypeIAVL, db)
	ms.MountStoreWithDB(keyParams, sdk.StoreTypeIAVL, db)
	ms.MountStoreWithDB(tkeyParams, sdk.StoreTypeTransient, db)
	ms.LoadLatestVersion()

	ps := subspace.NewSubspace(cdc, keyParams, tkeyParams, types.DefaultParamspace)
	ak := NewAccountKeeper(cdc, authCapKey, ps, types.ProtoBaseAccount)
	fck := types.NewFeeCollectionKeeper(cdc, fckCapKey)
	ctx := sdk.NewContext(ms, abci.Header{ChainID: "test-chain-id"}, false, log.NewNopLogger())

	ak.SetParams(ctx, types.DefaultParams())

	return testInput{cdc: cdc, ctx: ctx, ak: ak, fck: fck}
}
