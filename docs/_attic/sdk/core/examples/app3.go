package app

import (
	cryptoAmino "github.com/tendermint/tendermint/crypto/encoding/amino"
	cmn "github.com/tendermint/tendermint/libs/common"
	dbm "github.com/tendermint/tendermint/libs/db"
	"github.com/tendermint/tendermint/libs/log"

	bapp "github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/bank"
)

const (
	app3Name = "App3"
)

func NewApp3(logger log.Logger, db dbm.DB) *bapp.BaseApp {

	// Create the codec with registered Msg types
	cdc := UpdatedCodec()

	// Create the base application object.
	app := bapp.NewBaseApp(app3Name, logger, db, auth.DefaultTxDecoder(cdc))

	// Create a key for accessing the account store.
	keyAccount := sdk.NewKVStoreKey(auth.StoreKey)
	keyFees := sdk.NewKVStoreKey(auth.FeeStoreKey) // TODO

	// Set various mappers/keepers to interact easily with underlying stores
	accountKeeper := auth.NewAccountKeeper(cdc, keyAccount, auth.ProtoBaseAccount)
	bankKeeper := bank.NewBaseKeeper(accountKeeper)
	feeKeeper := auth.NewFeeCollectionKeeper(cdc, keyFees)

	app.SetAnteHandler(auth.NewAnteHandler(accountKeeper, feeKeeper))

	// Register message routes.
	// Note the handler gets access to
	app.Router().
		AddRoute("bank", bank.NewHandler(bankKeeper))

	// Mount stores and load the latest state.
	app.MountStoresIAVL(keyAccount, keyFees)
	err := app.LoadLatestVersion(keyAccount)
	if err != nil {
		cmn.Exit(err.Error())
	}
	return app
}

// Update codec from app2 to register imported modules
func UpdatedCodec() *codec.Codec {
	cdc := codec.New()
	cdc.RegisterInterface((*sdk.Msg)(nil), nil)
	cdc.RegisterConcrete(MsgSend{}, "example/MsgSend", nil)
	cdc.RegisterConcrete(MsgIssue{}, "example/MsgIssue", nil)
	auth.RegisterCodec(cdc)
	cryptoAmino.RegisterAmino(cdc)
	return cdc
}
