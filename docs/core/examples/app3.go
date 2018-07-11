package app

import (
	cmn "github.com/tendermint/tendermint/libs/common"
	dbm "github.com/tendermint/tendermint/libs/db"
	"github.com/tendermint/tendermint/libs/log"

	bapp "github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/bank"
)

const (
	app3Name = "App3"
)

func NewApp3(logger log.Logger, db dbm.DB) *bapp.BaseApp {

	// Create the codec with registered Msg types
	cdc := NewCodec()

	// Create the base application object.
	app := bapp.NewBaseApp(app3Name, cdc, logger, db)

	// Create a key for accessing the account store.
	keyAccount := sdk.NewKVStoreKey("acc")
	keyFees := sdk.NewKVStoreKey("fee") // TODO

	// Set various mappers/keepers to interact easily with underlying stores
	accountMapper := auth.NewAccountMapper(cdc, keyAccount, auth.ProtoBaseAccount)
	coinKeeper := bank.NewKeeper(accountMapper)
	feeKeeper := auth.NewFeeCollectionKeeper(cdc, keyFees)

	app.SetAnteHandler(auth.NewAnteHandler(accountMapper, feeKeeper))

	// Register message routes.
	// Note the handler gets access to
	app.Router().
		AddRoute("send", bank.NewHandler(coinKeeper))

	// Mount stores and load the latest state.
	app.MountStoresIAVL(keyAccount, keyFees)
	err := app.LoadLatestVersion(keyAccount)
	if err != nil {
		cmn.Exit(err.Error())
	}
	return app
}
