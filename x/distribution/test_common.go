// nolint
package distribution

import (
	abci "github.com/tendermint/tendermint/abci/types"
	dbm "github.com/tendermint/tendermint/libs/db"
	"github.com/tendermint/tendermint/libs/log"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	
	"github.com/cosmos/cosmos-sdk/x/auth"
	authTypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/distribution/types"
	"github.com/cosmos/cosmos-sdk/x/params/subspace"
	"github.com/cosmos/cosmos-sdk/x/supply/exported"
	"github.com/cosmos/cosmos-sdk/x/mock"
)

type testInput struct {
	cdc *codec.Codec
	ctx sdk.Context
	ak  auth.AccountKeeper
	sk  types.SupplyKeeper
	m   module.Manager
}

func setupTestInput() testInput {
	db := dbm.NewMemDB()

	cdc := codec.New()
	authTypes.RegisterCodec(cdc)
	cdc.RegisterInterface((*exported.ModuleAccountI)(nil), nil)
	cdc.RegisterConcrete(&mock.ModuleAccount{}, "cosmos-sdk/ModuleAccount", nil)
	codec.RegisterCrypto(cdc)

	authCapKey := sdk.NewKVStoreKey("authCapKey")
	keyParams := sdk.NewKVStoreKey("subspace")
	tkeyParams := sdk.NewTransientStoreKey("transient_subspace")

	ms := store.NewCommitMultiStore(db)
	ms.MountStoreWithDB(authCapKey, sdk.StoreTypeIAVL, db)
	ms.MountStoreWithDB(keyParams, sdk.StoreTypeIAVL, db)
	ms.MountStoreWithDB(tkeyParams, sdk.StoreTypeTransient, db)
	ms.LoadLatestVersion()

	ps := subspace.NewSubspace(cdc, keyParams, tkeyParams, authTypes.DefaultParamspace)
	ak := auth.NewAccountKeeper(cdc, authCapKey, ps, authTypes.ProtoBaseAccount)
	sk := mock.NewDummySupplyKeeper(ak)

	ctx := sdk.NewContext(ms, abci.Header{ChainID: "test-chain-id"}, false, log.NewNopLogger())

	ak.SetParams(ctx, authTypes.DefaultParams())

	return testInput{cdc: cdc, ctx: ctx, ak: ak, sk: sk}
}
