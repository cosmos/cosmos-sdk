package tmsp_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tendermint/basecoin/app"
	"github.com/tendermint/basecoin/types"
	wire "github.com/tendermint/go-wire"
	eyescli "github.com/tendermint/merkleeyes/client"
	cmn "github.com/tendermint/tmlibs/common"
	"github.com/tendermint/tmlibs/log"
)

func TestSendTx(t *testing.T) {
	eyesCli := eyescli.NewLocalClient("", 0)
	chainID := "test_chain_id"
	bcApp := app.NewBasecoin(eyesCli)
	bcApp.SetLogger(log.TestingLogger().With("module", "app"))
	bcApp.SetOption("base/chain_id", chainID)
	// t.Log(bcApp.Info())

	test1PrivAcc := types.PrivAccountFromSecret("test1")
	test2PrivAcc := types.PrivAccountFromSecret("test2")

	// Seed Basecoin with account
	test1Acc := test1PrivAcc.Account
	test1Acc.Balance = types.Coins{{"", 1000}}
	accOpt, err := json.Marshal(test1Acc)
	require.Nil(t, err)
	bcApp.SetOption("base/account", string(accOpt))

	// Construct a SendTx signature
	tx := &types.SendTx{
		Gas: 0,
		Fee: types.Coin{"", 0},
		Inputs: []types.TxInput{
			types.NewTxInput(test1PrivAcc.Account.PubKey, types.Coins{{"", 1}}, 1),
		},
		Outputs: []types.TxOutput{
			types.TxOutput{
				Address: test2PrivAcc.Account.PubKey.Address(),
				Coins:   types.Coins{{"", 1}},
			},
		},
	}

	// Sign request
	signBytes := tx.SignBytes(chainID)
	// t.Log("Sign bytes: %X\n", signBytes)
	sig := test1PrivAcc.Sign(signBytes)
	tx.Inputs[0].Signature = sig
	// t.Log("Signed TX bytes: %X\n", wire.BinaryBytes(types.TxS{tx}))

	// Write request
	txBytes := wire.BinaryBytes(types.TxS{tx})
	res := bcApp.DeliverTx(txBytes)
	// t.Log(res)
	assert.True(t, res.IsOK(), "Failed: %v", res.Error())
}

func TestSequence(t *testing.T) {
	eyesCli := eyescli.NewLocalClient("", 0)
	chainID := "test_chain_id"
	bcApp := app.NewBasecoin(eyesCli)
	bcApp.SetOption("base/chain_id", chainID)
	// t.Log(bcApp.Info())

	// Get the test account
	test1PrivAcc := types.PrivAccountFromSecret("test1")
	test1Acc := test1PrivAcc.Account
	test1Acc.Balance = types.Coins{{"", 1 << 53}}
	accOpt, err := json.Marshal(test1Acc)
	require.Nil(t, err)
	bcApp.SetOption("base/account", string(accOpt))

	sequence := int(1)
	// Make a bunch of PrivAccounts
	privAccounts := types.RandAccounts(1000, 1000000, 0)
	privAccountSequences := make(map[string]int)
	// Send coins to each account

	for i := 0; i < len(privAccounts); i++ {
		privAccount := privAccounts[i]

		tx := &types.SendTx{
			Gas: 2,
			Fee: types.Coin{"", 2},
			Inputs: []types.TxInput{
				types.NewTxInput(test1Acc.PubKey, types.Coins{{"", 1000002}}, sequence),
			},
			Outputs: []types.TxOutput{
				types.TxOutput{
					Address: privAccount.Account.PubKey.Address(),
					Coins:   types.Coins{{"", 1000000}},
				},
			},
		}
		sequence += 1

		// Sign request
		signBytes := tx.SignBytes(chainID)
		sig := test1PrivAcc.Sign(signBytes)
		tx.Inputs[0].Signature = sig
		// t.Log("ADDR: %X -> %X\n", tx.Inputs[0].Address, tx.Outputs[0].Address)

		// Write request
		txBytes := wire.BinaryBytes(struct{ types.Tx }{tx})
		res := bcApp.DeliverTx(txBytes)
		assert.True(t, res.IsOK(), "DeliverTx error: %v", res.Error())
	}

	res := bcApp.Commit()
	assert.True(t, res.IsOK(), "Failed Commit: %v", res.Error())

	t.Log("-------------------- RANDOM SENDS --------------------")

	// Now send coins between these accounts
	for i := 0; i < 10000; i++ {
		randA := cmn.RandInt() % len(privAccounts)
		randB := cmn.RandInt() % len(privAccounts)
		if randA == randB {
			continue
		}

		privAccountA := privAccounts[randA]
		privAccountASequence := privAccountSequences[privAccountA.Account.PubKey.KeyString()]
		privAccountSequences[privAccountA.Account.PubKey.KeyString()] = privAccountASequence + 1
		privAccountB := privAccounts[randB]

		tx := &types.SendTx{
			Gas: 2,
			Fee: types.Coin{"", 2},
			Inputs: []types.TxInput{
				types.NewTxInput(privAccountA.PubKey, types.Coins{{"", 3}}, privAccountASequence+1),
			},
			Outputs: []types.TxOutput{
				types.TxOutput{
					Address: privAccountB.PubKey.Address(),
					Coins:   types.Coins{{"", 1}},
				},
			},
		}

		// Sign request
		signBytes := tx.SignBytes(chainID)
		sig := privAccountA.Sign(signBytes)
		tx.Inputs[0].Signature = sig
		// t.Log("ADDR: %X -> %X\n", tx.Inputs[0].Address, tx.Outputs[0].Address)

		// Write request
		txBytes := wire.BinaryBytes(struct{ types.Tx }{tx})
		res := bcApp.DeliverTx(txBytes)
		assert.True(t, res.IsOK(), "DeliverTx error: %v", res.Error())
	}
}
