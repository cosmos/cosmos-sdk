package app

import (
	"encoding/json"
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/examples/basecoin/types"
	"github.com/cosmos/cosmos-sdk/examples/basecoin/x/cool"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/bank"

	abci "github.com/tendermint/abci/types"
	crypto "github.com/tendermint/go-crypto"
	dbm "github.com/tendermint/tmlibs/db"
	"github.com/tendermint/tmlibs/log"
)

// helper variables and functions

var (
	// Construct genesis key/accounts
	priv1 = crypto.GenPrivKeyEd25519()
	addr1 = priv1.PubKey().Address()
	addr2 = crypto.GenPrivKeyEd25519().PubKey().Address()
	coins = sdk.Coins{{"foocoin", 10}}

	// Construct a SendMsg
	sendMsg = bank.SendMsg{
		Inputs:  []bank.Input{bank.NewInput(addr1, coins)},
		Outputs: []bank.Output{bank.NewOutput(addr2, coins)},
	}

	whatCoolMsg1 = cool.WhatCoolMsg{
		Sender:         addr1,
		CoolerThanCool: "icecold",
	}

	whatCoolMsg2 = cool.WhatCoolMsg{
		Sender:         addr1,
		CoolerThanCool: "icecold",
	}

	setWhatCoolMsg = cool.SetWhatCoolMsg{
		Sender:   addr1,
		WhatCool: "goodbye",
	}
)

func newBasecoinApp() *BasecoinApp {
	logger := log.NewTMLogger(log.NewSyncWriter(os.Stdout)).With("module", "sdk/app")
	db := dbm.NewMemDB()
	return NewBasecoinApp(logger, db)
}

//_______________________________________________________________________

func TestMsgs(t *testing.T) {
	bapp := newBasecoinApp()

	msgs := []struct {
		msg sdk.Msg
	}{
		{sendMsg},
		{whatCoolMsg1},
		{setWhatCoolMsg},
	}

	for i, m := range msgs {
		sig := priv1.Sign(m.msg.GetSignBytes())
		tx := sdk.NewStdTx(m.msg, []sdk.StdSignature{{
			PubKey:    priv1.PubKey(),
			Signature: sig,
		}})

		// just marshal/unmarshal!
		cdc := MakeCodec()
		txBytes, err := cdc.MarshalBinary(tx)
		require.NoError(t, err, "i: %v", i)

		// Run a Check
		cres := bapp.CheckTx(txBytes)
		assert.Equal(t, sdk.CodeUnrecognizedAddress,
			sdk.CodeType(cres.Code), "i: %v, log: %v", i, cres.Log)

		// Simulate a Block
		bapp.BeginBlock(abci.RequestBeginBlock{})
		dres := bapp.DeliverTx(txBytes)
		assert.Equal(t, sdk.CodeUnrecognizedAddress,
			sdk.CodeType(dres.Code), "i: %v, log: %v", i, dres.Log)
	}
}

func TestGenesis(t *testing.T) {
	logger := log.NewTMLogger(log.NewSyncWriter(os.Stdout)).With("module", "sdk/app")
	db := dbm.NewMemDB()
	bapp := NewBasecoinApp(logger, db)

	// Construct some genesis bytes to reflect basecoin/types/AppAccount
	pk := crypto.GenPrivKeyEd25519().PubKey()
	addr := pk.Address()
	coins, err := sdk.ParseCoins("77foocoin,99barcoin")
	require.Nil(t, err)
	baseAcc := auth.BaseAccount{
		Address: addr,
		Coins:   coins,
	}
	acc := &types.AppAccount{baseAcc, "foobart"}

	genesisState := types.GenesisState{
		Accounts: []*types.GenesisAccount{
			types.NewGenesisAccount(acc),
		},
	}
	stateBytes, err := json.MarshalIndent(genesisState, "", "\t")

	vals := []abci.Validator{}
	bapp.InitChain(abci.RequestInitChain{vals, stateBytes})
	bapp.Commit()

	// A checkTx context
	ctx := bapp.BaseApp.NewContext(true, abci.Header{})
	res1 := bapp.accountMapper.GetAccount(ctx, baseAcc.Address)
	assert.Equal(t, acc, res1)

	// reload app and ensure the account is still there
	bapp = NewBasecoinApp(logger, db)
	ctx = bapp.BaseApp.NewContext(true, abci.Header{})
	res1 = bapp.accountMapper.GetAccount(ctx, baseAcc.Address)
	assert.Equal(t, acc, res1)

}

func TestSendMsgWithAccounts(t *testing.T) {
	bapp := newBasecoinApp()

	// Construct some genesis bytes to reflect basecoin/types/AppAccount
	// Give 77 foocoin to the first key
	coins, err := sdk.ParseCoins("77foocoin")
	require.Nil(t, err)
	baseAcc := auth.BaseAccount{
		Address: addr1,
		Coins:   coins,
	}
	acc1 := &types.AppAccount{baseAcc, "foobart"}

	// Construct genesis state
	genesisState := types.GenesisState{
		Accounts: []*types.GenesisAccount{
			types.NewGenesisAccount(acc1),
		},
	}
	stateBytes, err := json.MarshalIndent(genesisState, "", "\t")
	require.Nil(t, err)

	// Initialize the chain
	vals := []abci.Validator{}
	bapp.InitChain(abci.RequestInitChain{vals, stateBytes})
	bapp.Commit()

	// A checkTx context (true)
	ctxCheck := bapp.BaseApp.NewContext(true, abci.Header{})
	res1 := bapp.accountMapper.GetAccount(ctxCheck, addr1)
	assert.Equal(t, acc1, res1)

	// Sign the tx
	chainID := "" // TODO: InitChain should get the ChainID
	sequence := int64(0)
	sig := priv1.Sign(sdk.StdSignBytes(chainID, sequence, sendMsg))
	tx := sdk.NewStdTx(sendMsg, []sdk.StdSignature{{
		PubKey:    priv1.PubKey(),
		Signature: priv1.Sign(sendMsg.GetSignBytes()),
	}})

	// Run a Check
	res := bapp.Check(tx)
	assert.Equal(t, sdk.CodeOK, res.Code, res.Log)

	// Simulate a Block
	bapp.BeginBlock(abci.RequestBeginBlock{})
	res = bapp.Deliver(tx)
	assert.Equal(t, sdk.CodeOK, res.Code, res.Log)

	// Check balances
	ctxDeliver := bapp.BaseApp.NewContext(false, abci.Header{})
	res2 := bapp.accountMapper.GetAccount(ctxDeliver, addr1)
	res3 := bapp.accountMapper.GetAccount(ctxDeliver, addr2)
	assert.Equal(t, fmt.Sprintf("%v", res2.GetCoins()), "67foocoin")
	assert.Equal(t, fmt.Sprintf("%v", res3.GetCoins()), "10foocoin")

	// Delivering again should cause replay error
	res = bapp.Deliver(tx)
	assert.Equal(t, sdk.CodeInvalidSequence, res.Code, res.Log)

	// bumping the txnonce number without resigning should be an auth error
	sequence += 1
	tx.Signatures[0].Sequence = sequence
	res = bapp.Deliver(tx)
	assert.Equal(t, sdk.CodeUnauthorized, res.Code, res.Log)

	// resigning the tx with the bumped sequence should work
	sig = priv1.Sign(sdk.StdSignBytes(chainID, sequence, tx.Msg))
	tx.Signatures[0].Signature = sig
	res = bapp.Deliver(tx)
	assert.Equal(t, sdk.CodeOK, res.Code, res.Log)
}

//func TestWhatCoolMsg(t *testing.T) {
//bapp := newBasecoinApp()

//// Construct genesis state
//// Construct some genesis bytes to reflect basecoin/types/AppAccount
//// Give 77 foocoin to the first key
//coins, err := sdk.ParseCoins("1icecold")
//require.Nil(t, err)
//baseAcc := auth.BaseAccount{
//Address: addr1,
//Coins:   coins,
//}
//acc1 := &types.AppAccount{baseAcc, "foobart"}

//// Construct genesis state
//genesisState := types.GenesisState{
//Accounts: []*types.GenesisAccount{
//types.NewGenesisAccount(acc1),
//},
//}
//stateBytes, err := json.MarshalIndent(genesisState, "", "\t")
//require.Nil(t, err)

//// Initialize the chain (nil)
//vals := []abci.Validator{}
//bapp.InitChain(abci.RequestInitChain{vals, stateBytes})
//bapp.Commit()

//// A checkTx context (true)
//ctxCheck := bapp.BaseApp.NewContext(true, abci.Header{})
//res1 := bapp.accountMapper.GetAccount(ctxCheck, addr1)
//assert.Equal(t, acc1, res1)

//// Sign the tx
//tx := sdk.NewStdTx(whatCoolMsg1, []sdk.StdSignature{{
//PubKey:    priv1.PubKey(),
//Signature: priv1.Sign(whatCoolMsg1.GetSignBytes()),
//}})

//// Run a Check
//res := bapp.Check(tx)
//assert.Equal(t, sdk.CodeOK, res.Code, res.Log)

//// Simulate a Block
//bapp.BeginBlock(abci.RequestBeginBlock{})
//res = bapp.Deliver(tx)
//assert.Equal(t, sdk.CodeOK, res.Code, res.Log)

//// Check balances
//ctxDeliver := bapp.BaseApp.NewContext(false, abci.Header{})
//res2 := bapp.accountMapper.GetAccount(ctxDeliver, addr1)
//assert.Equal(t, "70icecold", fmt.Sprintf("%v", res2.GetCoins()))
//}
