package app

import (
	"encoding/json"
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"
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

	priv1     = crypto.GenPrivKeyEd25519()
	addr1     = priv1.PubKey().Address()
	priv2     = crypto.GenPrivKeyEd25519()
	addr2     = priv2.PubKey().Address()
	addr3     = crypto.GenPrivKeyEd25519().PubKey().Address()
	priv4     = crypto.GenPrivKeyEd25519()
	addr4     = priv4.PubKey().Address()
	coins     = sdk.Coins{{"foocoin", 10}}
	halfCoins = sdk.Coins{{"foocoin", 5}}
	manyCoins = sdk.Coins{{"foocoin", 1}, {"barcoin", 1}}
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

	sendMsg5 = bank.SendMsg{
		Inputs: []bank.Input{
			bank.NewInput(addr1, manyCoins),
		},
		Outputs: []bank.Output{
			bank.NewOutput(addr2, manyCoins),
		},
	}
)

func loggerAndDB() (log.Logger, dbm.DB) {
	logger := log.NewTMLogger(log.NewSyncWriter(os.Stdout)).With("module", "sdk/app")
	db := dbm.NewMemDB()
	return logger, db
}

func newGaiaApp() *GaiaApp {
	logger, db := loggerAndDB()
	return NewGaiaApp(logger, db)
}

func setGenesis(gapp *GaiaApp, accs ...*auth.BaseAccount) error {
	genaccs := make([]GenesisAccount, len(accs))
	for i, acc := range accs {
		genaccs[i] = NewGenesisAccount(acc)
	}

	genesisState := GenesisState{
		Accounts:  genaccs,
		StakeData: stake.GetDefaultGenesisState(),
	}

	stateBytes, err := json.MarshalIndent(genesisState, "", "\t")
	if err != nil {
		return err
	}

	// Initialize the chain
	vals := []abci.Validator{}
	gapp.InitChain(abci.RequestInitChain{vals, stateBytes})
	gapp.Commit()

	return nil
}

//_______________________________________________________________________

func TestMsgs(t *testing.T) {
	gapp := newGaiaApp()
	require.Nil(t, setGenesis(gapp))

	msgs := []struct {
		msg sdk.Msg
	}{
		{sendMsg1},
	}

	for i, m := range msgs {
		// Run a CheckDeliver
		SignCheckDeliver(t, gapp, m.msg, []int64{int64(i)}, false, priv1)
	}
}

func TestSortGenesis(t *testing.T) {
	logger, dbs := loggerAndDB()
	gapp := NewGaiaApp(logger, dbs)

	// Note the order: the coins are unsorted!
	coinDenom1, coinDenom2 := "foocoin", "barcoin"

	str := `{
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
    }`
	genState := fmt.Sprintf(str, addr1.String(), coinDenom1, coinDenom2)

	// Initialize the chain
	vals := []abci.Validator{}
	gapp.InitChain(abci.RequestInitChain{vals, []byte(genState)})
	gapp.Commit()

	// Unsorted coins means invalid
	err := sendMsg5.ValidateBasic()
	require.Equal(t, sdk.CodeInvalidCoins, err.ABCICode(), err.ABCILog())

	// Sort coins, should be valid
	sendMsg5.Inputs[0].Coins.Sort()
	sendMsg5.Outputs[0].Coins.Sort()
	err = sendMsg5.ValidateBasic()
	require.Nil(t, err)

	// Ensure we can send
	require.Nil(t, setGenesis(gapp)) // initialize the pool
	SignCheckDeliver(t, gapp, sendMsg5, []int64{0}, true, priv1)
}

func TestGenesis(t *testing.T) {
	logger, dbs := loggerAndDB()
	gapp := NewGaiaApp(logger, dbs)

	// Construct some genesis bytes to reflect GaiaAccount
	pk := crypto.GenPrivKeyEd25519().PubKey()
	addr := pk.Address()
	coins, err := sdk.ParseCoins("77foocoin,99barcoin")
	require.Nil(t, err)
	baseAcc := &auth.BaseAccount{
		Address: addr,
		Coins:   coins,
	}

	err = setGenesis(gapp, baseAcc)
	assert.Nil(t, err)

	// A checkTx context
	ctx := gapp.BaseApp.NewContext(true, abci.Header{})
	res1 := gapp.accountMapper.GetAccount(ctx, baseAcc.Address)
	assert.Equal(t, baseAcc, res1)

	// reload app and ensure the account is still there
	gapp = NewGaiaApp(logger, dbs)
	ctx = gapp.BaseApp.NewContext(true, abci.Header{})
	res1 = gapp.accountMapper.GetAccount(ctx, baseAcc.Address)
	assert.Equal(t, baseAcc, res1)
}

func TestSendMsgWithAccounts(t *testing.T) {
	gapp := newGaiaApp()

	// Construct some genesis bytes to reflect GaiaAccount
	// Give 77 foocoin to the first key
	coins, err := sdk.ParseCoins("77foocoin")
	require.Nil(t, err)
	baseAcc := &auth.BaseAccount{
		Address: addr1,
		Coins:   coins,
	}

	// Construct genesis state
	err = setGenesis(gapp, baseAcc)
	require.Nil(t, err)

	// A checkTx context (true)
	ctxCheck := gapp.BaseApp.NewContext(true, abci.Header{})
	res1 := gapp.accountMapper.GetAccount(ctxCheck, addr1)
	assert.Equal(t, baseAcc, res1.(*auth.BaseAccount))

	// Run a CheckDeliver
	SignCheckDeliver(t, gapp, sendMsg1, []int64{0}, true, priv1)

	// Check balances
	CheckBalance(t, gapp, addr1, "67foocoin")
	CheckBalance(t, gapp, addr2, "10foocoin")

	// Delivering again should cause replay error
	SignCheckDeliver(t, gapp, sendMsg1, []int64{0}, false, priv1)

	// bumping the txnonce number without resigning should be an auth error
	tx := genTx(sendMsg1, []int64{0}, priv1)
	tx.Signatures[0].Sequence = 1
	res := gapp.Deliver(tx)

	assert.Equal(t, sdk.CodeUnauthorized, res.Code, res.Log)

	// resigning the tx with the bumped sequence should work
	SignCheckDeliver(t, gapp, sendMsg1, []int64{1}, true, priv1)
}

func TestSendMsgMultipleOut(t *testing.T) {
	gapp := newGaiaApp()

	genCoins, err := sdk.ParseCoins("42foocoin")
	require.Nil(t, err)

	acc1 := &auth.BaseAccount{
		Address: addr1,
		Coins:   genCoins,
	}

	acc2 := &auth.BaseAccount{
		Address: addr2,
		Coins:   genCoins,
	}

	err = setGenesis(gapp, acc1, acc2)
	require.Nil(t, err)

	// Simulate a Block
	SignCheckDeliver(t, gapp, sendMsg2, []int64{0}, true, priv1)

	// Check balances
	CheckBalance(t, gapp, addr1, "32foocoin")
	CheckBalance(t, gapp, addr2, "47foocoin")
	CheckBalance(t, gapp, addr3, "5foocoin")
}

func TestSengMsgMultipleInOut(t *testing.T) {
	gapp := newGaiaApp()

	genCoins, err := sdk.ParseCoins("42foocoin")
	require.Nil(t, err)

	acc1 := &auth.BaseAccount{
		Address: addr1,
		Coins:   genCoins,
	}
	acc2 := &auth.BaseAccount{
		Address: addr2,
		Coins:   genCoins,
	}
	acc4 := &auth.BaseAccount{
		Address: addr4,
		Coins:   genCoins,
	}

	err = setGenesis(gapp, acc1, acc2, acc4)
	assert.Nil(t, err)

	// CheckDeliver
	SignCheckDeliver(t, gapp, sendMsg3, []int64{0, 0}, true, priv1, priv4)

	// Check balances
	CheckBalance(t, gapp, addr1, "32foocoin")
	CheckBalance(t, gapp, addr4, "32foocoin")
	CheckBalance(t, gapp, addr2, "52foocoin")
	CheckBalance(t, gapp, addr3, "10foocoin")
}

func TestSendMsgDependent(t *testing.T) {
	gapp := newGaiaApp()

	genCoins, err := sdk.ParseCoins("42foocoin")
	require.Nil(t, err)

	acc1 := &auth.BaseAccount{
		Address: addr1,
		Coins:   genCoins,
	}

	err = setGenesis(gapp, acc1)
	require.Nil(t, err)

	// CheckDeliver
	SignCheckDeliver(t, gapp, sendMsg1, []int64{0}, true, priv1)

	// Check balances
	CheckBalance(t, gapp, addr1, "32foocoin")
	CheckBalance(t, gapp, addr2, "10foocoin")

	// Simulate a Block
	SignCheckDeliver(t, gapp, sendMsg4, []int64{0}, true, priv2)

	// Check balances
	CheckBalance(t, gapp, addr1, "42foocoin")
}

func TestIBCMsgs(t *testing.T) {
	gapp := newGaiaApp()

	sourceChain := "source-chain"
	destChain := "dest-chain"

	baseAcc := &auth.BaseAccount{
		Address: addr1,
		Coins:   coins,
	}

	err := setGenesis(gapp, baseAcc)
	require.Nil(t, err)

	// A checkTx context (true)
	ctxCheck := gapp.BaseApp.NewContext(true, abci.Header{})
	res1 := gapp.accountMapper.GetAccount(ctxCheck, addr1)
	assert.Equal(t, baseAcc, res1)

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

	SignCheckDeliver(t, gapp, transferMsg, []int64{0}, true, priv1)
	CheckBalance(t, gapp, addr1, "")
	SignCheckDeliver(t, gapp, transferMsg, []int64{1}, false, priv1)
	SignCheckDeliver(t, gapp, receiveMsg, []int64{2}, true, priv1)
	CheckBalance(t, gapp, addr1, "10foocoin")
	SignCheckDeliver(t, gapp, receiveMsg, []int64{3}, false, priv1)
}

func TestStakeMsgs(t *testing.T) {
	gapp := newGaiaApp()

	genCoins, err := sdk.ParseCoins("42fermion")
	require.Nil(t, err)
	bondCoin, err := sdk.ParseCoin("10fermion")
	require.Nil(t, err)

	acc1 := &auth.BaseAccount{
		Address: addr1,
		Coins:   genCoins,
	}
	acc2 := &auth.BaseAccount{
		Address: addr2,
		Coins:   genCoins,
	}

	err = setGenesis(gapp, acc1, acc2)
	require.Nil(t, err)

	// A checkTx context (true)
	ctxCheck := gapp.BaseApp.NewContext(true, abci.Header{})
	res1 := gapp.accountMapper.GetAccount(ctxCheck, addr1)
	res2 := gapp.accountMapper.GetAccount(ctxCheck, addr2)
	require.Equal(t, acc1, res1)
	require.Equal(t, acc2, res2)

	// Declare Candidacy

	description := stake.NewDescription("foo_moniker", "", "", "")
	declareCandidacyMsg := stake.NewMsgDeclareCandidacy(
		addr1, priv1.PubKey(), bondCoin, description,
	)
	SignCheckDeliver(t, gapp, declareCandidacyMsg, []int64{0}, true, priv1)

	ctxDeliver := gapp.BaseApp.NewContext(false, abci.Header{})
	res1 = gapp.accountMapper.GetAccount(ctxDeliver, addr1)
	require.Equal(t, genCoins.Minus(sdk.Coins{bondCoin}), res1.GetCoins())
	candidate, found := gapp.stakeKeeper.GetCandidate(ctxDeliver, addr1)
	require.True(t, found)
	require.Equal(t, candidate.Address, addr1)

	// Edit Candidacy

	description = stake.NewDescription("bar_moniker", "", "", "")
	editCandidacyMsg := stake.NewMsgEditCandidacy(
		addr1, description,
	)
	SignDeliver(t, gapp, editCandidacyMsg, []int64{1}, true, priv1)

	candidate, found = gapp.stakeKeeper.GetCandidate(ctxDeliver, addr1)
	require.True(t, found)
	require.Equal(t, candidate.Description, description)

	// Delegate

	delegateMsg := stake.NewMsgDelegate(
		addr2, addr1, bondCoin,
	)
	SignDeliver(t, gapp, delegateMsg, []int64{0}, true, priv2)

	ctxDeliver = gapp.BaseApp.NewContext(false, abci.Header{})
	res2 = gapp.accountMapper.GetAccount(ctxDeliver, addr2)
	require.Equal(t, genCoins.Minus(sdk.Coins{bondCoin}), res2.GetCoins())
	bond, found := gapp.stakeKeeper.GetDelegatorBond(ctxDeliver, addr2, addr1)
	require.True(t, found)
	require.Equal(t, bond.DelegatorAddr, addr2)

	// Unbond

	unbondMsg := stake.NewMsgUnbond(
		addr2, addr1, "MAX",
	)
	SignDeliver(t, gapp, unbondMsg, []int64{1}, true, priv2)

	ctxDeliver = gapp.BaseApp.NewContext(false, abci.Header{})
	res2 = gapp.accountMapper.GetAccount(ctxDeliver, addr2)
	require.Equal(t, genCoins, res2.GetCoins())
	_, found = gapp.stakeKeeper.GetDelegatorBond(ctxDeliver, addr2, addr1)
	require.False(t, found)
}

//____________________________________________________________________________________

func CheckBalance(t *testing.T, gapp *GaiaApp, addr sdk.Address, balExpected string) {
	ctxDeliver := gapp.BaseApp.NewContext(false, abci.Header{})
	res2 := gapp.accountMapper.GetAccount(ctxDeliver, addr)
	assert.Equal(t, balExpected, fmt.Sprintf("%v", res2.GetCoins()))
}

func genTx(msg sdk.Msg, seq []int64, priv ...crypto.PrivKeyEd25519) sdk.StdTx {
	sigs := make([]sdk.StdSignature, len(priv))
	for i, p := range priv {
		sigs[i] = sdk.StdSignature{
			PubKey:    p.PubKey(),
			Signature: p.Sign(sdk.StdSignBytes(chainID, seq, fee, msg)),
			Sequence:  seq[i],
		}
	}

	return sdk.NewStdTx(msg, fee, sigs)

}

func SignCheckDeliver(t *testing.T, gapp *GaiaApp, msg sdk.Msg, seq []int64, expPass bool, priv ...crypto.PrivKeyEd25519) {

	// Sign the tx
	tx := genTx(msg, seq, priv...)

	// Run a Check
	res := gapp.Check(tx)
	if expPass {
		require.Equal(t, sdk.CodeOK, res.Code, res.Log)
	} else {
		require.NotEqual(t, sdk.CodeOK, res.Code, res.Log)
	}

	// Simulate a Block
	gapp.BeginBlock(abci.RequestBeginBlock{})
	res = gapp.Deliver(tx)
	if expPass {
		require.Equal(t, sdk.CodeOK, res.Code, res.Log)
	} else {
		require.NotEqual(t, sdk.CodeOK, res.Code, res.Log)
	}
	gapp.EndBlock(abci.RequestEndBlock{})

	// XXX fix code or add explaination as to why using commit breaks a bunch of these tests
	//gapp.Commit()
}

// XXX the only reason we are using Sign Deliver here is because the tests
// break on check tx the second time you use SignCheckDeliver in a test because
// the checktx state has not been updated likely because commit is not being
// called!
func SignDeliver(t *testing.T, gapp *GaiaApp, msg sdk.Msg, seq []int64, expPass bool, priv ...crypto.PrivKeyEd25519) {

	// Sign the tx
	tx := genTx(msg, seq, priv...)

	// Simulate a Block
	gapp.BeginBlock(abci.RequestBeginBlock{})
	res := gapp.Deliver(tx)
	if expPass {
		require.Equal(t, sdk.CodeOK, res.Code, res.Log)
	} else {
		require.NotEqual(t, sdk.CodeOK, res.Code, res.Log)
	}
	gapp.EndBlock(abci.RequestEndBlock{})
}
