package testnet

import (
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
)

// DelegatorPrivKeys is a slice of secp256k1.PrivKey.
type DelegatorPrivKeys []*secp256k1.PrivKey

// NewDelegatorPrivKeysreturns a DelegatorPrivKeys of length n,
// where each set of keys is dynamically generated.
func NewDelegatorPrivKeys(n int) DelegatorPrivKeys {
	dpk := make(DelegatorPrivKeys, n)

	for i := range dpk {
		dpk[i] = secp256k1.GenPrivKey()
	}

	return dpk
}

// BaseAccounts returns the base accounts corresponding to the delegators' public keys.
func (dpk DelegatorPrivKeys) BaseAccounts() BaseAccounts {
	ba := make(BaseAccounts, len(dpk))

	for i, pk := range dpk {
		pubKey := pk.PubKey()

		const accountNumber = 0
		const sequenceNumber = 0

		ba[i] = authtypes.NewBaseAccount(
			pubKey.Address().Bytes(), pubKey, accountNumber, sequenceNumber,
		)
	}

	return ba
}

// BaseAccounts is a slice of [*authtypes.BaseAccount].
type BaseAccounts []*authtypes.BaseAccount

// Balances creates a slice of [banktypes.Balance] for each account in ba,
// where each balance has an identical Coins value of the singleBalance argument.
func (ba BaseAccounts) Balances(singleBalance sdk.Coins) []banktypes.Balance {
	balances := make([]banktypes.Balance, len(ba))

	for i, b := range ba {
		balances[i] = banktypes.Balance{
			Address: b.GetAddress().String(),
			Coins:   singleBalance,
		}
	}

	return balances
}
