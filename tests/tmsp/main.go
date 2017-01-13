package main

import (
	"fmt"

	"github.com/tendermint/basecoin/app"
	"github.com/tendermint/basecoin/tests"
	"github.com/tendermint/basecoin/types"
	. "github.com/tendermint/go-common"
	"github.com/tendermint/go-wire"
	eyescli "github.com/tendermint/merkleeyes/client"
)

func main() {
	testSendTx()
	testSequence()
}

func testSendTx() {
	eyesCli := eyescli.NewLocalClient()
	chainID := "test_chain_id"
	bcApp := app.NewBasecoin(eyesCli)
	bcApp.SetOption("base/chainID", chainID)
	fmt.Println(bcApp.Info())

	test1PrivAcc := tests.PrivAccountFromSecret("test1")
	test2PrivAcc := tests.PrivAccountFromSecret("test2")

	// Seed Basecoin with account
	test1Acc := test1PrivAcc.Account
	test1Acc.Balance = types.Coins{{"", 1000}}
	fmt.Println(bcApp.SetOption("base/account", string(wire.JSONBytes(test1Acc))))

	res := bcApp.Commit()
	if res.IsErr() {
		Exit(Fmt("Failed Commit: %v", res.Error()))
	}

	// Construct a SendTx signature
	tx := &types.SendTx{
		Fee: 0,
		Gas: 0,
		Inputs: []types.TxInput{
			tests.MakeInput(test1PrivAcc.Account.PubKey, types.Coins{{"", 1}}, 1),
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
	fmt.Printf("Sign bytes: %X\n", signBytes)
	sig := test1PrivAcc.PrivKey.Sign(signBytes)
	tx.Inputs[0].Signature = sig
	//fmt.Println("tx:", tx)
	fmt.Printf("Signed TX bytes: %X\n", wire.BinaryBytes(struct{ types.Tx }{tx}))

	// Write request
	txBytes := wire.BinaryBytes(struct{ types.Tx }{tx})
	res = bcApp.AppendTx(txBytes)
	fmt.Println(res)
	if res.IsErr() {
		Exit(Fmt("Failed: %v", res.Error()))
	}
}

func testSequence() {
	eyesCli := eyescli.NewLocalClient()
	chainID := "test_chain_id"
	bcApp := app.NewBasecoin(eyesCli)
	bcApp.SetOption("base/chainID", chainID)
	fmt.Println(bcApp.Info())

	// Get the test account
	test1PrivAcc := tests.PrivAccountFromSecret("test1")
	test1Acc := test1PrivAcc.Account
	test1Acc.Balance = types.Coins{{"", 1 << 53}}
	fmt.Println(bcApp.SetOption("base/account", string(wire.JSONBytes(test1Acc))))

	res := bcApp.Commit()
	if res.IsErr() {
		Exit(Fmt("Failed Commit: %v", res.Error()))
	}

	sequence := int(1)
	// Make a bunch of PrivAccounts
	privAccounts := tests.RandAccounts(1000, 1000000, 0)
	privAccountSequences := make(map[string]int)
	// Send coins to each account

	for i := 0; i < len(privAccounts); i++ {
		privAccount := privAccounts[i]

		tx := &types.SendTx{
			Fee: 2,
			Gas: 2,
			Inputs: []types.TxInput{
				tests.MakeInput(test1Acc.PubKey, types.Coins{{"", 1000002}}, sequence),
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
		sig := test1PrivAcc.PrivKey.Sign(signBytes)
		tx.Inputs[0].Signature = sig
		// fmt.Printf("ADDR: %X -> %X\n", tx.Inputs[0].Address, tx.Outputs[0].Address)

		// Write request
		txBytes := wire.BinaryBytes(struct{ types.Tx }{tx})
		res := bcApp.AppendTx(txBytes)
		if res.IsErr() {
			Exit("AppendTx error: " + res.Error())
		}

	}

	fmt.Println("-------------------- RANDOM SENDS --------------------")

	res = bcApp.Commit()
	if res.IsErr() {
		Exit(Fmt("Failed Commit: %v", res.Error()))
	}

	// Now send coins between these accounts
	for i := 0; i < 10000; i++ {
		randA := RandInt() % len(privAccounts)
		randB := RandInt() % len(privAccounts)
		if randA == randB {
			continue
		}

		privAccountA := privAccounts[randA]
		privAccountASequence := privAccountSequences[privAccountA.Account.PubKey.KeyString()]
		privAccountSequences[privAccountA.Account.PubKey.KeyString()] = privAccountASequence + 1
		privAccountB := privAccounts[randB]

		tx := &types.SendTx{
			Fee: 2,
			Gas: 2,
			Inputs: []types.TxInput{
				tests.MakeInput(privAccountA.Account.PubKey, types.Coins{{"", 3}}, privAccountASequence+1),
			},
			Outputs: []types.TxOutput{
				types.TxOutput{
					Address: privAccountB.Account.PubKey.Address(),
					Coins:   types.Coins{{"", 1}},
				},
			},
		}

		// Sign request
		signBytes := tx.SignBytes(chainID)
		sig := privAccountA.PrivKey.Sign(signBytes)
		tx.Inputs[0].Signature = sig
		// fmt.Printf("ADDR: %X -> %X\n", tx.Inputs[0].Address, tx.Outputs[0].Address)

		// Write request
		txBytes := wire.BinaryBytes(struct{ types.Tx }{tx})
		res := bcApp.AppendTx(txBytes)
		if res.IsErr() {
			Exit("AppendTx error: " + res.Error())
		}
	}
}
