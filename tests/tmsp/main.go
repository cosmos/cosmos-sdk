package main

import (
	"fmt"

	"github.com/tendermint/basecoin/app"
	"github.com/tendermint/basecoin/tests"
	"github.com/tendermint/basecoin/types"
	. "github.com/tendermint/go-common"
	"github.com/tendermint/go-crypto"
	"github.com/tendermint/go-wire"
	"github.com/tendermint/go-wire/expr"
	govtypes "github.com/tendermint/governmint/types"
	eyescli "github.com/tendermint/merkleeyes/client"
	tmsp "github.com/tendermint/tmsp/types"
)

func main() {
	testSendTx()
	testGov()
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

	// Construct a SendTx signature
	tx := &types.SendTx{
		Fee: 0,
		Gas: 0,
		Inputs: []types.TxInput{
			types.TxInput{
				Address:  test1PrivAcc.Account.PubKey.Address(),
				PubKey:   test1PrivAcc.Account.PubKey, // TODO is this needed?
				Coins:    types.Coins{{"", 1}},
				Sequence: 1,
			},
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
	res := bcApp.AppendTx(txBytes)
	fmt.Println(res)
	if res.IsErr() {
		Exit(Fmt("Failed: %v", res.Error()))
	}
}

func testGov() {
	eyesCli := eyescli.NewLocalClient()
	chainID := "test_chain_id"
	bcApp := app.NewBasecoin(eyesCli)
	bcApp.SetOption("base/chainID", chainID)
	fmt.Println(bcApp.Info())

	adminPrivAcc := tests.PrivAccountFromSecret("admin")
	val0PrivKey := crypto.GenPrivKeyEd25519FromSecret([]byte("val0"))
	val1PrivKey := crypto.GenPrivKeyEd25519FromSecret([]byte("val1"))
	val2PrivKey := crypto.GenPrivKeyEd25519FromSecret([]byte("val2"))

	// Seed Basecoin with admin using PrivAccount
	adminAcc := adminPrivAcc.Account
	adminEntity := govtypes.Entity{
		Addr:   adminAcc.PubKey.Address(),
		PubKey: adminAcc.PubKey,
	}
	log := bcApp.SetOption("gov/admin", string(wire.JSONBytes(adminEntity)))
	if log != "Success" {
		Exit(Fmt("Failed to set option: %v", log))
	}
	adminAccount := types.Account{
		PubKey:   adminAcc.PubKey,
		Sequence: 0,
		Balance:  types.Coins{{"", 1 << 53}},
	}
	log = bcApp.SetOption("base/account", string(wire.JSONBytes(adminAccount)))
	if log != "Success" {
		Exit(Fmt("Failed to set option: %v", log))
	}

	// Call InitChain to initialize the validator set
	bcApp.InitChain([]*tmsp.Validator{
		{PubKey: val0PrivKey.PubKey().Bytes(), Power: 1},
		{PubKey: val1PrivKey.PubKey().Bytes(), Power: 1},
		{PubKey: val2PrivKey.PubKey().Bytes(), Power: 1},
	})

	// Query for validator set
	res := bcApp.Query(expr.MustCompile(`x02 x01 "gov/g/validators"`))
	if res.IsErr() {
		Exit(Fmt("Failed to query validators: %v", res.Error()))
	}
	group := govtypes.Group{}
	err := wire.ReadBinaryBytes(res.Data, &group)
	if err != nil {
		Exit(Fmt("Unexpected query response bytes: %X error: %v",
			res.Data, err))
	}
	// fmt.Println("Initialized gov/g/validators", group)

	// Mutate the validator set.
	proposal := govtypes.Proposal{
		ID:          "my_proposal_id",
		VoteGroupID: "admin",
		StartHeight: 0,
		EndHeight:   0,
		Info: &govtypes.GroupUpdateProposalInfo{
			UpdateGroupID: "validators",
			NextVersion:   0,
			ChangedMembers: []govtypes.Member{
				{nil, 1}, // TODO Fill this out.
			},
		},
	}
	proposalTx := &govtypes.ProposalTx{
		EntityAddr: adminEntity.Addr,
		Proposal:   proposal,
	}
	proposalTx.Signature = adminPrivAcc.Sign(proposalTx.SignBytes())
	tx := &types.AppTx{
		Fee:  1,
		Gas:  1,
		Type: app.PluginTypeByteGov, // XXX Remove typebytes?
		Input: types.TxInput{
			Address:  adminEntity.Addr,
			Coins:    types.Coins{{"", 1}},
			Sequence: 1,
			PubKey:   adminEntity.PubKey,
		},
		Data: wire.BinaryBytes(struct{ govtypes.Tx }{proposalTx}),
	}
	tx.SetSignature(adminPrivAcc.Sign(tx.SignBytes(chainID)))
	res = bcApp.AppendTx(wire.BinaryBytes(struct{ types.Tx }{tx}))
	if res.IsErr() {
		Exit(Fmt("Failed to mutate validators: %v", res.Error()))
	}
	fmt.Println(res)

	// TODO more tests...
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
				types.TxInput{
					Address:  test1Acc.PubKey.Address(),
					PubKey:   test1Acc.PubKey, // TODO is this needed?
					Coins:    types.Coins{{"", 1000002}},
					Sequence: sequence,
				},
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
				types.TxInput{
					Address:  privAccountA.Account.PubKey.Address(),
					PubKey:   privAccountA.Account.PubKey,
					Coins:    types.Coins{{"", 3}},
					Sequence: privAccountASequence + 1,
				},
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
