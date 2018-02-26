package app

import (
	"encoding/json"

	abci "github.com/tendermint/abci/types"
	crypto "github.com/tendermint/go-crypto"
	"github.com/tendermint/go-wire"
	cmn "github.com/tendermint/tmlibs/common"
	dbm "github.com/tendermint/tmlibs/db"
	"github.com/tendermint/tmlibs/log"

	bam "github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/bank"

	"github.com/cosmos/cosmos-sdk/examples/basecoin/types"
	"github.com/cosmos/cosmos-sdk/examples/basecoin/x/sketchy"
)

const (
	appName = "BasecoinApp"
)

var (
	cdc *wire.Codec
)

func init() {
	cdc = MakeTxCodec()
}

// Extended ABCI application
type BasecoinApp struct {
	*bam.BaseApp

	// keys to access the substores
	capKeyMainStore *sdk.KVStoreKey

	// Manage getting and setting accounts
	accountMapper sdk.AccountMapper
}

func NewBasecoinApp(logger log.Logger, db dbm.DB) *BasecoinApp {
	capKeyMainStore := sdk.NewKVStoreKey("main")

	accountMapper := auth.NewAccountMapperSealed(
		capKeyMainStore,     // target store
		&types.AppAccount{}, // prototype
	)
	coinKeeper := bank.NewCoinKeeper(accountMapper)

	ah := auth.NewAnteHandler(accountMapper)

	// create your application object
	var app = &BasecoinApp{
		BaseApp:         bam.NewBaseApp(appName, logger, db, txDecoder, ah),
		capKeyMainStore: capKeyMainStore,
		accountMapper:   accountMapper,
	}

	// add handlers
	app.Router().AddRoute("bank", bank.NewHandler(coinKeeper))
	app.Router().AddRoute("sketchy", sketchy.NewHandler())

	// initialize BaseApp
	app.SetInitChainer(app.initChainer)
	app.MountStoresIAVL(app.capKeyMainStore)
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
func txDecoder(txBytes []byte) (sdk.Tx, sdk.Error) {
	var tx = sdk.StdTx{}
	// StdTx.Msg is an interface. The concrete types
	// are registered by MakeTxCodec in bank.RegisterWire.
	err := cdc.UnmarshalBinary(txBytes, &tx)
	if err != nil {
		return nil, sdk.ErrTxParse("").TraceCause(err, "")
	}
	return tx, nil
}

// custom logic for basecoin initialization
func (app *BasecoinApp) initChainer(ctx sdk.Context, req abci.RequestInitChain) abci.ResponseInitChain {
	stateJSON := req.AppStateBytes

	genesisState := new(types.GenesisState)
	err := json.Unmarshal(stateJSON, genesisState)
	if err != nil {
		panic(err) // TODO https://github.com/cosmos/cosmos-sdk/issues/468
		// return sdk.ErrGenesisParse("").TraceCause(err, "")
	}

	for _, gacc := range genesisState.Accounts {
		acc, err := gacc.ToAppAccount()
		if err != nil {
			panic(err) // TODO https://github.com/cosmos/cosmos-sdk/issues/468
			//	return sdk.ErrGenesisParse("").TraceCause(err, "")
		}
		app.accountMapper.SetAccount(ctx, acc)
	}
	return abci.ResponseInitChain{}
}
