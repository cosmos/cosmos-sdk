package app

import (
	"encoding/json"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/examples/basecoin/types"
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

// define the custom logic for basecoin initialization
func (app *BasecoinApp) initBaseAppInitStater() {
	accountMapper := app.accountMapper
	ctxCheckTx := app.BaseApp.NewContext(true, nil)
	ctxDeliverTx := app.BaseApp.NewContext(false, nil)

	app.BaseApp.SetInitStater(func(stateJSON []byte) sdk.Error {

		var accs []*types.AppAccount

		err := json.Unmarshal(stateJSON, &accs)
		if err != nil {
			return sdk.ErrGenesisParse("").TraceCause(err, "")
		}

		for _, acc := range accs {
			accountMapper.SetAccount(ctxCheckTx, acc)
			accountMapper.SetAccount(ctxDeliverTx, acc)
		}
		return nil
	})
}
