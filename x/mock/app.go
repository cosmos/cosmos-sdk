package mock

import (
	"os"

	bam "github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/wire"
	"github.com/cosmos/cosmos-sdk/x/auth"
	abci "github.com/tendermint/tendermint/abci/types"
	cfg "github.com/tendermint/tendermint/config"
	"github.com/tendermint/tendermint/crypto"
	dbm "github.com/tendermint/tendermint/libs/db"
	"github.com/tendermint/tendermint/libs/log"
)

const chainID = ""

// App extends an ABCI application.
type App struct {
	*bam.BaseApp
	Cdc        *wire.Codec // Cdc is public since the codec is passed into the module anyways
	KeyMain    *sdk.KVStoreKey
	KeyAccount *sdk.KVStoreKey

	// TODO: Abstract this out from not needing to be auth specifically
	AccountMapper       auth.AccountMapper
	FeeCollectionKeeper auth.FeeCollectionKeeper

	GenesisAccounts []auth.Account
}

// NewApp partially constructs a new app on the memstore for module and genesis
// testing.
func NewApp() *App {
	logger := log.NewTMLogger(log.NewSyncWriter(os.Stdout)).With("module", "sdk/app")
	config := cfg.DefaultConfig()
	ctx := sdk.NewServerContext(config, logger)
	db := dbm.NewMemDB()

	// Create the cdc with some standard codecs
	cdc := wire.NewCodec()
	sdk.RegisterWire(cdc)
	wire.RegisterCrypto(cdc)
	auth.RegisterWire(cdc)

	// Create your application object
	app := &App{
		BaseApp:    bam.NewBaseApp("mock", cdc, ctx, db),
		Cdc:        cdc,
		KeyMain:    sdk.NewKVStoreKey("main"),
		KeyAccount: sdk.NewKVStoreKey("acc"),
	}

	// Define the accountMapper
	app.AccountMapper = auth.NewAccountMapper(
		app.Cdc,
		app.KeyAccount,
		&auth.BaseAccount{},
	)

	// Initialize the app. The chainers and blockers can be overwritten before
	// calling complete setup.
	app.SetInitChainer(app.InitChainer)
	app.SetAnteHandler(auth.NewAnteHandler(app.AccountMapper, app.FeeCollectionKeeper))

	return app
}

// CompleteSetup completes the application setup after the routes have been
// registered.
func (app *App) CompleteSetup(newKeys []*sdk.KVStoreKey) error {
	newKeys = append(newKeys, app.KeyMain)
	newKeys = append(newKeys, app.KeyAccount)

	app.MountStoresIAVL(newKeys...)
	err := app.LoadLatestVersion(app.KeyMain)

	return err
}

// InitChainer performs custom logic for initialization.
func (app *App) InitChainer(ctx sdk.Context, _ abci.RequestInitChain) abci.ResponseInitChain {
	// Load the genesis accounts
	for _, genacc := range app.GenesisAccounts {
		acc := app.AccountMapper.NewAccountWithAddress(ctx, genacc.GetAddress())
		acc.SetCoins(genacc.GetCoins())
		app.AccountMapper.SetAccount(ctx, acc)
	}

	return abci.ResponseInitChain{}
}

// CreateGenAccounts generates genesis accounts loaded with coins, and returns
// their addresses, pubkeys, and privkeys.
func CreateGenAccounts(numAccs int, genCoins sdk.Coins) (genAccs []auth.Account, addrs []sdk.Address, pubKeys []crypto.PubKey, privKeys []crypto.PrivKey) {
	for i := 0; i < numAccs; i++ {
		privKey := crypto.GenPrivKeyEd25519()
		pubKey := privKey.PubKey()
		addr := pubKey.Address()

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

// SetGenesis sets the mock app genesis accounts.
func SetGenesis(app *App, accs []auth.Account) {
	// Pass the accounts in via the application (lazy) instead of through
	// RequestInitChain.
	app.GenesisAccounts = accs

	app.InitChain(abci.RequestInitChain{})
	app.Commit()
}

// GenTx generates a signed mock transaction.
func GenTx(msgs []sdk.Msg, accnums []int64, seq []int64, priv ...crypto.PrivKey) auth.StdTx {
	// Make the transaction free
	fee := auth.StdFee{
		sdk.Coins{sdk.NewCoin("foocoin", 0)},
		100000,
	}

	sigs := make([]auth.StdSignature, len(priv))
	memo := "testmemotestmemo"

	for i, p := range priv {
		sig, err := p.Sign(auth.StdSignBytes(chainID, accnums[i], seq[i], fee, msgs, memo))
		if err != nil {
			panic(err)
		}

		sigs[i] = auth.StdSignature{
			PubKey:        p.PubKey(),
			Signature:     sig,
			AccountNumber: accnums[i],
			Sequence:      seq[i],
		}
	}

	return auth.NewStdTx(msgs, fee, sigs, memo)
}

// GenSequenceOfTxs generates a set of signed transactions of messages, such
// that they differ only by having the sequence numbers incremented between
// every transaction.
func GenSequenceOfTxs(msgs []sdk.Msg, accnums []int64, initSeqNums []int64, numToGenerate int, priv ...crypto.PrivKey) []auth.StdTx {
	txs := make([]auth.StdTx, numToGenerate, numToGenerate)
	for i := 0; i < numToGenerate; i++ {
		txs[i] = GenTx(msgs, accnums, initSeqNums, priv...)
		incrementAllSequenceNumbers(initSeqNums)
	}

	return txs
}

func incrementAllSequenceNumbers(initSeqNums []int64) {
	for i := 0; i < len(initSeqNums); i++ {
		initSeqNums[i]++
	}
}
