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
	"github.com/cosmos/cosmos-sdk/x/ibc"

	abci "github.com/tendermint/abci/types"
	crypto "github.com/tendermint/go-crypto"
	dbm "github.com/tendermint/tmlibs/db"
	"github.com/tendermint/tmlibs/log"
)

// Construct some global addrs and txs for tests.
var (
	chainID = "" // TODO

	accName = "foobart"

	priv1     = crypto.GenPrivKeyEd25519()
	addr1     = priv1.PubKey().Address()
	priv2     = crypto.GenPrivKeyEd25519()
	addr2     = priv2.PubKey().Address()
	addr3     = crypto.GenPrivKeyEd25519().PubKey().Address()
	priv4     = crypto.GenPrivKeyEd25519()
	addr4     = priv4.PubKey().Address()
	coins     = sdk.Coins{{"foocoin", 10}}
	halfCoins = sdk.Coins{{"foocoin", 5}}
	fee       = sdk.StdFee{
		sdk.Coins{{"foocoin", 0}},
		0,
	}

	sendMsg1 = bank.SendMsg{
		Inputs:  []bank.Input{bank.NewInput(addr1, coins)},
		Outputs: []bank.Output{bank.NewOutput(addr2, coins)},
	}

	sendMsg2 = bank.SendMsg{
		Inputs: []bank.Input{bank.NewInput(addr1, coins)},
		Outputs: []bank.Output{
			bank.NewOutput(addr2, halfCoins),
			bank.NewOutput(addr3, halfCoins),
		},
	}

	sendMsg3 = bank.SendMsg{
		Inputs: []bank.Input{
			bank.NewInput(addr1, coins),
			bank.NewInput(addr4, coins),
		},
		Outputs: []bank.Output{
			bank.NewOutput(addr2, coins),
			bank.NewOutput(addr3, coins),
		},
	}

	sendMsg4 = bank.SendMsg{
		Inputs: []bank.Input{
			bank.NewInput(addr2, coins),
		},
		Outputs: []bank.Output{
			bank.NewOutput(addr1, coins),
		},
	}

	quizMsg1 = cool.QuizMsg{
		Sender:     addr1,
		CoolAnswer: "icecold",
	}

	quizMsg2 = cool.QuizMsg{
		Sender:     addr1,
		CoolAnswer: "badvibesonly",
	}

	setTrendMsg1 = cool.SetTrendMsg{
		Sender: addr1,
		Cool:   "icecold",
	}

	setTrendMsg2 = cool.SetTrendMsg{
		Sender: addr1,
		Cool:   "badvibesonly",
	}

	setTrendMsg3 = cool.SetTrendMsg{
		Sender: addr1,
		Cool:   "warmandkind",
	}
)

func newBasecoinApp() *BasecoinApp {
	logger := log.NewTMLogger(log.NewSyncWriter(os.Stdout)).With("module", "sdk/app")
	db := dbm.NewMemDB()
	return NewBasecoinApp(logger, db)
}

func setGenesisAccounts(bapp *BasecoinApp, accs ...auth.BaseAccount) error {
	genaccs := make([]*types.GenesisAccount, len(accs))
	for i, acc := range accs {
		genaccs[i] = types.NewGenesisAccount(&types.AppAccount{acc, accName})
	}

	genesisState := types.GenesisState{
		Accounts: genaccs,
	}

	stateBytes, err := json.MarshalIndent(genesisState, "", "\t")
	if err != nil {
		return err
	}

	// Initialize the chain
	vals := []abci.Validator{}
	bapp.InitChain(abci.RequestInitChain{vals, stateBytes})
	bapp.Commit()

	return nil
}

//_______________________________________________________________________

func TestMsgs(t *testing.T) {
	bapp := newBasecoinApp()

	msgs := []struct {
		msg sdk.Msg
	}{
		{sendMsg1},
		{quizMsg1},
		{setTrendMsg1},
	}

	sequences := []int64{0}
	for i, m := range msgs {
		sig := priv1.Sign(sdk.StdSignBytes(chainID, sequences, fee, m.msg))
		tx := sdk.NewStdTx(m.msg, fee, []sdk.StdSignature{{
			PubKey:    priv1.PubKey(),
			Signature: sig,
		}})

		// just marshal/unmarshal!
		cdc := MakeCodec()
		txBytes, err := cdc.MarshalBinary(tx)
		require.NoError(t, err, "i: %v", i)

		// Run a Check
		cres := bapp.CheckTx(txBytes)
		assert.Equal(t, sdk.CodeUnknownAddress,
			sdk.CodeType(cres.Code), "i: %v, log: %v", i, cres.Log)

		// Simulate a Block
		bapp.BeginBlock(abci.RequestBeginBlock{})
		dres := bapp.DeliverTx(txBytes)
		assert.Equal(t, sdk.CodeUnknownAddress,
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
	err = setGenesisAccounts(bapp, baseAcc)
	assert.Nil(t, err)

	// A checkTx context
	ctx := bapp.BaseApp.NewContext(true, abci.Header{})
	res1 := bapp.accountMapper.GetAccount(ctx, baseAcc.Address)
	assert.Equal(t, baseAcc, res1.(*types.AppAccount).BaseAccount)
	/*
		// reload app and ensure the account is still there
		bapp = NewBasecoinApp(logger, db)
		ctx = bapp.BaseApp.NewContext(true, abci.Header{})
		res1 = bapp.accountMapper.GetAccount(ctx, baseAcc.Address)
		assert.Equal(t, acc, res1)
	*/
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

	// Construct genesis state
	err = setGenesisAccounts(bapp, baseAcc)
	assert.Nil(t, err)

	// A checkTx context (true)
	ctxCheck := bapp.BaseApp.NewContext(true, abci.Header{})
	res1 := bapp.accountMapper.GetAccount(ctxCheck, addr1)
	assert.Equal(t, baseAcc, res1.(*types.AppAccount).BaseAccount)

	// Sign the tx
	sequences := []int64{0}
	sig := priv1.Sign(sdk.StdSignBytes(chainID, sequences, fee, sendMsg1))
	tx := sdk.NewStdTx(sendMsg1, fee, []sdk.StdSignature{{
		PubKey:    priv1.PubKey(),
		Signature: sig,
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
	tx.Signatures[0].Sequence = 1
	res = bapp.Deliver(tx)
	assert.Equal(t, sdk.CodeUnauthorized, res.Code, res.Log)

	// resigning the tx with the bumped sequence should work
	sequences = []int64{1}
	sig = priv1.Sign(sdk.StdSignBytes(chainID, sequences, fee, tx.Msg))
	tx.Signatures[0].Signature = sig
	res = bapp.Deliver(tx)
	assert.Equal(t, sdk.CodeOK, res.Code, res.Log)
}

func TestSendMsgMultipleOut(t *testing.T) {
	bapp := newBasecoinApp()

	genCoins, err := sdk.ParseCoins("42foocoin")
	require.Nil(t, err)

	acc1 := auth.BaseAccount{
		Address: addr1,
		Coins:   genCoins,
	}

	acc2 := auth.BaseAccount{
		Address: addr2,
		Coins:   genCoins,
	}

	err = setGenesisAccounts(bapp, acc1, acc2)
	assert.Nil(t, err)

	sequences := []int64{0}
	sig := priv1.Sign(sdk.StdSignBytes(chainID, sequences, fee, sendMsg2))
	tx := sdk.NewStdTx(sendMsg2, fee, []sdk.StdSignature{{
		PubKey:    priv1.PubKey(),
		Signature: sig,
	}})

	// Simulate a Block
	bapp.BeginBlock(abci.RequestBeginBlock{})
	res := bapp.Deliver(tx)
	assert.Equal(t, sdk.CodeOK, res.Code, res.Log)

	// Check balances
	ctx := bapp.BaseApp.NewContext(false, abci.Header{})
	acc := bapp.accountMapper.GetAccount(ctx, addr1)
	assert.Equal(t, fmt.Sprintf("%v", acc.GetCoins()), "32foocoin")
	acc = bapp.accountMapper.GetAccount(ctx, addr2)
	assert.Equal(t, fmt.Sprintf("%v", acc.GetCoins()), "47foocoin")
	acc = bapp.accountMapper.GetAccount(ctx, addr3)
	assert.Equal(t, fmt.Sprintf("%v", acc.GetCoins()), "5foocoin")
}

func TestSengMsgMultipleInOut(t *testing.T) {
	bapp := newBasecoinApp()

	genCoins, err := sdk.ParseCoins("42foocoin")
	require.Nil(t, err)

	acc1 := auth.BaseAccount{
		Address: addr1,
		Coins:   genCoins,
	}

	acc2 := auth.BaseAccount{
		Address: addr2,
		Coins:   genCoins,
	}

	acc4 := auth.BaseAccount{
		Address: addr4,
		Coins:   genCoins,
	}

	err = setGenesisAccounts(bapp, acc1, acc2, acc4)
	assert.Nil(t, err)

	sequences := []int64{0, 0}
	signbz := sdk.StdSignBytes(chainID, sequences, fee, sendMsg3)
	sig1 := priv1.Sign(signbz)
	sig4 := priv4.Sign(signbz)
	tx := sdk.NewStdTx(sendMsg3, fee, []sdk.StdSignature{
		{
			PubKey:    priv1.PubKey(),
			Signature: sig1,
		},
		{
			PubKey:    priv4.PubKey(),
			Signature: sig4,
		},
	})

	// Simulate a Block
	bapp.BeginBlock(abci.RequestBeginBlock{})
	res := bapp.Deliver(tx)
	assert.Equal(t, sdk.CodeOK, res.Code, res.Log)

	// Check balances
	ctx := bapp.BaseApp.NewContext(false, abci.Header{})
	acc := bapp.accountMapper.GetAccount(ctx, addr1)
	assert.Equal(t, fmt.Sprintf("%v", acc.GetCoins()), "32foocoin")
	acc = bapp.accountMapper.GetAccount(ctx, addr4)
	assert.Equal(t, fmt.Sprintf("%v", acc.GetCoins()), "32foocoin")
	acc = bapp.accountMapper.GetAccount(ctx, addr2)
	assert.Equal(t, fmt.Sprintf("%v", acc.GetCoins()), "52foocoin")
	acc = bapp.accountMapper.GetAccount(ctx, addr3)
	assert.Equal(t, fmt.Sprintf("%v", acc.GetCoins()), "10foocoin")

}

func TestSendMsgDependent(t *testing.T) {
	bapp := newBasecoinApp()

	genCoins, err := sdk.ParseCoins("42foocoin")
	require.Nil(t, err)

	acc1 := auth.BaseAccount{
		Address: addr1,
		Coins:   genCoins,
	}

	err = setGenesisAccounts(bapp, acc1)
	assert.Nil(t, err)

	sequences := []int64{0}

	// Simulate a block
	signbz := sdk.StdSignBytes(chainID, sequences, fee, sendMsg1)
	sig1 := priv1.Sign(signbz)
	tx := sdk.NewStdTx(sendMsg1, fee, []sdk.StdSignature{{
		PubKey:    priv1.PubKey(),
		Signature: sig1,
	}})

	bapp.BeginBlock(abci.RequestBeginBlock{})
	res := bapp.Deliver(tx)
	assert.Equal(t, sdk.CodeOK, res.Code, res.Log)

	// Check balances
	ctx := bapp.BaseApp.NewContext(false, abci.Header{})
	acc := bapp.accountMapper.GetAccount(ctx, addr1)
	assert.Equal(t, fmt.Sprintf("%v", acc.GetCoins()), "32foocoin")
	acc = bapp.accountMapper.GetAccount(ctx, addr2)
	assert.Equal(t, fmt.Sprintf("%v", acc.GetCoins()), "10foocoin")

	// Simulate a Block
	signbz = sdk.StdSignBytes(chainID, sequences, fee, sendMsg4)
	sig2 := priv2.Sign(signbz)
	tx = sdk.NewStdTx(sendMsg4, fee, []sdk.StdSignature{{
		PubKey:    priv2.PubKey(),
		Signature: sig2,
	}})

	res = bapp.Deliver(tx)
	assert.Equal(t, sdk.CodeOK, res.Code, res.Log)

	// Check balances
	ctx = bapp.BaseApp.NewContext(false, abci.Header{})
	acc = bapp.accountMapper.GetAccount(ctx, addr1)
	assert.Equal(t, fmt.Sprintf("%v", acc.GetCoins()), "42foocoin")
}

func TestQuizMsg(t *testing.T) {
	bapp := newBasecoinApp()

	// Construct genesis state
	// Construct some genesis bytes to reflect basecoin/types/AppAccount
	coins := sdk.Coins{}
	baseAcc := auth.BaseAccount{
		Address: addr1,
		Coins:   coins,
	}
	acc1 := &types.AppAccount{baseAcc, "foobart"}

	// Construct genesis state
	genesisState := map[string]interface{}{
		"accounts": []*types.GenesisAccount{
			types.NewGenesisAccount(acc1),
		},
		"cool": map[string]string{
			"trend": "ice-cold",
		},
	}
	stateBytes, err := json.MarshalIndent(genesisState, "", "\t")
	require.Nil(t, err)

	// Initialize the chain (nil)
	vals := []abci.Validator{}
	bapp.InitChain(abci.RequestInitChain{vals, stateBytes})
	bapp.Commit()

	// A checkTx context (true)
	ctxCheck := bapp.BaseApp.NewContext(true, abci.Header{})
	res1 := bapp.accountMapper.GetAccount(ctxCheck, addr1)
	assert.Equal(t, acc1, res1)

	// Set the trend, submit a really cool quiz and check for reward
	SignCheckDeliver(t, bapp, setTrendMsg1, 0, true)
	SignCheckDeliver(t, bapp, quizMsg1, 1, true)
	CheckBalance(t, bapp, "69icecold")
	SignCheckDeliver(t, bapp, quizMsg2, 2, false) // result without reward
	CheckBalance(t, bapp, "69icecold")
	SignCheckDeliver(t, bapp, quizMsg1, 3, true)
	CheckBalance(t, bapp, "138icecold")
	SignCheckDeliver(t, bapp, setTrendMsg2, 4, true) // reset the trend
	SignCheckDeliver(t, bapp, quizMsg1, 5, false)    // the same answer will nolonger do!
	CheckBalance(t, bapp, "138icecold")
	SignCheckDeliver(t, bapp, quizMsg2, 6, true) // earlier answer now relavent again
	CheckBalance(t, bapp, "69badvibesonly,138icecold")
	SignCheckDeliver(t, bapp, setTrendMsg3, 7, false) // expect to fail to set the trend to something which is not cool

}

func TestHandler(t *testing.T) {
	bapp := newBasecoinApp()

	sourceChain := "source-chain"
	destChain := "dest-chain"

	baseAcc := auth.BaseAccount{
		Address: addr1,
		Coins:   coins,
	}
	acc1 := &types.AppAccount{baseAcc, "foobart"}
	err := setGenesisAccounts(bapp, baseAcc)
	assert.Nil(t, err)

	// A checkTx context (true)
	ctxCheck := bapp.BaseApp.NewContext(true, abci.Header{})
	res1 := bapp.accountMapper.GetAccount(ctxCheck, addr1)
	assert.Equal(t, acc1, res1)

	packet := ibc.IBCPacket{
		SrcAddr:   addr1,
		DestAddr:  addr1,
		Coins:     coins,
		SrcChain:  sourceChain,
		DestChain: destChain,
	}

	transferMsg := ibc.IBCTransferMsg{
		IBCPacket: packet,
	}

	receiveMsg := ibc.IBCReceiveMsg{
		IBCPacket: packet,
		Relayer:   addr1,
		Sequence:  0,
	}

	SignCheckDeliver(t, bapp, transferMsg, 0, true)
	CheckBalance(t, bapp, "")
	SignCheckDeliver(t, bapp, transferMsg, 1, false)
	SignCheckDeliver(t, bapp, receiveMsg, 2, true)
	CheckBalance(t, bapp, "10foocoin")
	SignCheckDeliver(t, bapp, receiveMsg, 3, false)
}

// TODO describe the use of this function
func SignCheckDeliver(t *testing.T, bapp *BasecoinApp, msg sdk.Msg, seq int64, expPass bool) {

	// Sign the tx
	tx := sdk.NewStdTx(msg, fee, []sdk.StdSignature{{
		PubKey:    priv1.PubKey(),
		Signature: priv1.Sign(sdk.StdSignBytes(chainID, []int64{seq}, fee, msg)),
		Sequence:  seq,
	}})

	// Run a Check
	res := bapp.Check(tx)
	if expPass {
		require.Equal(t, sdk.CodeOK, res.Code, res.Log)
	} else {
		require.NotEqual(t, sdk.CodeOK, res.Code, res.Log)
	}

	// Simulate a Block
	bapp.BeginBlock(abci.RequestBeginBlock{})
	res = bapp.Deliver(tx)
	if expPass {
		require.Equal(t, sdk.CodeOK, res.Code, res.Log)
	} else {
		require.NotEqual(t, sdk.CodeOK, res.Code, res.Log)
	}
	bapp.EndBlock(abci.RequestEndBlock{})
	//bapp.Commit()
}

func CheckBalance(t *testing.T, bapp *BasecoinApp, balExpected string) {
	ctxDeliver := bapp.BaseApp.NewContext(false, abci.Header{})
	res2 := bapp.accountMapper.GetAccount(ctxDeliver, addr1)
	assert.Equal(t, balExpected, fmt.Sprintf("%v", res2.GetCoins()))
}
