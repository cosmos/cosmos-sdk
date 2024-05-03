package simulation

import (
	"fmt"
	"math/rand"

	"golang.org/x/exp/maps"

	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Account contains a privkey, pubkey, address tuple
// eventually more useful data can be placed in here.
// (e.g. number of coins)
type Account struct {
	PrivKey cryptotypes.PrivKey
	PubKey  cryptotypes.PubKey
	Address sdk.AccAddress
	ConsKey cryptotypes.PrivKey
}

// Equals returns true if two accounts are equal
func (acc Account) Equals(acc2 Account) bool {
	return acc.Address.Equals(acc2.Address)
}

// RandomAcc picks and returns a random account and its index from an array.
func RandomAcc(r *rand.Rand, accs []Account) (Account, int) {
	idx := r.Intn(len(accs))
	return accs[idx], idx
}

// RandomAccounts generates n random accounts without duplicates.
func RandomAccounts(r *rand.Rand, n int) []Account {
	accs := make(map[string]Account, n)
	for len(accs) < n {
		// don't need that much entropy for simulation
		privkeySeed := make([]byte, 15)

		if _, err := r.Read(privkeySeed); err != nil {
			panic(err)
		}
		privKey := secp256k1.GenPrivKeyFromSecret(privkeySeed)
		addr := sdk.AccAddress(privKey.PubKey().Address())
		accs[string(addr.Bytes())] = Account{
			PrivKey: privKey,
			PubKey:  privKey.PubKey(),
			Address: addr,
			ConsKey: ed25519.GenPrivKeyFromSecret(privkeySeed),
		}
	}
	return maps.Values(accs)
}

// FindAccount iterates over all the simulation accounts to find the one that matches
// the given address
func FindAccount(accs []Account, address sdk.Address) (Account, bool) {
	for _, acc := range accs {
		if acc.Address.Equals(address) {
			return acc, true
		}
	}

	return Account{}, false
}

// RandomFees returns a random fee by selecting a random coin denomination and
// amount from the account's available balance. If the user doesn't have enough
// funds for paying fees, it returns empty coins.
func RandomFees(r *rand.Rand, spendableCoins sdk.Coins) (sdk.Coins, error) {
	if spendableCoins.Empty() {
		return nil, nil
	}

	perm := r.Perm(len(spendableCoins))
	var randCoin sdk.Coin
	for _, index := range perm {
		randCoin = spendableCoins[index]
		if !randCoin.Amount.IsZero() {
			break
		}
	}

	if randCoin.Amount.IsZero() {
		return nil, fmt.Errorf("no coins found for random fees")
	}

	amt, err := RandPositiveInt(r, randCoin.Amount)
	if err != nil {
		return nil, err
	}

	// Create a random fee and verify the fees are within the account's spendable
	// balance.
	fees := sdk.NewCoins(sdk.NewCoin(randCoin.Denom, amt))

	return fees, nil
}
