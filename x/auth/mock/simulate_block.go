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

func CreateRandomGenesisAccounts(r *rand.Rand, addrs []sdk.Address, denoms []string) []auth.BaseAccount {
	accts := make([]auth.BaseAccount, len(addrs), len(addrs))
	maxNumCoins := 2 << 50
	for i := 0; i < len(accts); i++ {
		coins := make([]sdk.Coin, len(denoms), len(denoms))
		for j := 0; j < len(denoms); j++ {
			coins[j] = sdk.Coin{Denom: denoms[j], Amount: int64(r.Intn(maxNumCoins))}
		}
		accts[i] = auth.NewBaseAccountWithAddress(addrs[i])
		accts[i].SetCoins(coins)
	}
	return accts
}

// set the mock app genesis
func RandomSetGenesis(r *rand.Rand, app *App, addrs []sdk.Address, denoms []string) {
	accts := make([]auth.Account, len(addrs), len(addrs))
	maxNumCoins := 2 << 50
	for i := 0; i < len(accts); i++ {
		coins := make([]sdk.Coin, len(denoms), len(denoms))
		for j := 0; j < len(denoms); j++ {
			coins[j] = sdk.Coin{Denom: denoms[j], Amount: int64(r.Intn(maxNumCoins))}
		}
		baseAcc := auth.NewBaseAccountWithAddress(addrs[i])
		(&baseAcc).SetCoins(coins)
		accts[i] = &baseAcc
	}
	SetGenesis(app, accts)
}

func GenerateNPrivKeys(N int) (keys []crypto.PrivKey) {
	// TODO Randomize this between ed25519 and secp256k1
	keys = make([]crypto.PrivKey, N, N)
	for i := 0; i < N; i++ {
		keys[i] = crypto.GenPrivKeyEd25519()
	}
	return
}

func GenerateNPrivKeyAddressPairs(N int) (keys []crypto.PrivKey, addrs []sdk.Address) {
	keys = make([]crypto.PrivKey, N, N)
	addrs = make([]sdk.Address, N, N)
	for i := 0; i < N; i++ {
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
func GenTx(msg sdk.Msg, seq []int64, priv ...crypto.PrivKeyEd25519) auth.StdTx {

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
func SignCheckDeliver(t *testing.T, app *baseapp.BaseApp, msg sdk.Msg, seq []int64, expPass bool, priv ...crypto.PrivKeyEd25519) {

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
