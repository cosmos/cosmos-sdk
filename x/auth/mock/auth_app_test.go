package mock

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/bank"

	abci "github.com/tendermint/abci/types"
	crypto "github.com/tendermint/go-crypto"
)

// test auth module messages

var (
	priv1 = crypto.GenPrivKeyEd25519()
	addr1 = priv1.PubKey().Address()
	priv2 = crypto.GenPrivKeyEd25519()
	addr2 = priv2.PubKey().Address()

	coins    = sdk.Coins{sdk.NewCoin("foocoin", 10)}
	sendMsg1 = bank.MsgSend{
		Inputs:  []bank.Input{bank.NewInput(addr1, coins)},
		Outputs: []bank.Output{bank.NewOutput(addr2, coins)},
	}
)

// initialize the mock application for this module
func getMockApp(t *testing.T) *App {
	mapp := NewApp()

	coinKeeper := bank.NewKeeper(mapp.AccountMapper)
	mapp.Router().AddRoute("bank", bank.NewHandler(coinKeeper))
	mapp.Router().AddRoute("auth", auth.NewHandler(mapp.AccountMapper))

	require.NoError(t, mapp.CompleteSetup([]*sdk.KVStoreKey{}))
	return mapp
}

func TestMsgChangePubKey(t *testing.T) {
	mapp := getMockApp(t)

	// Construct some genesis bytes to reflect basecoin/types/AppAccount
	// Give 77 foocoin to the first key
	coins := sdk.Coins{sdk.NewCoin("foocoin", 77)}
	acc1 := &auth.BaseAccount{
		Address: addr1,
		Coins:   coins,
	}
	accs := []auth.Account{acc1}

	// Construct genesis state
	SetGenesis(mapp, accs)

	// A checkTx context (true)
	ctxCheck := mapp.BaseApp.NewContext(true, abci.Header{})
	res1 := mapp.AccountMapper.GetAccount(ctxCheck, addr1)
	assert.Equal(t, acc1, res1.(*auth.BaseAccount))

	// Run a CheckDeliver
	SignCheckDeliver(t, mapp.BaseApp, []sdk.Msg{sendMsg1}, []int64{0}, []int64{0}, true, priv1)

	// Check balances
	CheckBalance(t, mapp, addr1, sdk.Coins{sdk.NewCoin("foocoin", 67)})
	CheckBalance(t, mapp, addr2, sdk.Coins{sdk.NewCoin("foocoin", 10)})

	changePubKeyMsg := auth.MsgChangeKey{
		Address:   addr1,
		NewPubKey: priv2.PubKey(),
	}

	mapp.BeginBlock(abci.RequestBeginBlock{})
	ctxDeliver := mapp.BaseApp.NewContext(false, abci.Header{})
	acc2 := mapp.AccountMapper.GetAccount(ctxDeliver, addr1)

	// send a MsgChangePubKey
	SignCheckDeliver(t, mapp.BaseApp, []sdk.Msg{changePubKeyMsg}, []int64{0}, []int64{1}, true, priv1)
	acc2 = mapp.AccountMapper.GetAccount(ctxDeliver, addr1)

	assert.True(t, priv2.PubKey().Equals(acc2.GetPubKey()))

	// signing a SendMsg with the old privKey should be an auth error
	mapp.BeginBlock(abci.RequestBeginBlock{})
	tx := GenTx([]sdk.Msg{sendMsg1}, []int64{0}, []int64{2}, priv1)
	res := mapp.Deliver(tx)
	assert.Equal(t, sdk.ToABCICode(sdk.CodespaceRoot, sdk.CodeUnauthorized), res.Code, res.Log)

	// resigning the tx with the new correct priv key should work
	SignCheckDeliver(t, mapp.BaseApp, []sdk.Msg{sendMsg1}, []int64{0}, []int64{2}, true, priv2)

	// Check balances
	CheckBalance(t, mapp, addr1, sdk.Coins{sdk.NewCoin("foocoin", 57)})
	CheckBalance(t, mapp, addr2, sdk.Coins{sdk.NewCoin("foocoin", 20)})
}
