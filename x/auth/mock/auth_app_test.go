package mock

import (
	"testing"

	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/bank"

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

	require.NoError(t, mapp.CompleteSetup([]*sdk.KVStoreKey{}))
	return mapp
}
