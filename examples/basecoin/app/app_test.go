package app

import (
	"encoding/json"
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/examples/basecoin/types"
	tests "github.com/cosmos/cosmos-sdk/mock"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/wire"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/bank"
	"github.com/cosmos/cosmos-sdk/x/ibc"
	"github.com/cosmos/cosmos-sdk/x/stake"

	abci "github.com/tendermint/abci/types"
	crypto "github.com/tendermint/go-crypto"
	dbm "github.com/tendermint/tmlibs/db"
	"github.com/tendermint/tmlibs/log"
)

// Construct some global addrs and txs for tests.
var (
	chainID = "" // TODO

	accName = "foobart"

	keys, addrs = tests.GenerateNPrivKeyAddressPairs(4)
	coins       = sdk.Coins{{"foocoin", 10}}
	halfCoins   = sdk.Coins{{"foocoin", 5}}
	manyCoins   = sdk.Coins{{"foocoin", 1}, {"barcoin", 1}}
	fee         = auth.StdFee{
		sdk.Coins{{"foocoin", 0}},
		100000,
	}

	sendMsg1 = bank.MsgSend{
		Inputs:  []bank.Input{bank.NewInput(addrs[0], coins)},
		Outputs: []bank.Output{bank.NewOutput(addrs[1], coins)},
	}

	sendMsg2 = bank.MsgSend{
		Inputs: []bank.Input{bank.NewInput(addrs[0], coins)},
		Outputs: []bank.Output{
			bank.NewOutput(addrs[1], halfCoins),
			bank.NewOutput(addrs[2], halfCoins),
		},
	}

	sendMsg3 = bank.MsgSend{
		Inputs: []bank.Input{
			bank.NewInput(addrs[0], coins),
			bank.NewInput(addrs[3], coins),
		},
		Outputs: []bank.Output{
			bank.NewOutput(addrs[1], coins),
			bank.NewOutput(addrs[2], coins),
		},
	}

	sendMsg4 = bank.MsgSend{
		Inputs: []bank.Input{
			bank.NewInput(addrs[1], coins),
		},
		Outputs: []bank.Output{
			bank.NewOutput(addrs[0], coins),
		},
	}

	sendMsg5 = bank.MsgSend{
		Inputs: []bank.Input{
			bank.NewInput(addrs[0], manyCoins),
		},
		Outputs: []bank.Output{
			bank.NewOutput(addrs[1], manyCoins),
		},
	}
)

func setGenesis(bapp *BasecoinApp, accs ...auth.BaseAccount) error {
	genaccs := make([]*types.GenesisAccount, len(accs))
	for i, acc := range accs {
		genaccs[i] = types.NewGenesisAccount(&types.AppAccount{acc, accName})
	}

	genesisState := types.GenesisState{
		Accounts:  genaccs,
		StakeData: stake.DefaultGenesisState(),
	}

	stateBytes, err := wire.MarshalJSONIndent(bapp.cdc, genesisState)
	if err != nil {
		return err
	}

	// Initialize the chain
	vals := []abci.Validator{}
	bapp.InitChain(abci.RequestInitChain{Validators: vals, AppStateBytes: stateBytes})
	bapp.Commit()

	return nil
}

func loggerAndDB() (log.Logger, dbm.DB) {
	logger := log.NewTMLogger(log.NewSyncWriter(os.Stdout)).With("module", "sdk/app")
	db := dbm.NewMemDB()
	return logger, db
}

func newBasecoinApp() *BasecoinApp {
	logger, db := loggerAndDB()
	return NewBasecoinApp(logger, db)
}

//_______________________________________________________________________

func TestMsgs(t *testing.T) {
	bapp := newBasecoinApp()
	require.Nil(t, setGenesis(bapp))

	msgs := []struct {
		msg sdk.Msg
	}{
		{sendMsg1},
	}

	for i, m := range msgs {
		// Run a CheckDeliver
		SignCheckDeliver(t, bapp, m.msg, []int64{int64(i)}, false, keys[0])
	}
}

func TestSortGenesis(t *testing.T) {
	logger, db := loggerAndDB()
	bapp := NewBasecoinApp(logger, db)

	// Note the order: the coins are unsorted!
	coinDenom1, coinDenom2 := "foocoin", "barcoin"

	genState := fmt.Sprintf(`{
      "accounts": [{
        "address": "%s",
        "coins": [
          {
            "denom": "%s",
            "amount": 10
          },
          {
            "denom": "%s",
            "amount": 20
          }
        ]
      }]
    }`, addrs[0].String(), coinDenom1, coinDenom2)

	// Initialize the chain
	vals := []abci.Validator{}
	bapp.InitChain(abci.RequestInitChain{Validators: vals, AppStateBytes: []byte(genState)})
	bapp.Commit()

	// Unsorted coins means invalid
	err := sendMsg5.ValidateBasic()
	require.Equal(t, sdk.CodeInvalidCoins, err.Code(), err.ABCILog())

	// Sort coins, should be valid
	sendMsg5.Inputs[0].Coins.Sort()
	sendMsg5.Outputs[0].Coins.Sort()
	err = sendMsg5.ValidateBasic()
	require.Nil(t, err)

	// Ensure we can send
	SignCheckDeliver(t, bapp, sendMsg5, []int64{0}, true, keys[0])
}

func TestGenesis(t *testing.T) {
	logger, db := loggerAndDB()
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

	err = setGenesis(bapp, baseAcc)
	require.Nil(t, err)

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

func TestMsgChangePubKey(t *testing.T) {

	bapp := newBasecoinApp()

	// Construct some genesis bytes to reflect basecoin/types/AppAccount
	// Give 77 foocoin to the first key
	coins, err := sdk.ParseCoins("77foocoin")
	require.Nil(t, err)
	baseAcc := auth.BaseAccount{
		Address: addrs[0],
		Coins:   coins,
	}

	// Construct genesis state
	err = setGenesis(bapp, baseAcc)
	require.Nil(t, err)

	// A checkTx context (true)
	ctxCheck := bapp.BaseApp.NewContext(true, abci.Header{})
	res1 := bapp.accountMapper.GetAccount(ctxCheck, addrs[0])
	assert.Equal(t, baseAcc, res1.(*types.AppAccount).BaseAccount)

	// Run a CheckDeliver
	SignCheckDeliver(t, bapp, sendMsg1, []int64{0}, true, keys[0])

	// Check balances
	CheckBalance(t, bapp, addrs[0], "67foocoin")
	CheckBalance(t, bapp, addrs[1], "10foocoin")

	changePubKeyMsg := auth.MsgChangeKey{
		Address:   addrs[0],
		NewPubKey: keys[1].PubKey(),
	}

	ctxDeliver := bapp.BaseApp.NewContext(false, abci.Header{})
	acc := bapp.accountMapper.GetAccount(ctxDeliver, addrs[0])

	// send a MsgChangePubKey
	SignCheckDeliver(t, bapp, changePubKeyMsg, []int64{1}, true, keys[0])
	acc = bapp.accountMapper.GetAccount(ctxDeliver, addrs[0])

	assert.True(t, keys[1].PubKey().Equals(acc.GetPubKey()))

	// signing a SendMsg with the old privKey should be an auth error
	tx := genTx(sendMsg1, []int64{2}, keys[0])
	res := bapp.Deliver(tx)
	assert.Equal(t, sdk.ToABCICode(sdk.CodespaceRoot, sdk.CodeUnauthorized), res.Code, res.Log)

	// resigning the tx with the new correct priv key should work
	SignCheckDeliver(t, bapp, sendMsg1, []int64{2}, true, keys[1])

	// Check balances
	CheckBalance(t, bapp, addrs[0], "57foocoin")
	CheckBalance(t, bapp, addrs[1], "20foocoin")
}

func TestMsgSendWithAccounts(t *testing.T) {
	bapp := newBasecoinApp()

	// Construct some genesis bytes to reflect basecoin/types/AppAccount
	// Give 77 foocoin to the first key
	coins, err := sdk.ParseCoins("77foocoin")
	require.Nil(t, err)
	baseAcc := auth.BaseAccount{
		Address: addrs[0],
		Coins:   coins,
	}

	// Construct genesis state
	err = setGenesis(bapp, baseAcc)
	require.Nil(t, err)

	// A checkTx context (true)
	ctxCheck := bapp.BaseApp.NewContext(true, abci.Header{})
	res1 := bapp.accountMapper.GetAccount(ctxCheck, addrs[0])
	assert.Equal(t, baseAcc, res1.(*types.AppAccount).BaseAccount)

	// Run a CheckDeliver
	SignCheckDeliver(t, bapp, sendMsg1, []int64{0}, true, keys[0])

	// Check balances
	CheckBalance(t, bapp, addrs[0], "67foocoin")
	CheckBalance(t, bapp, addrs[1], "10foocoin")

	// Delivering again should cause replay error
	SignCheckDeliver(t, bapp, sendMsg1, []int64{0}, false, keys[0])

	// bumping the txnonce number without resigning should be an auth error
	tx := genTx(sendMsg1, []int64{0}, keys[0])
	tx.Signatures[0].Sequence = 1
	res := bapp.Deliver(tx)

	assert.Equal(t, sdk.ToABCICode(sdk.CodespaceRoot, sdk.CodeUnauthorized), res.Code, res.Log)

	// resigning the tx with the bumped sequence should work
	SignCheckDeliver(t, bapp, sendMsg1, []int64{1}, true, keys[0])
}

func TestMsgSendMultipleOut(t *testing.T) {
	bapp := newBasecoinApp()

	genCoins, err := sdk.ParseCoins("42foocoin")
	require.Nil(t, err)

	acc1 := auth.BaseAccount{
		Address: addrs[0],
		Coins:   genCoins,
	}

	acc2 := auth.BaseAccount{
		Address: addrs[1],
		Coins:   genCoins,
	}

	// Construct genesis state
	err = setGenesis(bapp, acc1, acc2)
	require.Nil(t, err)

	// Simulate a Block
	SignCheckDeliver(t, bapp, sendMsg2, []int64{0}, true, keys[0])

	// Check balances
	CheckBalance(t, bapp, addrs[0], "32foocoin")
	CheckBalance(t, bapp, addrs[1], "47foocoin")
	CheckBalance(t, bapp, addrs[2], "5foocoin")
}

func TestSengMsgMultipleInOut(t *testing.T) {
	bapp := newBasecoinApp()

	genCoins, err := sdk.ParseCoins("42foocoin")
	require.Nil(t, err)

	acc1 := auth.BaseAccount{
		Address: addrs[0],
		Coins:   genCoins,
	}

	acc2 := auth.BaseAccount{
		Address: addrs[1],
		Coins:   genCoins,
	}

	acc4 := auth.BaseAccount{
		Address: addrs[3],
		Coins:   genCoins,
	}

	err = setGenesis(bapp, acc1, acc2, acc4)
	assert.Nil(t, err)

	// CheckDeliver
	SignCheckDeliver(t, bapp, sendMsg3, []int64{0, 0}, true, keys[0], keys[3])

	// Check balances
	CheckBalance(t, bapp, addrs[0], "32foocoin")
	CheckBalance(t, bapp, addrs[3], "32foocoin")
	CheckBalance(t, bapp, addrs[1], "52foocoin")
	CheckBalance(t, bapp, addrs[2], "10foocoin")
}

func TestMsgSendDependent(t *testing.T) {
	bapp := newBasecoinApp()

	genCoins, err := sdk.ParseCoins("42foocoin")
	require.Nil(t, err)

	acc1 := auth.BaseAccount{
		Address: addrs[0],
		Coins:   genCoins,
	}

	// Construct genesis state
	err = setGenesis(bapp, acc1)
	require.Nil(t, err)

	err = setGenesis(bapp, acc1)
	assert.Nil(t, err)

	// CheckDeliver
	SignCheckDeliver(t, bapp, sendMsg1, []int64{0}, true, keys[0])

	// Check balances
	CheckBalance(t, bapp, addrs[0], "32foocoin")
	CheckBalance(t, bapp, addrs[1], "10foocoin")

	// Simulate a Block
	SignCheckDeliver(t, bapp, sendMsg4, []int64{0}, true, keys[1])

	// Check balances
	CheckBalance(t, bapp, addrs[0], "42foocoin")
}

func TestMsgQuiz(t *testing.T) {
	bapp := newBasecoinApp()

	// Construct genesis state
	// Construct some genesis bytes to reflect basecoin/types/AppAccount
	baseAcc := auth.BaseAccount{
		Address: addrs[0],
		Coins:   nil,
	}
	acc1 := &types.AppAccount{baseAcc, "foobart"}

	// Construct genesis state
	genesisState := map[string]interface{}{
		"accounts": []*types.GenesisAccount{
			types.NewGenesisAccount(acc1),
		},
	}
	stateBytes, err := json.MarshalIndent(genesisState, "", "\t")
	require.Nil(t, err)

	// Initialize the chain (nil)
	vals := []abci.Validator{}
	bapp.InitChain(abci.RequestInitChain{Validators: vals, AppStateBytes: stateBytes})
	bapp.Commit()

	// A checkTx context (true)
	ctxCheck := bapp.BaseApp.NewContext(true, abci.Header{})
	res1 := bapp.accountMapper.GetAccount(ctxCheck, addrs[0])
	assert.Equal(t, acc1, res1)

}

func TestIBCMsgs(t *testing.T) {
	bapp := newBasecoinApp()

	sourceChain := "source-chain"
	destChain := "dest-chain"

	baseAcc := auth.BaseAccount{
		Address: addrs[0],
		Coins:   coins,
	}
	acc1 := &types.AppAccount{baseAcc, "foobart"}

	err := setGenesis(bapp, baseAcc)
	assert.Nil(t, err)

	// A checkTx context (true)
	ctxCheck := bapp.BaseApp.NewContext(true, abci.Header{})
	res1 := bapp.accountMapper.GetAccount(ctxCheck, addrs[0])
	assert.Equal(t, acc1, res1)

	packet := ibc.IBCPacket{
		SrcAddr:   addrs[0],
		DestAddr:  addrs[0],
		Coins:     coins,
		SrcChain:  sourceChain,
		DestChain: destChain,
	}

	transferMsg := ibc.IBCTransferMsg{
		IBCPacket: packet,
	}

	receiveMsg := ibc.IBCReceiveMsg{
		IBCPacket: packet,
		Relayer:   addrs[0],
		Sequence:  0,
	}

	SignCheckDeliver(t, bapp, transferMsg, []int64{0}, true, keys[0])
	CheckBalance(t, bapp, addrs[0], "")
	SignCheckDeliver(t, bapp, transferMsg, []int64{1}, false, keys[0])
	SignCheckDeliver(t, bapp, receiveMsg, []int64{2}, true, keys[0])
	CheckBalance(t, bapp, addrs[0], "10foocoin")
	SignCheckDeliver(t, bapp, receiveMsg, []int64{3}, false, keys[0])
}

func genTx(msg sdk.Msg, seq []int64, priv ...crypto.PrivKey) auth.StdTx {
	sigs := make([]auth.StdSignature, len(priv))
	for i, p := range priv {
		sigs[i] = auth.StdSignature{
			PubKey:    p.PubKey(),
			Signature: p.Sign(auth.StdSignBytes(chainID, seq, fee, msg)),
			Sequence:  seq[i],
		}
	}

	return auth.NewStdTx(msg, fee, sigs)

}

func SignCheckDeliver(t *testing.T, bapp *BasecoinApp, msg sdk.Msg, seq []int64, expPass bool, priv ...crypto.PrivKey) {
	// Sign the tx
	tx := genTx(msg, seq, priv...)
	tests.CheckDeliver(t, bapp.BaseApp, tx, expPass)
	//bapp.Commit()
}

func CheckBalance(t *testing.T, bapp *BasecoinApp, addr sdk.Address, balExpected string) {
	ctxDeliver := bapp.BaseApp.NewContext(false, abci.Header{})
	res2 := bapp.accountMapper.GetAccount(ctxDeliver, addr)
	assert.Equal(t, balExpected, fmt.Sprintf("%v", res2.GetCoins()))
}
