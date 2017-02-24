package types

// Helper functions for testing

import (
	cmn "github.com/tendermint/go-common"
	"github.com/tendermint/go-crypto"
)

// Creates a PrivAccount from secret.
// The amount is not set.
func PrivAccountFromSecret(secret string) PrivAccount {
	privKey := crypto.GenPrivKeyEd25519FromSecret([]byte(secret))
	privAccount := PrivAccount{
		PrivKeyS: crypto.PrivKeyS{privKey},
		Account: Account{
			PubKey: crypto.PubKeyS{privKey.PubKey()},
		},
	}
	return privAccount
}

// Make `num` random accounts
func RandAccounts(num int, minAmount int64, maxAmount int64) []PrivAccount {
	privAccs := make([]PrivAccount, num)
	for i := 0; i < num; i++ {

		balance := minAmount
		if maxAmount > minAmount {
			balance += cmn.RandInt64() % (maxAmount - minAmount)
		}

		privKey := crypto.GenPrivKeyEd25519()
		pubKey := crypto.PubKeyS{privKey.PubKey()}
		privAccs[i] = PrivAccount{
			PrivKeyS: crypto.PrivKeyS{privKey},
			Account: Account{
				PubKey:  pubKey,
				Balance: Coins{Coin{"", balance}},
			},
		}
	}

	return privAccs
}
