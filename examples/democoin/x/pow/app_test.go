package pow

import (
	"testing"

	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/bank"
	"github.com/cosmos/cosmos-sdk/x/mock"

	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/crypto/ed25519"
)

var (
	priv1 = ed25519.GenPrivKey()
	addr1 = sdk.AccAddress(priv1.PubKey().Address())
)

// initialize the mock application for this module
func getMockApp(t *testing.T) *mock.App {
	mapp := mock.NewApp()

	RegisterWire(mapp.Cdc)
	keyPOW := sdk.NewKVStoreKey("pow")
	coinKeeper := bank.NewKeeper(mapp.AccountMapper)
	config := Config{"pow", 1}
	keeper := NewKeeper(keyPOW, config, coinKeeper, mapp.RegisterCodespace(DefaultCodespace))
	mapp.Router().AddRoute("pow", keeper.Handler)

	mapp.SetInitChainer(getInitChainer(mapp, keeper))

	require.NoError(t, mapp.CompleteSetup([]*sdk.KVStoreKey{keyPOW}))

	mapp.Seal()

	return mapp
}

// overwrite the mock init chainer
func getInitChainer(mapp *mock.App, keeper Keeper) sdk.InitChainer {
	return func(ctx sdk.Context, req abci.RequestInitChain) abci.ResponseInitChain {
		mapp.InitChainer(ctx, req)

		genesis := Genesis{
			Difficulty: 1,
			Count:      0,
		}
		InitGenesis(ctx, keeper, genesis)

		return abci.ResponseInitChain{}
	}
}

func TestMsgMine(t *testing.T) {
	mapp := getMockApp(t)

	// Construct genesis state
	acc1 := &auth.BaseAccount{
		Address: addr1,
		Coins:   nil,
	}
	accs := []auth.Account{acc1}

	// Initialize the chain (nil)
	mock.SetGenesis(mapp, accs)

	// A checkTx context (true)
	ctxCheck := mapp.BaseApp.NewContext(true, abci.Header{})
	res1 := mapp.AccountMapper.GetAccount(ctxCheck, addr1)
	require.Equal(t, acc1, res1)

	// Mine and check for reward
	mineMsg1 := GenerateMsgMine(addr1, 1, 2)
	mock.SignCheckDeliver(t, mapp.BaseApp, []sdk.Msg{mineMsg1}, []int64{0}, []int64{0}, true, priv1)
	mock.CheckBalance(t, mapp, addr1, sdk.Coins{{"pow", sdk.NewInt(1)}})
	// Mine again and check for reward
	mineMsg2 := GenerateMsgMine(addr1, 2, 3)
	mock.SignCheckDeliver(t, mapp.BaseApp, []sdk.Msg{mineMsg2}, []int64{0}, []int64{1}, true, priv1)
	mock.CheckBalance(t, mapp, addr1, sdk.Coins{{"pow", sdk.NewInt(2)}})
	// Mine again - should be invalid
	mock.SignCheckDeliver(t, mapp.BaseApp, []sdk.Msg{mineMsg2}, []int64{0}, []int64{1}, false, priv1)
	mock.CheckBalance(t, mapp, addr1, sdk.Coins{{"pow", sdk.NewInt(2)}})
}
