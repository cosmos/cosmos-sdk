package app

import (
	"encoding/json"
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/examples/democoin/types"
	"github.com/cosmos/cosmos-sdk/examples/democoin/x/cool"
	"github.com/cosmos/cosmos-sdk/examples/democoin/x/pow"
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

	priv1 = crypto.GenPrivKeyEd25519()
	addr1 = priv1.PubKey().Address()
	addr2 = crypto.GenPrivKeyEd25519().PubKey().Address()
	coins = sdk.Coins{{"foocoin", 10}}
	fee   = sdk.StdFee{
		sdk.Coins{{"foocoin", 0}},
		0,
	}

	sendMsg = bank.SendMsg{
		Inputs:  []bank.Input{bank.NewInput(addr1, coins)},
		Outputs: []bank.Output{bank.NewOutput(addr2, coins)},
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

func loggerAndDBs() (log.Logger, map[string]dbm.DB) {
	logger := log.NewTMLogger(log.NewSyncWriter(os.Stdout)).With("module", "sdk/app")
	dbs := map[string]dbm.DB{
		"main":    dbm.NewMemDB(),
		"acc":     dbm.NewMemDB(),
		"pow":     dbm.NewMemDB(),
		"ibc":     dbm.NewMemDB(),
		"staking": dbm.NewMemDB(),
	}
	return logger, dbs
}

func newDemocoinApp() *DemocoinApp {
	logger, dbs := loggerAndDBs()
	return NewDemocoinApp(logger, dbs)
}

//_______________________________________________________________________

func TestMsgs(t *testing.T) {
	bapp := newDemocoinApp()

	msgs := []struct {
		msg sdk.Msg
	}{
		{sendMsg},
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
	logger, dbs := loggerAndDBs()
	bapp := NewDemocoinApp(logger, dbs)

	// Construct some genesis bytes to reflect democoin/types/AppAccount
	pk := crypto.GenPrivKeyEd25519().PubKey()
	addr := pk.Address()
	coins, err := sdk.ParseCoins("77foocoin,99barcoin")
	require.Nil(t, err)
	baseAcc := auth.BaseAccount{
		Address: addr,
		Coins:   coins,
	}
	acc := &types.AppAccount{baseAcc, "foobart"}

	genesisState := map[string]interface{}{
		"accounts": []*types.GenesisAccount{
			types.NewGenesisAccount(acc),
		},
		"cool": map[string]string{
			"trend": "ice-cold",
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
	bapp = NewDemocoinApp(logger, dbs)
	ctx = bapp.BaseApp.NewContext(true, abci.Header{})
	res1 = bapp.accountMapper.GetAccount(ctx, baseAcc.Address)
	assert.Equal(t, acc, res1)
}

func TestSendMsgWithAccounts(t *testing.T) {
	bapp := newDemocoinApp()

	// Construct some genesis bytes to reflect democoin/types/AppAccount
	// Give 77 foocoin to the first key
	coins, err := sdk.ParseCoins("77foocoin")
	require.Nil(t, err)
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

	// Initialize the chain
	vals := []abci.Validator{}
	bapp.InitChain(abci.RequestInitChain{vals, stateBytes})
	bapp.Commit()

	// A checkTx context (true)
	ctxCheck := bapp.BaseApp.NewContext(true, abci.Header{})
	res1 := bapp.accountMapper.GetAccount(ctxCheck, addr1)
	assert.Equal(t, acc1, res1)

	// Sign the tx
	sequences := []int64{0}
	sig := priv1.Sign(sdk.StdSignBytes(chainID, sequences, fee, sendMsg))
	tx := sdk.NewStdTx(sendMsg, fee, []sdk.StdSignature{{
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

func TestMineMsg(t *testing.T) {
	bapp := newDemocoinApp()

	// Construct genesis state
	// Construct some genesis bytes to reflect democoin/types/AppAccount
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
		"pow": map[string]uint64{
			"difficulty": 1,
			"count":      0,
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

	// Mine and check for reward
	mineMsg1 := pow.GenerateMineMsg(addr1, 1, 2)
	SignCheckDeliver(t, bapp, mineMsg1, 0, true)
	CheckBalance(t, bapp, "1pow")
	// Mine again and check for reward
	mineMsg2 := pow.GenerateMineMsg(addr1, 2, 3)
	SignCheckDeliver(t, bapp, mineMsg2, 1, true)
	CheckBalance(t, bapp, "2pow")
	// Mine again - should be invalid
	SignCheckDeliver(t, bapp, mineMsg2, 1, false)
	CheckBalance(t, bapp, "2pow")

}

func TestQuizMsg(t *testing.T) {
	bapp := newDemocoinApp()

	// Construct genesis state
	// Construct some genesis bytes to reflect democoin/types/AppAccount
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
	bapp := newDemocoinApp()

	sourceChain := "source-chain"
	destChain := "dest-chain"

	vals := []abci.Validator{}
	baseAcc := auth.BaseAccount{
		Address: addr1,
		Coins:   coins,
	}
	acc1 := &types.AppAccount{baseAcc, "foobart"}
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
	bapp.InitChain(abci.RequestInitChain{vals, stateBytes})
	bapp.Commit()

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
func SignCheckDeliver(t *testing.T, bapp *DemocoinApp, msg sdk.Msg, seq int64, expPass bool) {

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

func CheckBalance(t *testing.T, bapp *DemocoinApp, balExpected string) {
	ctxDeliver := bapp.BaseApp.NewContext(false, abci.Header{})
	res2 := bapp.accountMapper.GetAccount(ctxDeliver, addr1)
	assert.Equal(t, balExpected, fmt.Sprintf("%v", res2.GetCoins()))
}
