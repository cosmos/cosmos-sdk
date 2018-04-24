package app

import (
	"encoding/json"
	"errors"
	"strings"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	abci "github.com/tendermint/abci/types"
	crypto "github.com/tendermint/go-crypto"
	tmtypes "github.com/tendermint/tendermint/types"
	cmn "github.com/tendermint/tmlibs/common"
	dbm "github.com/tendermint/tmlibs/db"
	"github.com/tendermint/tmlibs/log"

	bam "github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/server"
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
	keyMain    *sdk.KVStoreKey
	keyAccount *sdk.KVStoreKey
	keyIBC     *sdk.KVStoreKey
	keyStake   *sdk.KVStoreKey

	// Manage getting and setting accounts
	accountMapper sdk.AccountMapper
	coinKeeper    bank.Keeper
	ibcMapper     ibc.Mapper
	stakeKeeper   stake.Keeper

	// Handle fees
	feeHandler sdk.FeeHandler
}

func NewGaiaApp(logger log.Logger, db dbm.DB) *GaiaApp {
	// create your application object
	var app = &GaiaApp{
		BaseApp:    bam.NewBaseApp(appName, logger, db),
		cdc:        MakeCodec(),
		keyMain:    sdk.NewKVStoreKey("main"),
		keyAccount: sdk.NewKVStoreKey("acc"),
		keyIBC:     sdk.NewKVStoreKey("ibc"),
		keyStake:   sdk.NewKVStoreKey("stake"),
	}

	// define the accountMapper
	app.accountMapper = auth.NewAccountMapper(
		app.cdc,
		app.keyMain,         // target store
		&auth.BaseAccount{}, // prototype
	)

	// add handlers
	app.coinKeeper = bank.NewKeeper(app.accountMapper)
	app.ibcMapper = ibc.NewMapper(app.cdc, app.keyIBC, app.RegisterCodespace(ibc.DefaultCodespace))
	app.stakeKeeper = stake.NewKeeper(app.cdc, app.keyStake, app.coinKeeper, app.RegisterCodespace(stake.DefaultCodespace))

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
	app.MountStoresIAVL(app.keyMain, app.keyAccount, app.keyIBC, app.keyStake)
	app.SetAnteHandler(auth.NewAnteHandler(app.accountMapper, app.feeHandler))
	err := app.LoadLatestVersion(app.keyMain)
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
	err := app.cdc.UnmarshalJSON(stateJSON, genesisState)
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

//________________________________________________________________________________________

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

// convert GenesisAccount to auth.BaseAccount
func (ga *GenesisAccount) ToAccount() (acc *auth.BaseAccount) {
	return &auth.BaseAccount{
		Address: ga.Address,
		Coins:   ga.Coins.Sort(),
	}
}

var (
	flagAccounts = "accounts"
	flagOWK      = "overwrite-keys"
)

// get app init parameters for server init command
func GaiaAppInit() server.AppInit {
	fs := pflag.NewFlagSet("", pflag.ContinueOnError)
	fs.String(flagAccounts, "foobar-10fermion,10baz-true", "genesis accounts in form: name1-coins-isval:name2-coins-isval:...")
	fs.BoolP(flagOWK, "k", false, "overwrite the for the accounts created, if false and key exists init will fail")
	return server.AppInit{
		Flags:          fs,
		GenAppParams:   GaiaGenAppParams,
		AppendAppState: GaiaAppendAppState,
	}
}

// Create the core parameters for genesis initialization for gaia
// note that the pubkey input is this machines pubkey
func GaiaGenAppParams(cdc *wire.Codec, pubKey crypto.PubKey) (chainID string, validators []tmtypes.GenesisValidator, appState, cliPrint json.RawMessage, err error) {

	printMap := make(map[string]string)
	var candidates []stake.Candidate
	poolAssets := int64(0)
	chainID = cmn.Fmt("test-chain-%v", cmn.RandStr(6))

	// get genesis flag account information
	accountsStr := viper.GetString(flagAccounts)
	accounts := strings.Split(accountsStr, ":")
	genaccs := make([]GenesisAccount, len(accounts))
	for i, account := range accounts {
		p := strings.Split(account, "-")
		if len(p) != 3 {
			err = errors.New("input account has bad form, each account must be in form name-coins-isval, for example: foobar-10fermion,10baz-true")
			return
		}
		name := p[0]
		var coins sdk.Coins
		coins, err = sdk.ParseCoins(p[1])
		if err != nil {
			return
		}
		isValidator := false
		if p[2] == "true" {
			isValidator = true
		}

		var addr sdk.Address
		var secret string
		addr, secret, err = server.GenerateCoinKey()
		if err != nil {
			return
		}

		printMap["secret-"+name] = secret

		// create the genesis account
		accAuth := auth.NewBaseAccountWithAddress(addr)
		accAuth.Coins = coins
		acc := NewGenesisAccount(&accAuth)
		genaccs[i] = acc

		// add the validator
		if isValidator {

			// only use this machines pubkey the first time, all others are dummies
			var pk crypto.PubKey
			if i == 0 {
				pk = pubKey
			} else {
				pk = crypto.GenPrivKeyEd25519().PubKey()
			}

			freePower := int64(100)
			validator := tmtypes.GenesisValidator{
				PubKey: pk,
				Power:  freePower,
			}
			desc := stake.NewDescription(name, "", "", "")
			candidate := stake.NewCandidate(addr, pk, desc)
			candidate.Assets = sdk.NewRat(freePower)
			poolAssets += freePower
			validators = append(validators, validator)
			candidates = append(candidates, candidate)
		}
	}

	// create the print message
	bz, err := cdc.MarshalJSON(printMap)
	cliPrint = json.RawMessage(bz)

	stakeData := stake.GetDefaultGenesisState()
	stakeData.Candidates = candidates

	// assume everything is bonded from the get-go
	stakeData.Pool.TotalSupply = poolAssets
	stakeData.Pool.BondedShares = sdk.NewRat(poolAssets)

	genesisState := GenesisState{
		Accounts:  genaccs,
		StakeData: stakeData,
	}

	appState, err = wire.MarshalJSONIndent(cdc, genesisState)
	return
}

// append gaia app_state together, stitch the accounts together take the
// staking parameters from the first appState
func GaiaAppendAppState(cdc *wire.Codec, appState1, appState2 json.RawMessage) (appState json.RawMessage, err error) {
	var genState1, genState2 GenesisState
	err = cdc.UnmarshalJSON(appState1, &genState1)
	if err != nil {
		panic(err)
	}
	err = cdc.UnmarshalJSON(appState2, &genState2)
	if err != nil {
		panic(err)
	}
	genState1.Accounts = append(genState1.Accounts, genState2.Accounts...)

	return cdc.MarshalJSON(genState1)
}
