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

/////////////////////////////////////////////////////////////////

func MakeAccs(secrets ...string) (accs []PrivAccount) {
	for _, secret := range secrets {
		privAcc := PrivAccountFromSecret(secret)
		privAcc.Account.Balance = Coins{{"mycoin", 7}}
		accs = append(accs, privAcc)
	}
	return
}

func Accs2TxInputs(accs []PrivAccount, seq int) []TxInput {
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
func Accs2TxOutputs(accs []PrivAccount) []TxOutput {
	var txs []TxOutput
	for _, acc := range accs {
		tx := TxOutput{
			acc.Account.PubKey.Address(),
			Coins{{"mycoin", 4}}}
		txs = append(txs, tx)
	}
	return txs
}

func GetTx(seq int, accsIn, accsOut []PrivAccount) *SendTx {
	txs := &SendTx{
		Gas:     0,
		Fee:     Coin{"mycoin", 1},
		Inputs:  Accs2TxInputs(accsIn, seq),
		Outputs: Accs2TxOutputs(accsOut),
	}

	return txs
}

func SignTx(chainID string, tx *SendTx, accs []PrivAccount) {
	signBytes := tx.SignBytes(chainID)
	for i, _ := range tx.Inputs {
		tx.Inputs[i].Signature = crypto.SignatureS{accs[i].Sign(signBytes)}
	}
}
