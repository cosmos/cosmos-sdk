package keeper

import (
	"time"

	"github.com/tendermint/tendermint/crypto/ed25519"

	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/libs/log"
	dbm "github.com/tendermint/tm-db"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/bank"
	"github.com/cosmos/cosmos-sdk/x/msg_authorization/internal/types"
	"github.com/cosmos/cosmos-sdk/x/params"
	"github.com/cosmos/cosmos-sdk/x/staking"
	"github.com/cosmos/cosmos-sdk/x/supply"
)

func makeTestCodec() *codec.Codec {
	var cdc = codec.New()
	auth.RegisterCodec(cdc)
	types.RegisterCodec(cdc)
	supply.RegisterCodec(cdc)
	staking.RegisterCodec(cdc)
	sdk.RegisterCodec(cdc)
	codec.RegisterCrypto(cdc)

	return cdc
}
func SetupTestInput() (sdk.Context, auth.AccountKeeper, params.Keeper, bank.BaseKeeper, Keeper, sdk.Router) {
	db := dbm.NewMemDB()

	cdc := codec.New()
	auth.RegisterCodec(cdc)
	bank.RegisterCodec(cdc)
	sdk.RegisterCodec(cdc)
	codec.RegisterCrypto(cdc)

	keyAcc := sdk.NewKVStoreKey(auth.StoreKey)
	keyParams := sdk.NewKVStoreKey(params.StoreKey)
	keyAuthorization := sdk.NewKVStoreKey(types.StoreKey)
	tkeyParams := sdk.NewTransientStoreKey(params.TStoreKey)

	ms := store.NewCommitMultiStore(db)
	ms.MountStoreWithDB(keyAcc, sdk.StoreTypeIAVL, db)
	ms.MountStoreWithDB(keyParams, sdk.StoreTypeIAVL, db)
	ms.MountStoreWithDB(keyAuthorization, sdk.StoreTypeIAVL, db)
	ms.MountStoreWithDB(tkeyParams, sdk.StoreTypeTransient, db)

	ms.LoadLatestVersion()

	ctx := sdk.NewContext(ms, abci.Header{Time: time.Unix(0, 0)}, false, log.NewNopLogger())
	cdc = makeTestCodec()

	blacklistedAddrs := make(map[string]bool)

	paramsKeeper := params.NewKeeper(cdc, keyParams, tkeyParams)
	authKeeper := auth.NewAccountKeeper(cdc, keyAcc, paramsKeeper.Subspace(auth.DefaultParamspace), auth.ProtoBaseAccount)
	bankKeeper := bank.NewBaseKeeper(authKeeper, paramsKeeper.Subspace(bank.DefaultParamspace), blacklistedAddrs)
	bankKeeper.SetSendEnabled(ctx, true)

	router := baseapp.NewRouter()
	router.AddRoute("bank", bank.NewHandler(bankKeeper))

	authorizationKeeper := NewKeeper(keyAuthorization, cdc, router)

	authKeeper.SetParams(ctx, auth.DefaultParams())

	return ctx, authKeeper, paramsKeeper, bankKeeper, authorizationKeeper, router
}

var (
	granteePub    = ed25519.GenPrivKey().PubKey()
	granterPub    = ed25519.GenPrivKey().PubKey()
	recepientPub  = ed25519.GenPrivKey().PubKey()
	granteeAddr   = sdk.AccAddress(granteePub.Address())
	granterAddr   = sdk.AccAddress(granterPub.Address())
	recepientAddr = sdk.AccAddress(recepientPub.Address())
)
