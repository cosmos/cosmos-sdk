package coin

import (
	"github.com/tendermint/basecoin/modules/auth"
	crypto "github.com/tendermint/go-crypto"
	"github.com/tendermint/go-wire/data"

	"github.com/tendermint/basecoin"
)

// AccountWithKey is a helper for tests, that includes and account
// along with the private key to access it.
type AccountWithKey struct {
	Key crypto.PrivKey
	Account
}

// NewAccountWithKey creates an account with the given balance
// and a random private key
func NewAccountWithKey(coins Coins) *AccountWithKey {
	return &AccountWithKey{
		Key:     crypto.GenPrivKeyEd25519().Wrap(),
		Account: Account{Coins: coins},
	}
}

// Address returns the public key address for this account
func (a *AccountWithKey) Address() []byte {
	return a.Key.PubKey().Address()
}

// Actor returns the basecoin actor associated with this account
func (a *AccountWithKey) Actor() basecoin.Actor {
	return auth.SigPerm(a.Key.PubKey().Address())
}

// MakeOption returns a string to use with SetOption to initialize this account
//
// This is intended for use in test cases
func (a *AccountWithKey) MakeOption() string {
	info := GenesisAccount{
		Address:  a.Address(),
		Sequence: a.Sequence,
		Balance:  a.Coins,
	}
	js, err := data.ToJSON(info)
	if err != nil {
		panic(err)
	}
	return string(js)
}
