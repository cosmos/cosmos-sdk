package app

import (
	cmn "github.com/tendermint/tmlibs/common"
	dbm "github.com/tendermint/tmlibs/db"
	"github.com/tendermint/tmlibs/log"

	bapp "github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	app3Name = "App3"
)

func NewApp3(logger log.Logger, db dbm.DB) *bapp.BaseApp {

	cdc := NewCodec()

	// Create the base application object.
	app := bapp.NewBaseApp(app3Name, cdc, logger, db)

	// Create a key for accessing the account store.
	keyAccount := sdk.NewKVStoreKey("acc")
	keyIssuer := sdk.NewKVStoreKey("issuer")

	// TODO: accounts, ante handler

	// Register message routes.
	// Note the handler gets access to the account store.
	app.Router().
		AddRoute("bank", NewApp2Handler(keyAccount, keyIssuer))

	// Mount stores and load the latest state.
	app.MountStoresIAVL(keyAccount, keyIssuer)
	err := app.LoadLatestVersion(keyAccount)
	if err != nil {
		cmn.Exit(err.Error())
	}
	return app
}

//------------------------------------------------------------------
// StdTx

//------------------------------------------------------------------
// Account

//------------------------------------------------------------------
// Ante Handler
