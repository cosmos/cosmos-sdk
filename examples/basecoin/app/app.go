package app

import (
	"encoding/json"
	"fmt"

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
)

const appName = "BasecoinApp"

// Extended ABCI application
type BasecoinApp struct {
	*bam.BaseApp
	cdc *wire.Codec

	// keys to access the substores
	capKeyMainStore *sdk.KVStoreKey
	capKeyIBCStore  *sdk.KVStoreKey
}

func NewBasecoinApp(genesisPath string) *BasecoinApp {

	var app = &BasecoinApp{
		cdc:             makeCodex(),
		capKeyMainStore: sdk.NewKVStoreKey("main"),
		capKeyIBCStore:  sdk.NewKVStoreKey("ibc"),
	}

	var accMapper = auth.NewAccountMapper(
		app.capKeyMainStore, // target store
		&types.AppAccount{}, // prototype
	)

	app.BaseApp = bam.NewBaseAppExpanded(appName, accMapper)
	app.initBaseAppTxDecoder()
	app.initBaseAppInitStater(genesisPath)

	// Add the handlers
	app.Router().AddRoute("bank", bank.NewHandler(bank.NewCoinKeeper(app.AccountMapper())))
	app.Router().AddRoute("sketchy", sketchy.NewHandler())

	// load the stores
	if err := app.LoadLatestVersion(app.capKeyMainStore); err != nil {
		cmn.Exit(err.Error())
	}

	return app
}

// Wire requires registration of interfaces & concrete types. All
// interfaces to be encoded/decoded in a Msg must be registered
// here, along with all the concrete types that implement them.
func makeTxCodec() (cdc *wire.Codec) {
	cdc = wire.NewCodec()

	// Register crypto.[PubKey,PrivKey,Signature] types.
	crypto.RegisterWire(cdc)

	// Register bank.[SendMsg,IssueMsg] types.
	bank.RegisterWire(cdc)

	return
}

func (app *BasecoinApp) initBaseAppTxDecoder() {
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

// define the custom logic for basecoin initialization
func (app *BasecoinApp) initBaseAppInitStater(genesisPath string) {

	genesisAppState, err := bam.ReadGenesisAppState(genesisPath)
	if err != nil {
		panic(fmt.Errorf("error loading genesis state: %v", err))
	}

	// set up the cache store for ctx, get ctx
	// TODO: combine with InitChain and let tendermint invoke it.
	app.BaseApp.BeginBlock(abci.RequestBeginBlock{Header: abci.Header{}})
	ctx := app.BaseApp.NewContext(false, nil) // context for DeliverTx
	err = app.BaseApp.InitStater(ctx, genesisAppState)
	if err != nil {
		cmn.Exit(fmt.Sprintf("error initializing application genesis state: %v", err))
	}

	app.BaseApp.SetInitStater(func(ctx sdk.Context, state json.RawMessage) sdk.Error {
		if state == nil {
			return nil
		}

		genesisState := new(types.GenesisState)
		err := json.Unmarshal(state, genesisState)
		if err != nil {
			return sdk.ErrGenesisParse("").TraceCause(err, "")
		}

		for _, gacc := range genesisState.Accounts {
			acc, err := gacc.ToAppAccount()
			if err != nil {
				return sdk.ErrGenesisParse("").TraceCause(err, "")
			}
			app.AccountMapper().SetAccount(ctx, acc)
		}
		return nil
	})
}
