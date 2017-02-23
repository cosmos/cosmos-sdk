// Functions used in testing throughout
package testutils

import (
	"github.com/tendermint/basecoin/types"
	. "github.com/tendermint/go-common"
	"github.com/tendermint/go-crypto"
)

// Creates a PrivAccount from secret.
// The amount is not set.
func PrivAccountFromSecret(secret string) types.PrivAccount {
	privKey := crypto.GenPrivKeyEd25519FromSecret([]byte(secret))
	privAccount := types.PrivAccount{
		PrivKeyS: crypto.PrivKeyS{privKey},
		Account: types.Account{
			PubKey:   crypto.PubKeyS{privKey.PubKey()},
			Sequence: 0,
		},
	}
	return privAccount
}

// Make `num` random accounts
func RandAccounts(num int, minAmount int64, maxAmount int64) []types.PrivAccount {
	privAccs := make([]types.PrivAccount, num)
	for i := 0; i < num; i++ {

		balance := minAmount
		if maxAmount > minAmount {
			balance += RandInt64() % (maxAmount - minAmount)
		}

		privKey := crypto.GenPrivKeyEd25519()
		pubKey := crypto.PubKeyS{privKey.PubKey()}
		privAccs[i] = types.PrivAccount{
			PrivKeyS: crypto.PrivKeyS{privKey},
			Account: types.Account{
				PubKey:   pubKey,
				Sequence: 0,
				Balance:  types.Coins{types.Coin{"", balance}},
			},
		}
	}

	return privAccs
}
