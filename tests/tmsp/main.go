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
	testGov()
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
	bcApp.SetOption("chainID", "test_chain_id")
	bcApp.SetOption("account", string(wire.JSONBytes(tAcc)))

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
	fmt.Printf("SIGNBYTES %X", signBytes)
	sig := tPriv.PrivKey.Sign(signBytes)
	tx.Inputs[0].Signature = sig
	//fmt.Println("tx:", tx)

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
	log := bcApp.SetOption("GOV:admin", string(wire.JSONBytes(adminEntity)))
	if log != "Success" {
		Exit(Fmt("Failed to set option: %v", log))
	}
	// TODO test proposals or something.
}
