package main

import (
	"fmt"

	"github.com/tendermint/basecoin/app"
	"github.com/tendermint/basecoin/tests"
	"github.com/tendermint/basecoin/types"
	. "github.com/tendermint/go-common"
	"github.com/tendermint/go-wire"
	govtypes "github.com/tendermint/governmint/types"
	eyescli "github.com/tendermint/merkleeyes/client"
)

func main() {
	//testSendTx()
	//testGov()
	testSequence()
}

func testSendTx() {
	eyesCli := eyescli.NewLocalClient()
	bcApp := app.NewBasecoin(eyesCli)
	fmt.Println(bcApp.Info())

	tPriv := tests.PrivAccountFromSecret("test")
	tPriv2 := tests.PrivAccountFromSecret("test2")

	// Seed Basecoin with account
	tAcc := tPriv.Account
	tAcc.Balance = 1000
	fmt.Println(bcApp.SetOption("base/chainID", "test_chain_id"))
	fmt.Println(bcApp.SetOption("base/account", string(wire.JSONBytes(tAcc))))

	// Construct a SendTx signature
	tx := &types.SendTx{
		Inputs: []types.TxInput{
			types.TxInput{
				Address:  tPriv.Account.PubKey.Address(),
				PubKey:   tPriv.Account.PubKey, // TODO is this needed?
				Amount:   1,
				Sequence: 1,
			},
		},
		Outputs: []types.TxOutput{
			types.TxOutput{
				Address: tPriv2.Account.PubKey.Address(),
				Amount:  1,
			},
		},
	}

	// Sign request
	signBytes := tx.SignBytes("test_chain_id")
	fmt.Printf("Sign bytes: %X\n", signBytes)
	sig := tPriv.PrivKey.Sign(signBytes)
	tx.Inputs[0].Signature = sig
	//fmt.Println("tx:", tx)
	fmt.Printf("Signed TX bytes: %X\n", wire.BinaryBytes(tx))

	// Write request
	txBytes := wire.BinaryBytes(tx)
	res := bcApp.AppendTx(txBytes)
	fmt.Println(res)
	if res.IsErr() {
		Exit(Fmt("Failed: %v", res.Error()))
	}
}

func testGov() {
	eyesCli := eyescli.NewLocalClient()
	bcApp := app.NewBasecoin(eyesCli)
	fmt.Println(bcApp.Info())

	tPriv := tests.PrivAccountFromSecret("test")

	// Seed Basecoin with admin using PrivAccount
	tAcc := tPriv.Account
	adminEntity := govtypes.Entity{
		ID:     "",
		PubKey: tAcc.PubKey,
	}
	log := bcApp.SetOption("gov/admin", string(wire.JSONBytes(adminEntity)))
	if log != "Success" {
		Exit(Fmt("Failed to set option: %v", log))
	}
	// TODO test proposals or something.
}

func testSequence() {
	eyesCli := eyescli.NewLocalClient()
	bcApp := app.NewBasecoin(eyesCli)
	chainID := "test_chain_id"

	// Get the root account
	root := tests.PrivAccountFromSecret("test")
	rootAcc := root.Account
	rootAcc.Balance = 1 << 53
	fmt.Println(bcApp.SetOption("base/chainID", "test_chain_id"))
	fmt.Println(bcApp.SetOption("base/account", string(wire.JSONBytes(rootAcc))))

	sequence := int(1)
	// Make a bunch of PrivAccounts
	privAccounts := tests.RandAccounts(1000, 1000000, 0)
	privAccountSequences := make(map[string]int)

	// Send coins to each account
	for i := 0; i < len(privAccounts); i++ {
		privAccount := privAccounts[i]
		tx := &types.SendTx{
			Inputs: []types.TxInput{
				types.TxInput{
					Address:  root.Account.PubKey.Address(),
					PubKey:   root.Account.PubKey, // TODO is this needed?
					Amount:   1000002,
					Sequence: sequence,
				},
			},
			Outputs: []types.TxOutput{
				types.TxOutput{
					Address: privAccount.Account.PubKey.Address(),
					Amount:  1000000,
				},
			},
		}
		sequence += 1

		// Sign request
		signBytes := tx.SignBytes(chainID)
		sig := root.PrivKey.Sign(signBytes)
		tx.Inputs[0].Signature = sig
		//fmt.Println("tx:", tx)

		// Write request
		txBytes := wire.BinaryBytes(tx)
		res := bcApp.CheckTx(txBytes)
		if res.IsErr() {
			Exit("AppendTx error: " + res.Error())
		}
	}

	// Now send coins between these accounts
	for {
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
			Inputs: []types.TxInput{
				types.TxInput{
					Address:  privAccountA.Account.PubKey.Address(),
					PubKey:   privAccountA.Account.PubKey,
					Amount:   3,
					Sequence: privAccountASequence + 1,
				},
			},
			Outputs: []types.TxOutput{
				types.TxOutput{
					Address: privAccountB.Account.PubKey.Address(),
					Amount:  1,
				},
			},
		}

		// Sign request
		signBytes := tx.SignBytes(chainID)
		sig := privAccountA.PrivKey.Sign(signBytes)
		tx.Inputs[0].Signature = sig
		//fmt.Println("tx:", tx)

		// Write request
		txBytes := wire.BinaryBytes(tx)
		res := bcApp.AppendTx(txBytes)
		if res.IsErr() {
			Exit("AppendTx error: " + res.Error())
		}
	}
}
