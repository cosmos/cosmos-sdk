package pow

import (
	"testing"

	"github.com/stretchr/testify/assert"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/auth/mock"
	"github.com/cosmos/cosmos-sdk/x/bank"

	abci "github.com/tendermint/abci/types"
	crypto "github.com/tendermint/go-crypto"
)

var (
	priv1 = crypto.GenPrivKeyEd25519()
	addr1 = priv1.PubKey().Address()
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

	mapp.CompleteSetup(t, []*sdk.KVStoreKey{keyPOW})
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
	assert.Equal(t, acc1, res1)

	// Mine and check for reward
	mineMsg1 := GenerateMsgMine(addr1, 1, 2)
	mock.SignCheckDeliver(t, mapp.BaseApp, mineMsg1, []int64{0}, []int64{0}, true, priv1)
	mock.CheckBalance(t, mapp, addr1, sdk.Coins{{"pow", 1}})
	// Mine again and check for reward
	mineMsg2 := GenerateMsgMine(addr1, 2, 3)
	mock.SignCheckDeliver(t, mapp.BaseApp, mineMsg2, []int64{0}, []int64{1}, true, priv1)
	mock.CheckBalance(t, mapp, addr1, sdk.Coins{{"pow", 2}})
	// Mine again - should be invalid
	mock.SignCheckDeliver(t, mapp.BaseApp, mineMsg2, []int64{0}, []int64{1}, false, priv1)
	mock.CheckBalance(t, mapp, addr1, sdk.Coins{{"pow", 2}})
}
