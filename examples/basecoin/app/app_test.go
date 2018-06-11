package app

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/examples/basecoin/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/wire"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/stake"

	abci "github.com/tendermint/abci/types"
	crypto "github.com/tendermint/go-crypto"
	dbm "github.com/tendermint/tmlibs/db"
	"github.com/tendermint/tmlibs/log"
)

func setGenesis(bapp *BasecoinApp, accs ...auth.BaseAccount) error {
	genaccs := make([]*types.GenesisAccount, len(accs))
	for i, acc := range accs {
		genaccs[i] = types.NewGenesisAccount(&types.AppAccount{acc, "foobart"})
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

//_______________________________________________________________________

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
<<<<<<< HEAD

func TestMsgChangePubKey(t *testing.T) {

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
	err = setGenesis(bapp, baseAcc)
	require.Nil(t, err)

	// A checkTx context (true)
	ctxCheck := bapp.BaseApp.NewContext(true, abci.Header{})
	res1 := bapp.accountMapper.GetAccount(ctxCheck, addr1)
	fmt.Println("acct1: ", res1)
	assert.Equal(t, baseAcc, res1.(*types.AppAccount).BaseAccount)

	// CheckBalance(t, bapp, addr1, "77foocoin") //get an error, prob cuz the account doesnt exist yet in the state

	// Run a CheckDeliver
	SignCheckDeliver(t, bapp, sendMsg1, []int64{0}, true, priv1)

	// Check balances
	CheckBalance(t, bapp, addr1, "67foocoin")
	CheckBalance(t, bapp, addr2, "10foocoin")

	changePubKeyMsg := auth.MsgChangeKey{
		Address:   addr1,
		NewPubKey: priv2.PubKey(),
	}
	fmt.Println("changePubKeyMsg: ", changePubKeyMsg)

	ctxDeliver := bapp.BaseApp.NewContext(false, abci.Header{})
	acc := bapp.accountMapper.GetAccount(ctxDeliver, addr1)
	fmt.Println("acct1.5: ", acc)

	// send a MsgChangePubKey
	SignCheckDeliver(t, bapp, changePubKeyMsg, []int64{1}, true, priv1)
	acc = bapp.accountMapper.GetAccount(ctxDeliver, addr1)
	fmt.Println("acct2: ", acc)

	assert.True(t, priv2.PubKey().Equals(acc.GetPubKey()))

	// signing a SendMsg with the old privKey should be an auth error
	tx := genTx(sendMsg1, []int64{2}, priv1)
	res := bapp.Deliver(tx) //FAIL
	assert.Equal(t, sdk.ToABCICode(sdk.CodespaceRoot, sdk.CodeUnauthorized), res.Code, res.Log)

	// resigning the tx with the new correct priv key should work
	SignCheckDeliver(t, bapp, sendMsg1, []int64{2}, true, priv2) //FAIL

	// Check balances
	CheckBalance(t, bapp, addr1, "57foocoin")
	CheckBalance(t, bapp, addr2, "20foocoin")
}

func TestMsgSendWithAccounts(t *testing.T) {
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
	err = setGenesis(bapp, baseAcc)
	require.Nil(t, err)

	// A checkTx context (true)
	ctxCheck := bapp.BaseApp.NewContext(true, abci.Header{})
	res1 := bapp.accountMapper.GetAccount(ctxCheck, addr1)
	assert.Equal(t, baseAcc, res1.(*types.AppAccount).BaseAccount)

	// Run a CheckDeliver
	SignCheckDeliver(t, bapp, sendMsg1, []int64{0}, true, priv1)

	// Check balances
	CheckBalance(t, bapp, addr1, "67foocoin")
	CheckBalance(t, bapp, addr2, "10foocoin")

	// Delivering again should cause replay error
	SignCheckDeliver(t, bapp, sendMsg1, []int64{0}, false, priv1)

	// bumping the txnonce number without resigning should be an auth error
	tx := genTx(sendMsg1, []int64{0}, priv1)
	tx.Signatures[0].Sequence = 1
	res := bapp.Deliver(tx)

	assert.Equal(t, sdk.ToABCICode(sdk.CodespaceRoot, sdk.CodeUnauthorized), res.Code, res.Log)

	// resigning the tx with the bumped sequence should work
	SignCheckDeliver(t, bapp, sendMsg1, []int64{1}, true, priv1)
}

func TestMsgSendMultipleOut(t *testing.T) {
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

	// Construct genesis state
	err = setGenesis(bapp, acc1, acc2)
	require.Nil(t, err)

	// Simulate a Block
	SignCheckDeliver(t, bapp, sendMsg2, []int64{0}, true, priv1)

	// Check balances
	CheckBalance(t, bapp, addr1, "32foocoin")
	CheckBalance(t, bapp, addr2, "47foocoin")
	CheckBalance(t, bapp, addr3, "5foocoin")
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

	err = setGenesis(bapp, acc1, acc2, acc4)
	assert.Nil(t, err)

	// CheckDeliver
	SignCheckDeliver(t, bapp, sendMsg3, []int64{0, 0}, true, priv1, priv4)

	// Check balances
	CheckBalance(t, bapp, addr1, "32foocoin")
	CheckBalance(t, bapp, addr4, "32foocoin")
	CheckBalance(t, bapp, addr2, "52foocoin")
	CheckBalance(t, bapp, addr3, "10foocoin")
}

func TestMsgSendDependent(t *testing.T) {
	bapp := newBasecoinApp()

	genCoins, err := sdk.ParseCoins("42foocoin")
	require.Nil(t, err)

	acc1 := auth.BaseAccount{
		Address: addr1,
		Coins:   genCoins,
	}

	// Construct genesis state
	err = setGenesis(bapp, acc1)
	require.Nil(t, err)

	err = setGenesis(bapp, acc1)
	assert.Nil(t, err)

	// CheckDeliver
	SignCheckDeliver(t, bapp, sendMsg1, []int64{0}, true, priv1)

	// Check balances
	CheckBalance(t, bapp, addr1, "32foocoin")
	CheckBalance(t, bapp, addr2, "10foocoin")

	// Simulate a Block
	SignCheckDeliver(t, bapp, sendMsg4, []int64{0}, true, priv2)

	// Check balances
	CheckBalance(t, bapp, addr1, "42foocoin")
}

func TestMsgQuiz(t *testing.T) {
	bapp := newBasecoinApp()

	// Construct genesis state
	// Construct some genesis bytes to reflect basecoin/types/AppAccount
	baseAcc := auth.BaseAccount{
		Address: addr1,
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
	res1 := bapp.accountMapper.GetAccount(ctxCheck, addr1)
	assert.Equal(t, acc1, res1)

}

func TestIBCMsgs(t *testing.T) {
	bapp := newBasecoinApp()

	sourceChain := "source-chain"
	destChain := "dest-chain"

	baseAcc := auth.BaseAccount{
		Address: addr1,
		Coins:   coins,
	}
	acc1 := &types.AppAccount{baseAcc, "foobart"}

	err := setGenesis(bapp, baseAcc)
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

	SignCheckDeliver(t, bapp, transferMsg, []int64{0}, true, priv1)
	CheckBalance(t, bapp, addr1, "")
	SignCheckDeliver(t, bapp, transferMsg, []int64{1}, false, priv1)
	SignCheckDeliver(t, bapp, receiveMsg, []int64{2}, true, priv1)
	CheckBalance(t, bapp, addr1, "10foocoin")
	SignCheckDeliver(t, bapp, receiveMsg, []int64{3}, false, priv1)
}

func genTx(msg sdk.Msg, seq []int64, priv ...crypto.PrivKeyEd25519) auth.StdTx {
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

func SignCheckDeliver(t *testing.T, bapp *BasecoinApp, msg sdk.Msg, seq []int64, expPass bool, priv ...crypto.PrivKeyEd25519) {

	// Sign the tx
	tx := genTx(msg, seq, priv...)
	// Run a Check

	res := bapp.Check(tx)
	fmt.Println("res: ", res)
	fmt.Println("GO")

	if expPass {
		require.Equal(t, sdk.ABCICodeOK, res.Code, res.Log)
	} else {
		require.NotEqual(t, sdk.ABCICodeOK, res.Code, res.Log)
	}

	// Simulate a Block
	bapp.BeginBlock(abci.RequestBeginBlock{})
	res = bapp.Deliver(tx)
	if expPass {
		require.Equal(t, sdk.ABCICodeOK, res.Code, res.Log)
	} else {
		require.NotEqual(t, sdk.ABCICodeOK, res.Code, res.Log)
	}
	bapp.EndBlock(abci.RequestEndBlock{})
	//bapp.Commit()
}

func CheckBalance(t *testing.T, bapp *BasecoinApp, addr sdk.Address, balExpected string) {
	ctxDeliver := bapp.BaseApp.NewContext(false, abci.Header{})
	res2 := bapp.accountMapper.GetAccount(ctxDeliver, addr)
	assert.Equal(t, balExpected, fmt.Sprintf("%v", res2.GetCoins()))
}
=======
>>>>>>> fc0e4013278d41fab4f3ac73f28a42bc45889106
