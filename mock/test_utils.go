package mock

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
	crypto "github.com/tendermint/go-crypto"
	"math/rand"
	"testing"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/x/auth"
	abci "github.com/tendermint/abci/types"
)

func GenerateNPrivKeys(N int) (keys []crypto.PrivKey) {
	// TODO Randomize this between ed255 and secp
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

// TODO describe the use of this function
func CheckDeliver(t *testing.T, bapp *baseapp.BaseApp, tx sdk.Tx, expPass bool) sdk.Result {
	// Run a Check
	res := bapp.Check(tx)
	if expPass {
		require.Equal(t, sdk.ABCICodeOK, res.Code, res.Log)
	} else {
		require.NotEqual(t, sdk.ABCICodeOK, res.Code, res.Log)
	}

	// Simulate a Block
	bapp.BeginBlock(abci.RequestBeginBlock{})
	res = bapp.Deliver(tx)
	if expPass {
		require.Equal(t, sdk.ABCICodeOK, res.Code, res.Log)
	} else {
		require.NotEqual(t, sdk.ABCICodeOK, res.Code, res.Log)
	}
	bapp.EndBlock(abci.RequestEndBlock{})
	//bapp.Commit()
	return res
}

func CreateRandOperation(ops []Operation) func(r *rand.Rand) Operation {
	return func(r *rand.Rand) Operation {
		return ops[r.Intn(len(ops))]
	}
}
