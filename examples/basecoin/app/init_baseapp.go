package app

import (
	"github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// initCapKeys, initBaseApp, initStores, initHandlers.
func (app *BasecoinApp) initBaseApp() {
	bapp := baseapp.NewBaseApp(appName)
	app.BaseApp = bapp
	app.router = bapp.Router()
	app.initBaseAppTxDecoder()
	app.initBaseAppInitStater()
}

func (app *BasecoinApp) initBaseAppTxDecoder() {
	cdc := makeTxCodec()
	app.BaseApp.SetTxDecoder(func(txBytes []byte) (sdk.Tx, sdk.Error) {
		var tx = sdk.StdTx{}
		// StdTx.Msg is an interface whose concrete
		// types are registered in app/msgs.go.
		err := cdc.UnmarshalBinary(txBytes, &tx)
		if err != nil {
			return nil, sdk.ErrTxParse("").TraceCause(err, "")
		}
		return tx, nil
	})
}

// used to define the custom logic for initialization
func (app *BasecoinApp) initBaseAppInitStater() {
	accountMapper := app.accountMapper
	app.BaseApp.SetInitStater(func(ctx sdk.Context, stateJSON []byte) sdk.Error {
		// TODO: parse JSON
		//accountMapper.SetAccount(ctx, ...)
	})
}
