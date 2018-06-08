package mock

import (
	"testing"

	"github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	crypto "github.com/tendermint/go-crypto"

	abci "github.com/tendermint/abci/types"
	"math/rand"
)

var chainID = "" // TODO

// Set genesis accounts with random coin values using the provided addresses and
// coin denominations
func RandomSetGenesis(r *rand.Rand, app *App, addrs []sdk.Address, denoms []string) {
	accts := make([]auth.Account, len(addrs), len(addrs))
	maxNumCoins := 2 << 50
	for i := 0; i < len(accts); i++ {
		coins := make([]sdk.Coin, len(denoms), len(denoms))
		// generate a random coin for each denomination
		for j := 0; j < len(denoms); j++ {
			coins[j] = sdk.Coin{Denom: denoms[j], Amount: int64(r.Intn(maxNumCoins))}
		}
		baseAcc := auth.NewBaseAccountWithAddress(addrs[i])
		(&baseAcc).SetCoins(coins)
		accts[i] = &baseAcc
	}
	SetGenesis(app, accts)
}

// Generate n Private Keys
func GenerateNPrivKeys(n int) (keys []crypto.PrivKey) {
	// TODO Randomize this between ed25519 and secp256k1
	keys = make([]crypto.PrivKey, n, n)
	for i := 0; i < n; i++ {
		keys[i] = crypto.GenPrivKeyEd25519()
	}
	return
}

// Generate n pairs of Private Keys, and Addresses
func GenerateNPrivKeyAddressPairs(n int) (keys []crypto.PrivKey, addrs []sdk.Address) {
	keys = make([]crypto.PrivKey, n, n)
	addrs = make([]sdk.Address, n, n)
	for i := 0; i < n; i++ {
		keys[i] = crypto.GenPrivKeyEd25519()
		addrs[i] = keys[i].PubKey().Address()
	}
	return
}

// set the mock app genesis
func SetGenesis(app *App, accs []auth.Account) {
	// pass the accounts in via the application (lazy) instead of through RequestInitChain
	app.GenesisAccounts = accs

	app.InitChain(abci.RequestInitChain{})
	app.Commit()
}

// check an account balance
func CheckBalance(t *testing.T, app *App, addr sdk.Address, exp sdk.Coins) {
	ctxCheck := app.BaseApp.NewContext(true, abci.Header{})
	res := app.AccountMapper.GetAccount(ctxCheck, addr)
	assert.Equal(t, exp, res.GetCoins())
}

// generate a signed transaction
func GenTx(msg sdk.Msg, seq []int64, priv ...crypto.PrivKey) auth.StdTx {

	// make the transaction free
	fee := auth.StdFee{
		sdk.Coins{{"foocoin", 0}},
		100000,
	}

	sigs := make([]auth.StdSignature, len(priv))
	for i, p := range priv {
		sigs[i] = auth.StdSignature{
			PubKey:    p.PubKey(),
			Signature: p.Sign(auth.StdSignBytes(chainID, seq, fee, msg)),
			Sequence:  seq[i],
		}
	}
	return auth.NewStdTx(msg, fee, sigs)
}

// simulate a block
func SignCheckDeliver(t *testing.T, app *baseapp.BaseApp, msg sdk.Msg, seq []int64, expPass bool, priv ...crypto.PrivKey) {

	// Sign the tx
	tx := GenTx(msg, seq, priv...)

	// Run a Check
	res := app.Check(tx)
	if expPass {
		require.Equal(t, sdk.ABCICodeOK, res.Code, res.Log)
	} else {
		require.NotEqual(t, sdk.ABCICodeOK, res.Code, res.Log)
	}

	// Simulate a Block
	app.BeginBlock(abci.RequestBeginBlock{})
	res = app.Deliver(tx)
	if expPass {
		require.Equal(t, sdk.ABCICodeOK, res.Code, res.Log)
	} else {
		require.NotEqual(t, sdk.ABCICodeOK, res.Code, res.Log)
	}
	app.EndBlock(abci.RequestEndBlock{})

	app.Commit()
}

// XXX the only reason we are using Sign Deliver here is because the tests
// break on check tx the second time you use SignCheckDeliver in a test because
// the checktx state has not been updated likely because commit is not being
// called!
func SignDeliver(t *testing.T, app *baseapp.BaseApp, msg sdk.Msg, seq []int64, expPass bool, priv ...crypto.PrivKey) {

	// Sign the tx
	tx := GenTx(msg, seq, priv...)

	// Simulate a Block
	app.BeginBlock(abci.RequestBeginBlock{})
	res := app.Deliver(tx)
	if expPass {
		require.Equal(t, sdk.ABCICodeOK, res.Code, res.Log)
	} else {
		require.NotEqual(t, sdk.ABCICodeOK, res.Code, res.Log)
	}
	app.EndBlock(abci.RequestEndBlock{})
}

// Get all accounts in the accountMapper
func GetAllAccounts(mapper auth.AccountMapper, ctx sdk.Context) []auth.Account {
	accounts := []auth.Account{}
	appendAccount := func(acc auth.Account) (stop bool) {
		accounts = append(accounts, acc)
		return false
	}
	mapper.IterateAccounts(ctx, appendAccount)
	return accounts
}
