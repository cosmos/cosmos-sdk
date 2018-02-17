package app

import (
	"encoding/json"
	"fmt"
	"os"

	bam "github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/examples/basecoin/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/bank"
	"github.com/cosmos/cosmos-sdk/x/sketchy"

	abci "github.com/tendermint/abci/types"
	crypto "github.com/tendermint/go-crypto"
	"github.com/tendermint/go-wire"
	cmn "github.com/tendermint/tmlibs/common"
	dbm "github.com/tendermint/tmlibs/db"
	"github.com/tendermint/tmlibs/log"
)

const (
	appName = "BasecoinApp"
)

// Extended ABCI application
type BasecoinApp struct {
	*bam.BaseApp
	cdc *wire.Codec

	// keys to access the substores
	capKeyMainStore *sdk.KVStoreKey
	capKeyIBCStore  *sdk.KVStoreKey

	// Manage getting and setting accounts
	accountMapper sdk.AccountMapper
}

func NewBasecoinApp() *BasecoinApp {
	logger := log.NewTMLogger(log.NewSyncWriter(os.Stdout)).With("module", "sdk/app")
	db, err := dbm.NewGoLevelDB(appName, "data")
	if err != nil {
		// TODO: better
		fmt.Println(err)
		os.Exit(1)
	}

	// create your application object
	var app = &BasecoinApp{
		BaseApp:         bam.NewBaseApp(appName, logger, db),
		cdc:             MakeTxCodec(),
		capKeyMainStore: sdk.NewKVStoreKey("main"),
		capKeyIBCStore:  sdk.NewKVStoreKey("ibc"),
	}

	// define the accountMapper
	app.accountMapper = auth.NewAccountMapperSealed(
		app.capKeyMainStore, // target store
		&types.AppAccount{}, // prototype
	)

	// add handlers
	app.Router().AddRoute("bank", bank.NewHandler(bank.NewCoinKeeper(app.accountMapper)))
	app.Router().AddRoute("sketchy", sketchy.NewHandler())

	// initialize BaseApp
	app.SetTxDecoder()
	app.SetInitChainer()
	app.MountStoresIAVL(app.capKeyMainStore, app.capKeyIBCStore)
	app.SetAnteHandler(auth.NewAnteHandler(app.accountMapper))
	err = app.LoadLatestVersion(app.capKeyMainStore)
	if err != nil {
		cmn.Exit(err.Error())
	}

	return app
}

// custom tx codec
func MakeTxCodec() *wire.Codec {
	cdc := wire.NewCodec()
	crypto.RegisterWire(cdc) // Register crypto.[PubKey,PrivKey,Signature] types.
	bank.RegisterWire(cdc)   // Register bank.[SendMsg,IssueMsg] types.
	return cdc
}

// custom logic for transaction decoding
func (app *BasecoinApp) SetTxDecoder() {
	app.BaseApp.SetTxDecoder(func(txBytes []byte) (sdk.Tx, sdk.Error) {
		var tx = sdk.StdTx{}
		// StdTx.Msg is an interface whose concrete
		// types are registered in app/msgs.go.
		err := app.cdc.UnmarshalBinary(txBytes, &tx)
		if err != nil {
			return nil, sdk.ErrTxParse("").TraceCause(err, "")
		}
		return tx, nil
	})
}

// custom logic for basecoin initialization
func (app *BasecoinApp) SetInitChainer() {
	app.BaseApp.SetInitChainer(func(ctx sdk.Context, req abci.RequestInitChain) sdk.Error {
		stateJSON := req.AppStateBytes

		genesisState := new(types.GenesisState)
		err := json.Unmarshal(stateJSON, genesisState)
		if err != nil {
			return sdk.ErrGenesisParse("").TraceCause(err, "")
		}

		for _, gacc := range genesisState.Accounts {
			acc, err := gacc.ToAppAccount()
			if err != nil {
				return sdk.ErrGenesisParse("").TraceCause(err, "")
			}
			app.accountMapper.SetAccount(ctx, acc)
		}
		return nil
	})
}
