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

	ibcm "github.com/cosmos/cosmos-sdk/x/ibc"
	ibc "github.com/cosmos/cosmos-sdk/x/ibc/types"
	"github.com/cosmos/cosmos-sdk/x/simplestake"

	"github.com/cosmos/cosmos-sdk/examples/basecoin/types"
)

const (
	appName = "BasecoinApp"
)

// Extended ABCI application
type BasecoinApp struct {
	*bam.BaseApp
	cdc *wire.Codec

	// keys to access the substores
	capKeyMainStore    *sdk.KVStoreKey
	capKeyAccountStore *sdk.KVStoreKey
	capKeyIBCStore     *sdk.KVStoreKey
	capKeyStakingStore *sdk.KVStoreKey

	// Manage getting and setting accounts
	accountMapper sdk.AccountMapper
}

func NewBasecoinApp(logger log.Logger, dbs map[string]dbm.DB) *BasecoinApp {
	// create your application object
	var app = &BasecoinApp{
		BaseApp:            bam.NewBaseApp(appName, logger, dbs["main"]),
		cdc:                MakeCodec(),
		capKeyMainStore:    sdk.NewKVStoreKey("main"),
		capKeyAccountStore: sdk.NewKVStoreKey("acc"),
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

	ibcKeeper := ibc.NewKeeper(app.cdc, app.capKeyIBCStore)
	ibcKeeper.Dispatcher().
		AddDispatch("bank", bank.NewIBCHandler(coinKeeper))

	stakeKeeper := simplestake.NewKeeper(app.capKeyStakingStore, coinKeeper)
	app.Router().
		AddRoute("bank", bank.NewHandler(coinKeeper, ibcKeeper.Sender())).
		AddRoute("ibc", ibcm.NewHandler(ibcKeeper)).
		AddRoute("simplestake", simplestake.NewHandler(stakeKeeper))

	// initialize BaseApp
	app.SetTxDecoder(app.txDecoder)
	app.SetInitChainer(app.initChainer)
	app.MountStoreWithDB(app.capKeyMainStore, sdk.StoreTypeIAVL, dbs["main"])
	app.MountStoreWithDB(app.capKeyAccountStore, sdk.StoreTypeIAVL, dbs["acc"])
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
	const msgTypeIBCSend = 0x2
	const msgTypeIssue = 0x3
	const msgTypeQuiz = 0x4
	const msgTypeSetTrend = 0x5
	const msgTypeReceive = 0x6
	const msgTypeBondMsg = 0x7
	const msgTypeUnbondMsg = 0x8
	var _ = oldwire.RegisterInterface(
		struct{ sdk.Msg }{},
		oldwire.ConcreteType{bank.SendMsg{}, msgTypeSend},
		oldwire.ConcreteType{bank.IBCSendMsg{}, msgTypeIBCSend},
		oldwire.ConcreteType{bank.IssueMsg{}, msgTypeIssue},
		oldwire.ConcreteType{ibcm.ReceiveMsg{}, msgTypeReceive},
		oldwire.ConcreteType{simplestake.BondMsg{}, msgTypeBondMsg},
		oldwire.ConcreteType{simplestake.UnbondMsg{}, msgTypeUnbondMsg},
	)

	const accTypeApp = 0x1
	var _ = oldwire.RegisterInterface(
		struct{ sdk.Account }{},
		oldwire.ConcreteType{&types.AppAccount{}, accTypeApp},
	)

	const payTypeSend = 0x1
	var _ = oldwire.RegisterInterface(
		struct{ ibc.Payload }{},
		oldwire.ConcreteType{bank.SendPayload{}, payTypeSend},
	)

	cdc := wire.NewCodec()

	// cdc.RegisterInterface((*sdk.Msg)(nil), nil)
	// bank.RegisterWire(cdc)   // Register bank.[SendMsg,IssueMsg] types.
	// crypto.RegisterWire(cdc) // Register crypto.[PubKey,PrivKey,Signature] types.
	// ibc.RegisterWire(cdc) // Register ibc.[IBCTransferMsg, IBCReceiveMsg] types.
	return cdc
}

// custom logic for transaction decoding
func (app *BasecoinApp) txDecoder(txBytes []byte) (sdk.Tx, sdk.Error) {
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
