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
	authexported "github.com/cosmos/cosmos-sdk/x/auth/exported"
	"github.com/cosmos/cosmos-sdk/x/bank"
	bankexported "github.com/cosmos/cosmos-sdk/x/bank/exported"
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
	KeyBank    *sdk.KVStoreKey
	KeyParams  *sdk.KVStoreKey
	TKeyParams *sdk.TransientStoreKey

	// TODO: Abstract this out from not needing to be auth specifically
	AccountKeeper auth.AccountKeeper
	BankKeeper    bank.Keeper
	ParamsKeeper  params.Keeper

	GenesisAccounts  []authexported.Account
	GenesisBalances  []bankexported.GenesisBalance
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
		KeyBank:          sdk.NewKVStoreKey(bank.StoreKey),
		KeyParams:        sdk.NewKVStoreKey("params"),
		TKeyParams:       sdk.NewTransientStoreKey("transient_params"),
		TotalCoinsSupply: sdk.NewCoins(),
	}

	app.ParamsKeeper = params.NewKeeper(app.Cdc, app.KeyParams, app.TKeyParams)
	app.AccountKeeper = auth.NewAccountKeeper(
		app.Cdc,
		app.KeyAccount,
		app.ParamsKeeper.Subspace(auth.DefaultParamspace),
		auth.ProtoBaseAccount,
	)
	app.BankKeeper = bank.NewBaseKeeper(
		app.Cdc,
		app.KeyBank,
		app.AccountKeeper,
		app.ParamsKeeper.Subspace(bank.DefaultParamspace),
		make(map[string]bool),
	)
	supplyKeeper := NewDummySupplyKeeper(app.AccountKeeper, app.BankKeeper)

	// Initialize the app. The chainers and blockers can be overwritten before
	// calling complete setup.
	app.SetInitChainer(app.InitChainer)
	app.SetAnteHandler(auth.NewAnteHandler(app.AccountKeeper, supplyKeeper, auth.DefaultSigVerificationGasConsumer))

	// not sealing for custom extension
	return app
}

// CompleteSetup completes the application setup after the routes have been
// registered.
func (app *App) CompleteSetup(newKeys ...sdk.StoreKey) error {
	newKeys = append(
		newKeys,
		app.KeyMain, app.KeyAccount, app.KeyBank, app.KeyParams, app.TKeyParams,
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
func (app *App) InitChainer(ctx sdk.Context, _ abci.RequestInitChain) abci.ResponseInitChain {
	for _, genacc := range app.GenesisAccounts {
		acc := app.AccountKeeper.NewAccountWithAddress(ctx, genacc.GetAddress())
		app.AccountKeeper.SetAccount(ctx, acc)
	}

	for _, balance := range app.GenesisBalances {
		app.BankKeeper.SetBalances(ctx, balance.GetAddress(), balance.GetCoins())
	}

	auth.InitGenesis(ctx, app.AccountKeeper, auth.DefaultGenesisState())
	bank.InitGenesis(ctx, app.BankKeeper, bank.DefaultGenesisState())

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
func CreateGenAccounts(numAccs int, genCoins sdk.Coins) (
	genAccs []authexported.Account, genBalances []bankexported.GenesisBalance,
	addrs []sdk.AccAddress, pubKeys []crypto.PubKey, privKeys []crypto.PrivKey,
) {

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
		})
		genBalances = append(genBalances, bank.Balance{
			Address: addrKeysSlice[i].Address,
			Coins:   genCoins,
		})
	}

	return
}

// SetGenesis sets the mock app genesis accounts.
func SetGenesis(app *App, accs []authexported.Account, balances []bankexported.GenesisBalance) {
	// Pass the accounts in via the application (lazy) instead of through
	// RequestInitChain.
	app.GenesisAccounts = accs
	app.GenesisBalances = balances

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
func RandomSetGenesis(r *rand.Rand, app *App, addrs []sdk.AccAddress, denoms []string) {
	accounts := make([]authexported.Account, len(addrs))
	balances := make([]bankexported.GenesisBalance, len(addrs))
	randCoinIntervals := []BigInterval{
		{sdk.NewIntWithDecimal(1, 0), sdk.NewIntWithDecimal(1, 1)},
		{sdk.NewIntWithDecimal(1, 2), sdk.NewIntWithDecimal(1, 3)},
		{sdk.NewIntWithDecimal(1, 40), sdk.NewIntWithDecimal(1, 50)},
	}

	for i := 0; i < len(accounts); i++ {
		coins := make([]sdk.Coin, len(denoms))

		// generate a random coin for each denomination
		for j := 0; j < len(denoms); j++ {
			coins[j] = sdk.Coin{Denom: denoms[j],
				Amount: RandFromBigInterval(r, randCoinIntervals),
			}
		}

		app.TotalCoinsSupply = app.TotalCoinsSupply.Add(coins...)
		baseAcc := auth.NewBaseAccountWithAddress(addrs[i])

		accounts[i] = &baseAcc
		balances[i] = bank.Balance{Address: addrs[i], Coins: coins}
	}

	app.GenesisAccounts = accounts
	app.GenesisBalances = balances
}

func createCodec() *codec.Codec {
	cdc := codec.New()
	sdk.RegisterCodec(cdc)
	codec.RegisterCrypto(cdc)
	auth.RegisterCodec(cdc)
	return cdc
}
