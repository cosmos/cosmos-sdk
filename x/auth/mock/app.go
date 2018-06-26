package mock

import (
	"testing"

	"os"

	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/abci/types"
	dbm "github.com/tendermint/tmlibs/db"
	"github.com/tendermint/tmlibs/log"

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
func (app *App) CompleteSetup(t *testing.T, newKeys []*sdk.KVStoreKey) {

	newKeys = append(newKeys, app.KeyMain)
	newKeys = append(newKeys, app.KeyAccount)
	app.MountStoresIAVL(newKeys...)
	err := app.LoadLatestVersion(app.KeyMain)
	require.NoError(t, err)
}

// custom logic for initialization
func (app *App) InitChainer(ctx sdk.Context, _ abci.RequestInitChain) abci.ResponseInitChain {

	// load the accounts
	for _, genacc := range app.GenesisAccounts {
		acc := app.AccountMapper.NewAccountWithAddress(ctx, genacc.GetAddress())
		acc.SetCoins(genacc.GetCoins())
		app.AccountMapper.SetAccount(ctx, acc)
	}

	return abci.ResponseInitChain{}
}
