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

	crypto "github.com/tendermint/go-crypto"
	"github.com/tendermint/go-wire"
	cmn "github.com/tendermint/tmlibs/common"
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

func NewBasecoinApp(genesisPath string) *BasecoinApp {

	// create your application object
	var app = &BasecoinApp{
		BaseApp:         bam.NewBaseApp("BasecoinApp"),
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
	app.SetInitStater(genesisPath)
	app.MountStoresIAVL(app.capKeyMainStore, app.capKeyIBCStore)
	app.SetAnteHandler(auth.NewAnteHandler(app.accountMapper))
	err := app.LoadLatestVersion(app.capKeyMainStore)
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
func (app *BasecoinApp) SetInitStater(genesisPath string) {

	// TODO remove, use state ABCI
	genesisAppState, err := bam.LoadGenesisAppState(genesisPath)
	if err != nil {
		panic(fmt.Errorf("error loading genesis state: %v", err))
	}

	app.BaseApp.SetInitStater(func(ctx sdk.Context, state json.RawMessage) sdk.Error {

		// TODO use state ABCI
		if state == nil {
			state = genesisAppState
		}
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
			app.accountMapper.SetAccount(ctx, acc)
		}
		return nil
	})
}
