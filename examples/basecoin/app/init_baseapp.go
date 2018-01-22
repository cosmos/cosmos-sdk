package app

import (
	"github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
)

// initCapKeys, initBaseApp, initStores, initRoutes.
func (app *BasecoinApp) initBaseApp() {
	app.BaseApp = baseapp.NewBaseApp(appName)
	app.initBaseAppTxDecoder()
	app.initBaseAppAnteHandler()
}

func (app *BasecoinApp) initBaseAppTxDecoder() {
	cdc := makeTxCodec()
	app.BaseApp.SetTxDecoder(func(txBytes []byte) (sdk.Tx, error) {
		var tx = sdk.StdTx{}
		err := cdc.UnmarshalBinary(txBytes, &tx)
		return tx, err
	})
}

func (app *BasecoinApp) initBaseAppAnteHandler() {
	var authAnteHandler = auth.NewAnteHandler(app.accountMapper)
	app.BaseApp.SetDefaultAnteHandler(authAnteHandler)
}
