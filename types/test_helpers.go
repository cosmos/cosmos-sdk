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
		PrivKeyS: crypto.WrapPrivKey(privKey),
		Account: Account{
			PubKey: crypto.WrapPubKey(privKey.PubKey()),
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
		pubKey := crypto.WrapPubKey(privKey.PubKey())
		privAccs[i] = PrivAccount{
			PrivKeyS: crypto.WrapPrivKey(privKey),
			Account: Account{
				PubKey:  pubKey,
				Balance: Coins{Coin{"", balance}},
			},
		}
	}

	return privAccs
}

/////////////////////////////////////////////////////////////////

//func MakeAccs(secrets ...string) (accs []PrivAccount) {
//	for _, secret := range secrets {
//		privAcc := PrivAccountFromSecret(secret)
//		privAcc.Account.Balance = Coins{{"mycoin", 7}}
//		accs = append(accs, privAcc)
//	}
//	return
//}

func MakeAcc(secret string) PrivAccount {
	privAcc := PrivAccountFromSecret(secret)
	privAcc.Account.Balance = Coins{{"mycoin", 7}}
	return privAcc
}

func Accs2TxInputs(seq int, accs ...PrivAccount) []TxInput {
	var txs []TxInput
	for _, acc := range accs {
		tx := NewTxInput(
			acc.Account.PubKey,
			Coins{{"mycoin", 5}},
			seq)
		txs = append(txs, tx)
	}
	return txs
}

//turn a list of accounts into basic list of transaction outputs
func Accs2TxOutputs(accs ...PrivAccount) []TxOutput {
	var txs []TxOutput
	for _, acc := range accs {
		tx := TxOutput{
			acc.Account.PubKey.Address(),
			Coins{{"mycoin", 4}}}
		txs = append(txs, tx)
	}
	return txs
}

func GetTx(seq int, accOut PrivAccount, accsIn ...PrivAccount) *SendTx {
	txs := &SendTx{
		Gas:     0,
		Fee:     Coin{"mycoin", 1},
		Inputs:  Accs2TxInputs(seq, accsIn...),
		Outputs: Accs2TxOutputs(accOut),
	}

	return txs
}

func SignTx(chainID string, tx *SendTx, accs ...PrivAccount) {
	signBytes := tx.SignBytes(chainID)
	for i, _ := range tx.Inputs {
		tx.Inputs[i].Signature = crypto.SignatureS{accs[i].Sign(signBytes)}
	}
}
