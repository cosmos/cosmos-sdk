package simulation

import (
	"math/rand"

	"github.com/tendermint/tendermint/crypto"
	"github.com/tendermint/tendermint/crypto/ed25519"
	"github.com/tendermint/tendermint/crypto/secp256k1"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Account contains a privkey, pubkey, address tuple
// eventually more useful data can be placed in here.
// (e.g. number of coins)
type Account struct {
	PrivKey crypto.PrivKey
	PubKey  crypto.PubKey
	Address sdk.AccAddress
}

// are two accounts equal
func (acc Account) Equals(acc2 Account) bool {
	return acc.Address.Equals(acc2.Address)
}

// RandomAcc pick a random account from an array
func RandomAcc(r *rand.Rand, accs []Account) Account {
	return accs[r.Intn(
		len(accs),
	)]
}

// RandomAccounts generates n random accounts
func RandomAccounts(r *rand.Rand, n int) []Account {
	accs := make([]Account, n)
	for i := 0; i < n; i++ {
		// don't need that much entropy for simulation
		privkeySeed := make([]byte, 15)
		r.Read(privkeySeed)
		useSecp := r.Int63()%2 == 0
		if useSecp {
			accs[i].PrivKey = secp256k1.GenPrivKeySecp256k1(privkeySeed)
		} else {
			accs[i].PrivKey = ed25519.GenPrivKeyFromSecret(privkeySeed)
		}
		accs[i].PubKey = accs[i].PrivKey.PubKey()
		accs[i].Address = sdk.AccAddress(accs[i].PubKey.Address())
	}
	return accs
}
