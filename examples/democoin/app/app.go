package app

import (
	"encoding/json"

	abci "github.com/tendermint/abci/types"
	oldwire "github.com/tendermint/go-wire"
	cmn "github.com/tendermint/tmlibs/common"
	dbm "github.com/tendermint/tmlibs/db"
	"github.com/tendermint/tmlibs/log"

	bam "github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/wire"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/bank"
	"github.com/cosmos/cosmos-sdk/x/ibc"
	"github.com/cosmos/cosmos-sdk/x/simplestake"

	"github.com/cosmos/cosmos-sdk/examples/democoin/types"
	"github.com/cosmos/cosmos-sdk/examples/democoin/x/cool"
	"github.com/cosmos/cosmos-sdk/examples/democoin/x/pow"
	"github.com/cosmos/cosmos-sdk/examples/democoin/x/sketchy"
)

const (
	appName = "DemocoinApp"
)

// Extended ABCI application
type DemocoinApp struct {
	*bam.BaseApp
	cdc *wire.Codec

	// keys to access the substores
	capKeyMainStore    *sdk.KVStoreKey
	capKeyAccountStore *sdk.KVStoreKey
	capKeyPowStore     *sdk.KVStoreKey
	capKeyIBCStore     *sdk.KVStoreKey
	capKeyStakingStore *sdk.KVStoreKey

	// Manage getting and setting accounts
	accountMapper sdk.AccountMapper
}

func NewDemocoinApp(logger log.Logger, dbs map[string]dbm.DB) *DemocoinApp {
	// create your application object
	var app = &DemocoinApp{
		BaseApp:            bam.NewBaseApp(appName, logger, dbs["main"]),
		cdc:                MakeCodec(),
		capKeyMainStore:    sdk.NewKVStoreKey("main"),
		capKeyAccountStore: sdk.NewKVStoreKey("acc"),
		capKeyPowStore:     sdk.NewKVStoreKey("pow"),
		capKeyIBCStore:     sdk.NewKVStoreKey("ibc"),
		capKeyStakingStore: sdk.NewKVStoreKey("stake"),
	}

	// define the accountMapper
	app.accountMapper = auth.NewAccountMapperSealed(
		app.capKeyMainStore, // target store
		&types.AppAccount{}, // prototype
	)

	// add handlers
	coinKeeper := bank.NewCoinKeeper(app.accountMapper)
	coolKeeper := cool.NewKeeper(app.capKeyMainStore, coinKeeper)
	powKeeper := pow.NewKeeper(app.capKeyPowStore, pow.NewPowConfig("pow", int64(1)), coinKeeper)
	ibcMapper := ibc.NewIBCMapper(app.cdc, app.capKeyIBCStore)
	stakeKeeper := simplestake.NewKeeper(app.capKeyStakingStore, coinKeeper)
	app.Router().
		AddRoute("bank", bank.NewHandler(coinKeeper)).
		AddRoute("cool", cool.NewHandler(coolKeeper)).
		AddRoute("pow", powKeeper.Handler).
		AddRoute("sketchy", sketchy.NewHandler()).
		AddRoute("ibc", ibc.NewHandler(ibcMapper, coinKeeper)).
		AddRoute("simplestake", simplestake.NewHandler(stakeKeeper))

	// initialize BaseApp
	app.SetTxDecoder(app.txDecoder)
	app.SetInitChainer(app.initChainerFn(coolKeeper, powKeeper))
	app.MountStoreWithDB(app.capKeyMainStore, sdk.StoreTypeIAVL, dbs["main"])
	app.MountStoreWithDB(app.capKeyAccountStore, sdk.StoreTypeIAVL, dbs["acc"])
	app.MountStoreWithDB(app.capKeyPowStore, sdk.StoreTypeIAVL, dbs["pow"])
	app.MountStoreWithDB(app.capKeyIBCStore, sdk.StoreTypeIAVL, dbs["ibc"])
	app.MountStoreWithDB(app.capKeyStakingStore, sdk.StoreTypeIAVL, dbs["staking"])
	// NOTE: Broken until #532 lands
	//app.MountStoresIAVL(app.capKeyMainStore, app.capKeyIBCStore, app.capKeyStakingStore)
	app.SetAnteHandler(auth.NewAnteHandler(app.accountMapper))
	err := app.LoadLatestVersion(app.capKeyMainStore)
	if err != nil {
		cmn.Exit(err.Error())
	}

	return app
}

// custom tx codec
// TODO: use new go-wire
func MakeCodec() *wire.Codec {
	const msgTypeSend = 0x1
	const msgTypeIssue = 0x2
	const msgTypeQuiz = 0x3
	const msgTypeSetTrend = 0x4
	const msgTypeMine = 0x5
	const msgTypeIBCTransferMsg = 0x6
	const msgTypeIBCReceiveMsg = 0x7
	const msgTypeBondMsg = 0x8
	const msgTypeUnbondMsg = 0x9
	var _ = oldwire.RegisterInterface(
		struct{ sdk.Msg }{},
		oldwire.ConcreteType{bank.SendMsg{}, msgTypeSend},
		oldwire.ConcreteType{bank.IssueMsg{}, msgTypeIssue},
		oldwire.ConcreteType{cool.QuizMsg{}, msgTypeQuiz},
		oldwire.ConcreteType{cool.SetTrendMsg{}, msgTypeSetTrend},
		oldwire.ConcreteType{pow.MineMsg{}, msgTypeMine},
		oldwire.ConcreteType{ibc.IBCTransferMsg{}, msgTypeIBCTransferMsg},
		oldwire.ConcreteType{ibc.IBCReceiveMsg{}, msgTypeIBCReceiveMsg},
		oldwire.ConcreteType{simplestake.BondMsg{}, msgTypeBondMsg},
		oldwire.ConcreteType{simplestake.UnbondMsg{}, msgTypeUnbondMsg},
	)

	const accTypeApp = 0x1
	var _ = oldwire.RegisterInterface(
		struct{ sdk.Account }{},
		oldwire.ConcreteType{&types.AppAccount{}, accTypeApp},
	)
	cdc := wire.NewCodec()

	// cdc.RegisterInterface((*sdk.Msg)(nil), nil)
	// bank.RegisterWire(cdc)   // Register bank.[SendMsg,IssueMsg] types.
	// crypto.RegisterWire(cdc) // Register crypto.[PubKey,PrivKey,Signature] types.
	// ibc.RegisterWire(cdc) // Register ibc.[IBCTransferMsg, IBCReceiveMsg] types.
	return cdc
}

// custom logic for transaction decoding
func (app *DemocoinApp) txDecoder(txBytes []byte) (sdk.Tx, sdk.Error) {
	var tx = sdk.StdTx{}

	if len(txBytes) == 0 {
		return nil, sdk.ErrTxDecode("txBytes are empty")
	}

	// StdTx.Msg is an interface. The concrete types
	// are registered by MakeTxCodec in bank.RegisterWire.
	err := app.cdc.UnmarshalBinary(txBytes, &tx)
	if err != nil {
		return nil, sdk.ErrTxDecode("").TraceCause(err, "")
	}
	return tx, nil
}

// custom logic for democoin initialization
func (app *DemocoinApp) initChainerFn(coolKeeper cool.Keeper, powKeeper pow.Keeper) sdk.InitChainer {
	return func(ctx sdk.Context, req abci.RequestInitChain) abci.ResponseInitChain {
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

		// Application specific genesis handling
		err = coolKeeper.InitGenesis(ctx, genesisState.CoolGenesis)
		if err != nil {
			panic(err) // TODO https://github.com/cosmos/cosmos-sdk/issues/468
			//	return sdk.ErrGenesisParse("").TraceCause(err, "")
		}

		err = powKeeper.InitGenesis(ctx, genesisState.PowGenesis)
		if err != nil {
			panic(err) // TODO https://github.com/cosmos/cosmos-sdk/issues/468
			//	return sdk.ErrGenesisParse("").TraceCause(err, "")
		}

		return abci.ResponseInitChain{}
	}
}
