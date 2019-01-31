package ibc

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

// initialize the mock application for this module
func getMockApp(t *testing.T) *mock.App {
	mapp := mock.NewApp()

	RegisterCodec(mapp.Cdc)
	keyIBC := sdk.NewKVStoreKey("ibc")
	ibcMapper := NewMapper(mapp.Cdc, keyIBC, DefaultCodespace)
	bankKeeper := bank.NewBaseKeeper(mapp.AccountKeeper,
		mapp.ParamsKeeper.Subspace(bank.DefaultParamspace),
		bank.DefaultCodespace)
	mapp.Router().AddRoute("ibc", NewHandler(ibcMapper, bankKeeper))

	require.NoError(t, mapp.CompleteSetup(keyIBC))
	return mapp
}

func TestIBCMsgs(t *testing.T) {
	mapp := getMockApp(t)

	sourceChain := "source-chain"
	destChain := "dest-chain"

	priv1 := ed25519.GenPrivKey()
	addr1 := sdk.AccAddress(priv1.PubKey().Address())
	coins := sdk.Coins{sdk.NewInt64Coin("foocoin", 10)}
	var emptyCoins sdk.Coins

	acc := &auth.BaseAccount{
		Address: addr1,
		Coins:   coins,
	}
	accs := []auth.Account{acc}

	mock.SetGenesis(mapp, accs)

	// A checkTx context (true)
	ctxCheck := mapp.BaseApp.NewContext(true, abci.Header{})
	res1 := mapp.AccountKeeper.GetAccount(ctxCheck, addr1)
	require.Equal(t, acc, res1)

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

	mock.SignCheckDeliver(t, mapp.Cdc, mapp.BaseApp, []sdk.Msg{transferMsg}, []uint64{0}, []uint64{0}, true, true, priv1)
	mock.CheckBalance(t, mapp, addr1, emptyCoins)
	mock.SignCheckDeliver(t, mapp.Cdc, mapp.BaseApp, []sdk.Msg{transferMsg}, []uint64{0}, []uint64{1}, false, false, priv1)
	mock.SignCheckDeliver(t, mapp.Cdc, mapp.BaseApp, []sdk.Msg{receiveMsg}, []uint64{0}, []uint64{2}, true, true, priv1)
	mock.CheckBalance(t, mapp, addr1, coins)
	mock.SignCheckDeliver(t, mapp.Cdc, mapp.BaseApp, []sdk.Msg{receiveMsg}, []uint64{0}, []uint64{2}, false, false, priv1)
}
