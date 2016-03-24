package main

import (
	"fmt"

	"github.com/tendermint/basecoin/app"
	"github.com/tendermint/basecoin/tests"
	"github.com/tendermint/basecoin/types"
	. "github.com/tendermint/go-common"
	"github.com/tendermint/go-wire"
	eyescli "github.com/tendermint/merkleeyes/client"
	_ "github.com/tendermint/tendermint/rpc/core/types" // Register RPCResponse > Result types
)

/*
	Get the "test" account.
	PrivKey: 019F86D081884C7D659A2FEAA0C55AD015A3BF4F1B2B0B822CD15D6C15B0F00A0867D3B5EAF0C0BF6B5A602D359DAECC86A7A74053490EC37AE08E71360587C870
	PubKey: 0167D3B5EAF0C0BF6B5A602D359DAECC86A7A74053490EC37AE08E71360587C870
	Address: D9B727742AA29FA638DC63D70813C976014C4CE0
*/
func main() {
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
}
