package ibc

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

// initialize the mock application for this module
func getMockApp(t *testing.T) *mock.App {
	mapp := mock.NewApp()

	RegisterWire(mapp.Cdc)
	keyIBC := sdk.NewKVStoreKey("ibc")
	ibcMapper := NewMapper(mapp.Cdc, keyIBC, mapp.RegisterCodespace(DefaultCodespace))
	coinKeeper := bank.NewKeeper(mapp.AccountMapper)
	mapp.Router().AddRoute("ibc", NewHandler(ibcMapper, coinKeeper))

	mapp.CompleteSetup(t, []*sdk.KVStoreKey{keyIBC})
	return mapp
}

func TestIBCMsgs(t *testing.T) {
	mapp := getMockApp(t)

	sourceChain := "source-chain"
	destChain := "dest-chain"

	priv1 := crypto.GenPrivKeyEd25519()
	addr1 := priv1.PubKey().Address()
	coins := sdk.Coins{{"foocoin", 10}}
	var emptyCoins sdk.Coins

	acc := &auth.BaseAccount{
		Address: addr1,
		Coins:   coins,
	}
	accs := []auth.Account{acc}

	mock.SetGenesis(mapp, accs)

	// A checkTx context (true)
	ctxCheck := mapp.BaseApp.NewContext(true, abci.Header{})
	res1 := mapp.AccountMapper.GetAccount(ctxCheck, addr1)
	assert.Equal(t, acc, res1)

	packet := IBCPacket{
		SrcAddr:   addr1,
		DestAddr:  addr1,
		Coins:     coins,
		SrcChain:  sourceChain,
		DestChain: destChain,
	}

	transferMsg := IBCTransferMsg{
		IBCPacket: packet,
	}

	receiveMsg := IBCReceiveMsg{
		IBCPacket: packet,
		Relayer:   addr1,
		Sequence:  0,
	}

	mock.SignCheckDeliver(t, mapp.BaseApp, transferMsg, []int64{0},[]int64{0}, true, priv1)
	mock.CheckBalance(t, mapp, addr1, emptyCoins)
	mock.SignCheckDeliver(t, mapp.BaseApp, transferMsg, []int64{0}, []int64{1}, false, priv1)
	mock.SignCheckDeliver(t, mapp.BaseApp, receiveMsg, []int64{0}, []int64{2}, true, priv1)
	mock.CheckBalance(t, mapp, addr1, coins)
	mock.SignCheckDeliver(t, mapp.BaseApp, receiveMsg, []int64{0}, []int64{3}, false, priv1)
}
