package mock

import (
	"os"

	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/crypto"
	dbm "github.com/tendermint/tendermint/libs/db"
	"github.com/tendermint/tendermint/libs/log"

	bam "github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/wire"
	"github.com/cosmos/cosmos-sdk/x/auth"
)

// Extended ABCI application
type App struct {
	*bam.BaseApp
	Cdc        *wire.Codec // public since the codec is passed into the module anyways.
	KeyMain    *sdk.KVStoreKey
	KeyAccount *sdk.KVStoreKey

	// TODO: Abstract this out from not needing to be auth specifically
	AccountMapper       auth.AccountMapper
	FeeCollectionKeeper auth.FeeCollectionKeeper

	GenesisAccounts []auth.Account
}

// partially construct a new app on the memstore for module and genesis testing
func NewApp() *App {
	logger := log.NewTMLogger(log.NewSyncWriter(os.Stdout)).With("module", "sdk/app")
	db := dbm.NewMemDB()

	// create the cdc with some standard codecs
	cdc := wire.NewCodec()
	sdk.RegisterWire(cdc)
	wire.RegisterCrypto(cdc)
	auth.RegisterWire(cdc)

	// create your application object
	app := &App{
		BaseApp:    bam.NewBaseApp("mock", cdc, logger, db),
		Cdc:        cdc,
		KeyMain:    sdk.NewKVStoreKey("main"),
		KeyAccount: sdk.NewKVStoreKey("acc"),
	}

	// define the accountMapper
	app.AccountMapper = auth.NewAccountMapper(
		app.Cdc,
		app.KeyAccount,      // target store
		&auth.BaseAccount{}, // prototype
	)

	// initialize the app, the chainers and blockers can be overwritten before calling complete setup
	app.SetInitChainer(app.InitChainer)

	app.SetAnteHandler(auth.NewAnteHandler(app.AccountMapper, app.FeeCollectionKeeper))

	return app
}

// complete the application setup after the routes have been registered
func (app *App) CompleteSetup(newKeys []*sdk.KVStoreKey) error {
	newKeys = append(newKeys, app.KeyMain)
	newKeys = append(newKeys, app.KeyAccount)
	app.MountStoresIAVL(newKeys...)
	err := app.LoadLatestVersion(app.KeyMain)
	return err
}

// custom logic for initialization
func (app *App) InitChainer(ctx sdk.Context, _ abci.RequestInitChain) abci.ResponseInitChain {

	// load the accounts
	for _, genacc := range app.GenesisAccounts {
		acc := app.AccountMapper.NewAccountWithAddress(ctx, genacc.GetAddress())
		err := acc.SetCoins(genacc.GetCoins())
		if err != nil {
			// TODO: Handle with #870
			panic(err)
		}
		app.AccountMapper.SetAccount(ctx, acc)
	}

	return abci.ResponseInitChain{}
}

// Generate genesis accounts loaded with coins, and returns their addresses, pubkeys, and privkeys
func CreateGenAccounts(numAccs int64, genCoins sdk.Coins) (genAccs []auth.Account, addrs []sdk.Address, pubKeys []crypto.PubKey, privKeys []crypto.PrivKey) {
	for i := int64(0); i < numAccs; i++ {
		privKey := crypto.GenPrivKeyEd25519()
		pubKey := privKey.PubKey()
		addr := sdk.Address(pubKey.Address())

		genAcc := &auth.BaseAccount{
			Address: addr,
			Coins:   genCoins,
		}

		genAccs = append(genAccs, genAcc)
		privKeys = append(privKeys, privKey)
		pubKeys = append(pubKeys, pubKey)
		addrs = append(addrs, addr)
	}

	return
}
