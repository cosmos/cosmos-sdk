package app

import (
	"encoding/json"

	abci "github.com/tendermint/abci/types"
	cmn "github.com/tendermint/tmlibs/common"
	dbm "github.com/tendermint/tmlibs/db"
	"github.com/tendermint/tmlibs/log"

	bam "github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/wire"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/bank"
	"github.com/cosmos/cosmos-sdk/x/ibc"
	"github.com/cosmos/cosmos-sdk/x/stake"
)

const (
	appName = "GaiaApp"
)

// Extended ABCI application
type GaiaApp struct {
	*bam.BaseApp
	cdc *wire.Codec

	// keys to access the substores
	capKeyMainStore    *sdk.KVStoreKey
	capKeyAccountStore *sdk.KVStoreKey
	capKeyIBCStore     *sdk.KVStoreKey
	capKeyStakeStore   *sdk.KVStoreKey

	// Manage getting and setting accounts
	accountMapper sdk.AccountMapper
	coinKeeper    bank.CoinKeeper
	ibcMapper     ibc.IBCMapper
	stakeKeeper   stake.Keeper

	// Handle fees
	feeHandler sdk.FeeHandler
}

func NewGaiaApp(logger log.Logger, db dbm.DB) *GaiaApp {
	// create your application object
	var app = &GaiaApp{
		BaseApp:            bam.NewBaseApp(appName, logger, db),
		cdc:                MakeCodec(),
		capKeyMainStore:    sdk.NewKVStoreKey("main"),
		capKeyAccountStore: sdk.NewKVStoreKey("acc"),
		capKeyIBCStore:     sdk.NewKVStoreKey("ibc"),
		capKeyStakeStore:   sdk.NewKVStoreKey("stake"),
	}

	// define the accountMapper
	app.accountMapper = auth.NewAccountMapper(
		app.cdc,
		app.capKeyMainStore, // target store
		&auth.BaseAccount{}, // prototype
	).Seal()

	// add handlers
	app.coinKeeper = bank.NewCoinKeeper(app.accountMapper)
	app.ibcMapper = ibc.NewIBCMapper(app.cdc, app.capKeyIBCStore, app.RegisterCodespace(ibc.DefaultCodespace))
	app.stakeKeeper = stake.NewKeeper(app.cdc, app.capKeyStakeStore, app.coinKeeper, app.RegisterCodespace(stake.DefaultCodespace))

	app.Router().
		AddRoute("bank", bank.NewHandler(app.coinKeeper)).
		AddRoute("ibc", ibc.NewHandler(app.ibcMapper, app.coinKeeper)).
		AddRoute("stake", stake.NewHandler(app.stakeKeeper))

	// Define the feeHandler.
	app.feeHandler = auth.BurnFeeHandler

	// initialize BaseApp
	app.SetTxDecoder(app.txDecoder)
	app.SetInitChainer(app.initChainer)
	app.SetEndBlocker(stake.NewEndBlocker(app.stakeKeeper))
	app.MountStoresIAVL(app.capKeyMainStore, app.capKeyAccountStore, app.capKeyIBCStore, app.capKeyStakeStore)
	app.SetAnteHandler(auth.NewAnteHandler(app.accountMapper, app.feeHandler))
	err := app.LoadLatestVersion(app.capKeyMainStore)
	if err != nil {
		cmn.Exit(err.Error())
	}

	return app
}

// custom tx codec
func MakeCodec() *wire.Codec {
	var cdc = wire.NewCodec()

	// Register Msgs
	cdc.RegisterInterface((*sdk.Msg)(nil), nil)

	ibc.RegisterWire(cdc)
	bank.RegisterWire(cdc)
	stake.RegisterWire(cdc)

	// Register AppAccount
	cdc.RegisterInterface((*sdk.Account)(nil), nil)
	cdc.RegisterConcrete(&auth.BaseAccount{}, "gaia/Account", nil)

	// Register crypto.
	wire.RegisterCrypto(cdc)

	return cdc
}

// custom logic for transaction decoding
func (app *GaiaApp) txDecoder(txBytes []byte) (sdk.Tx, sdk.Error) {
	var tx = sdk.StdTx{}

	if len(txBytes) == 0 {
		return nil, sdk.ErrTxDecode("txBytes are empty")
	}

	// StdTx.Msg is an interface. The concrete types
	// are registered by MakeTxCodec
	err := app.cdc.UnmarshalBinary(txBytes, &tx)
	if err != nil {
		return nil, sdk.ErrTxDecode("").Trace(err.Error())
	}
	return tx, nil
}

// custom logic for gaia initialization
func (app *GaiaApp) initChainer(ctx sdk.Context, req abci.RequestInitChain) abci.ResponseInitChain {
	stateJSON := req.AppStateBytes

	genesisState := new(GenesisState)
	err := json.Unmarshal(stateJSON, genesisState)
	if err != nil {
		panic(err) // TODO https://github.com/cosmos/cosmos-sdk/issues/468
		// return sdk.ErrGenesisParse("").TraceCause(err, "")
	}

	// load the accounts
	for _, gacc := range genesisState.Accounts {
		acc := gacc.ToAccount()
		app.accountMapper.SetAccount(ctx, acc)
	}

	// load the initial stake information
	stake.InitGenesis(ctx, app.stakeKeeper, genesisState.StakeData)

	return abci.ResponseInitChain{}
}

//__________________________________________________________

// State to Unmarshal
type GenesisState struct {
	Accounts  []GenesisAccount   `json:"accounts"`
	StakeData stake.GenesisState `json:"stake"`
}

// GenesisAccount doesn't need pubkey or sequence
type GenesisAccount struct {
	Address sdk.Address `json:"address"`
	Coins   sdk.Coins   `json:"coins"`
}

func NewGenesisAccount(acc *auth.BaseAccount) GenesisAccount {
	return GenesisAccount{
		Address: acc.Address,
		Coins:   acc.Coins,
	}
}

// convert GenesisAccount to GaiaAccount
func (ga *GenesisAccount) ToAccount() (acc *auth.BaseAccount) {
	return &auth.BaseAccount{
		Address: ga.Address,
		Coins:   ga.Coins.Sort(),
	}
}

// DefaultGenAppState expects two args: an account address
// and a coin denomination, and gives lots of coins to that address.
func DefaultGenAppState(args []string, addr sdk.Address, coinDenom string) (json.RawMessage, error) {

	accAuth := auth.NewBaseAccountWithAddress(addr)
	accAuth.Coins = sdk.Coins{{"fermion", 100000}}
	acc := NewGenesisAccount(&accAuth)
	genaccs := []GenesisAccount{acc}

	genesisState := GenesisState{
		Accounts:  genaccs,
		StakeData: stake.GetDefaultGenesisState(),
	}

	stateBytes, err := json.MarshalIndent(genesisState, "", "\t")
	if err != nil {
		return nil, err
	}

	return stateBytes, nil
}
