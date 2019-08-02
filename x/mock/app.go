package mock

import (
	"bytes"
	"fmt"
	"math/rand"
	"os"
	"sort"

	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/crypto"
	"github.com/tendermint/tendermint/crypto/ed25519"
	"github.com/tendermint/tendermint/crypto/secp256k1"
	"github.com/tendermint/tendermint/libs/log"
	dbm "github.com/tendermint/tm-db"

	bam "github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/params"
)

const chainID = ""

// App extends an ABCI application, but with most of its parameters exported.
// They are exported for convenience in creating helper functions, as object
// capabilities aren't needed for testing.
type App struct {
	*bam.BaseApp
	Cdc        *codec.Codec // Cdc is public since the codec is passed into the module anyways
	KeyMain    *sdk.KVStoreKey
	KeyAccount *sdk.KVStoreKey
	KeyParams  *sdk.KVStoreKey
	TKeyParams *sdk.TransientStoreKey

	// TODO: Abstract this out from not needing to be auth specifically
	AccountKeeper auth.AccountKeeper
	ParamsKeeper  params.Keeper

	GenesisAccounts  []auth.Account
	TotalCoinsSupply sdk.Coins
}

// NewApp partially constructs a new app on the memstore for module and genesis
// testing.
func NewApp() *App {
	logger := log.NewTMLogger(log.NewSyncWriter(os.Stdout)).With("module", "sdk/app")
	db := dbm.NewMemDB()

	// Create the cdc with some standard codecs
	cdc := createCodec()

	// Create your application object
	app := &App{
		BaseApp:          bam.NewBaseApp("mock", logger, db, auth.DefaultTxDecoder(cdc)),
		Cdc:              cdc,
		KeyMain:          sdk.NewKVStoreKey(bam.MainStoreKey),
		KeyAccount:       sdk.NewKVStoreKey(auth.StoreKey),
		KeyParams:        sdk.NewKVStoreKey("params"),
		TKeyParams:       sdk.NewTransientStoreKey("transient_params"),
		TotalCoinsSupply: sdk.NewCoins(),
	}

	// define keepers
	app.ParamsKeeper = params.NewKeeper(app.Cdc, app.KeyParams, app.TKeyParams, params.DefaultCodespace)

	app.AccountKeeper = auth.NewAccountKeeper(
		app.Cdc,
		app.KeyAccount,
		app.ParamsKeeper.Subspace(auth.DefaultParamspace),
		auth.ProtoBaseAccount,
	)

	supplyKeeper := auth.NewDummySupplyKeeper(app.AccountKeeper)

	// Initialize the app. The chainers and blockers can be overwritten before
	// calling complete setup.
	app.SetInitChainer(app.InitChainer)
	app.SetAnteHandler(auth.NewAnteHandler(app.AccountKeeper, supplyKeeper, auth.DefaultSigVerificationGasConsumer))

	// Not sealing for custom extension

	return app
}

// CompleteSetup completes the application setup after the routes have been
// registered.
func (app *App) CompleteSetup(newKeys ...sdk.StoreKey) error {
	newKeys = append(
		newKeys,
		app.KeyMain, app.KeyAccount, app.KeyParams, app.TKeyParams,
	)

	for _, key := range newKeys {
		switch key.(type) {
		case *sdk.KVStoreKey:
			app.MountStore(key, sdk.StoreTypeIAVL)
		case *sdk.TransientStoreKey:
			app.MountStore(key, sdk.StoreTypeTransient)
		default:
			return fmt.Errorf("unsupported StoreKey: %+v", key)
		}
	}

	err := app.LoadLatestVersion(app.KeyMain)

	return err
}

// InitChainer performs custom logic for initialization.
// nolint: errcheck
func (app *App) InitChainer(ctx sdk.Context, _ abci.RequestInitChain) abci.ResponseInitChain {

	// Load the genesis accounts
	for _, genacc := range app.GenesisAccounts {
		acc := app.AccountKeeper.NewAccountWithAddress(ctx, genacc.GetAddress())
		acc.SetCoins(genacc.GetCoins())
		app.AccountKeeper.SetAccount(ctx, acc)
	}

	auth.InitGenesis(ctx, app.AccountKeeper, auth.DefaultGenesisState())

	return abci.ResponseInitChain{}
}

// Type that combines an Address with the privKey and pubKey to that address
type AddrKeys struct {
	Address sdk.AccAddress
	PubKey  crypto.PubKey
	PrivKey crypto.PrivKey
}

func NewAddrKeys(address sdk.AccAddress, pubKey crypto.PubKey,
	privKey crypto.PrivKey) AddrKeys {

	return AddrKeys{
		Address: address,
		PubKey:  pubKey,
		PrivKey: privKey,
	}
}

// implement `Interface` in sort package.
type AddrKeysSlice []AddrKeys

func (b AddrKeysSlice) Len() int {
	return len(b)
}

// Sorts lexographically by Address
func (b AddrKeysSlice) Less(i, j int) bool {
	// bytes package already implements Comparable for []byte.
	switch bytes.Compare(b[i].Address.Bytes(), b[j].Address.Bytes()) {
	case -1:
		return true
	case 0, 1:
		return false
	default:
		panic("not fail-able with `bytes.Comparable` bounded [-1, 1].")
	}
}

func (b AddrKeysSlice) Swap(i, j int) {
	b[j], b[i] = b[i], b[j]
}

// CreateGenAccounts generates genesis accounts loaded with coins, and returns
// their addresses, pubkeys, and privkeys.
func CreateGenAccounts(numAccs int, genCoins sdk.Coins) (genAccs []auth.Account,
	addrs []sdk.AccAddress, pubKeys []crypto.PubKey, privKeys []crypto.PrivKey) {

	addrKeysSlice := AddrKeysSlice{}

	for i := 0; i < numAccs; i++ {
		privKey := secp256k1.GenPrivKey()
		pubKey := privKey.PubKey()
		addr := sdk.AccAddress(pubKey.Address())

		addrKeysSlice = append(addrKeysSlice, NewAddrKeys(addr, pubKey, privKey))
	}

	sort.Sort(addrKeysSlice)

	for i := range addrKeysSlice {
		addrs = append(addrs, addrKeysSlice[i].Address)
		pubKeys = append(pubKeys, addrKeysSlice[i].PubKey)
		privKeys = append(privKeys, addrKeysSlice[i].PrivKey)
		genAccs = append(genAccs, &auth.BaseAccount{
			Address: addrKeysSlice[i].Address,
			Coins:   genCoins,
		})
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
func GenTx(msgs []sdk.Msg, accnums []uint64, seq []uint64, priv ...crypto.PrivKey) auth.StdTx {
	// Make the transaction free
	fee := auth.StdFee{
		Amount: sdk.NewCoins(sdk.NewInt64Coin("foocoin", 0)),
		Gas:    100000,
	}

	sigs := make([]auth.StdSignature, len(priv))
	memo := "testmemotestmemo"

	for i, p := range priv {
		sig, err := p.Sign(auth.StdSignBytes(chainID, accnums[i], seq[i], fee, msgs, memo))
		if err != nil {
			panic(err)
		}

		sigs[i] = auth.StdSignature{
			PubKey:    p.PubKey(),
			Signature: sig,
		}
	}

	return auth.NewStdTx(msgs, fee, sigs, memo)
}

// GeneratePrivKeys generates a total n secp256k1 private keys.
func GeneratePrivKeys(n int) (keys []crypto.PrivKey) {
	// TODO: Randomize this between ed25519 and secp256k1
	keys = make([]crypto.PrivKey, n)
	for i := 0; i < n; i++ {
		keys[i] = secp256k1.GenPrivKey()
	}

	return
}

// GeneratePrivKeyAddressPairs generates a total of n private key, address
// pairs.
func GeneratePrivKeyAddressPairs(n int) (keys []crypto.PrivKey, addrs []sdk.AccAddress) {
	keys = make([]crypto.PrivKey, n)
	addrs = make([]sdk.AccAddress, n)
	for i := 0; i < n; i++ {
		if rand.Int63()%2 == 0 {
			keys[i] = secp256k1.GenPrivKey()
		} else {
			keys[i] = ed25519.GenPrivKey()
		}
		addrs[i] = sdk.AccAddress(keys[i].PubKey().Address())
	}
	return
}

// GeneratePrivKeyAddressPairsFromRand generates a total of n private key, address
// pairs using the provided randomness source.
func GeneratePrivKeyAddressPairsFromRand(rand *rand.Rand, n int) (keys []crypto.PrivKey, addrs []sdk.AccAddress) {
	keys = make([]crypto.PrivKey, n)
	addrs = make([]sdk.AccAddress, n)
	for i := 0; i < n; i++ {
		secret := make([]byte, 32)
		_, err := rand.Read(secret)
		if err != nil {
			panic("Could not read randomness")
		}
		if rand.Int63()%2 == 0 {
			keys[i] = secp256k1.GenPrivKeySecp256k1(secret)
		} else {
			keys[i] = ed25519.GenPrivKeyFromSecret(secret)
		}
		addrs[i] = sdk.AccAddress(keys[i].PubKey().Address())
	}
	return
}

// RandomSetGenesis set genesis accounts with random coin values using the
// provided addresses and coin denominations.
// nolint: errcheck
func RandomSetGenesis(r *rand.Rand, app *App, addrs []sdk.AccAddress, denoms []string) {
	accts := make([]auth.Account, len(addrs))
	randCoinIntervals := []BigInterval{
		{sdk.NewIntWithDecimal(1, 0), sdk.NewIntWithDecimal(1, 1)},
		{sdk.NewIntWithDecimal(1, 2), sdk.NewIntWithDecimal(1, 3)},
		{sdk.NewIntWithDecimal(1, 40), sdk.NewIntWithDecimal(1, 50)},
	}

	for i := 0; i < len(accts); i++ {
		coins := make([]sdk.Coin, len(denoms))

		// generate a random coin for each denomination
		for j := 0; j < len(denoms); j++ {
			coins[j] = sdk.Coin{Denom: denoms[j],
				Amount: RandFromBigInterval(r, randCoinIntervals),
			}
		}

		app.TotalCoinsSupply = app.TotalCoinsSupply.Add(coins)
		baseAcc := auth.NewBaseAccountWithAddress(addrs[i])

		(&baseAcc).SetCoins(coins)
		accts[i] = &baseAcc
	}
	app.GenesisAccounts = accts
}

// GenSequenceOfTxs generates a set of signed transactions of messages, such
// that they differ only by having the sequence numbers incremented between
// every transaction.
func GenSequenceOfTxs(msgs []sdk.Msg, accnums []uint64, initSeqNums []uint64, numToGenerate int, priv ...crypto.PrivKey) []auth.StdTx {
	txs := make([]auth.StdTx, numToGenerate)
	for i := 0; i < numToGenerate; i++ {
		txs[i] = GenTx(msgs, accnums, initSeqNums, priv...)
		incrementAllSequenceNumbers(initSeqNums)
	}

	return txs
}

func incrementAllSequenceNumbers(initSeqNums []uint64) {
	for i := 0; i < len(initSeqNums); i++ {
		initSeqNums[i]++
	}
}

func createCodec() *codec.Codec {
	cdc := codec.New()
	sdk.RegisterCodec(cdc)
	codec.RegisterCrypto(cdc)
	auth.RegisterCodec(cdc)
	return cdc
}
