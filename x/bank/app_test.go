package bank

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/mock"

	"github.com/stretchr/testify/require"

	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/crypto"
	"github.com/tendermint/tendermint/crypto/ed25519"
)

type (
	expectedBalance struct {
		addr  sdk.AccAddress
		coins sdk.Coins
	}

	appTestCase struct {
		expPass          bool
		msgs             []sdk.Msg
		accNums          []int64
		accSeqs          []int64
		privKeys         []crypto.PrivKey
		expectedBalances []expectedBalance
	}
)

var (
	priv1 = ed25519.GenPrivKey()
	addr1 = sdk.AccAddress(priv1.PubKey().Address())
	priv2 = ed25519.GenPrivKey()
	addr2 = sdk.AccAddress(priv2.PubKey().Address())
	addr3 = sdk.AccAddress(ed25519.GenPrivKey().PubKey().Address())
	priv4 = ed25519.GenPrivKey()
	addr4 = sdk.AccAddress(priv4.PubKey().Address())

	coins     = sdk.Coins{sdk.NewInt64Coin("foocoin", 10)}
	halfCoins = sdk.Coins{sdk.NewInt64Coin("foocoin", 5)}
	manyCoins = sdk.Coins{sdk.NewInt64Coin("foocoin", 1), sdk.NewInt64Coin("barcoin", 1)}
	freeFee   = auth.NewStdFee(100000, sdk.Coins{sdk.NewInt64Coin("foocoin", 0)}...)

	sendMsg1 = MsgSend{
		Inputs:  []Input{NewInput(addr1, coins)},
		Outputs: []Output{NewOutput(addr2, coins)},
	}
	sendMsg2 = MsgSend{
		Inputs: []Input{NewInput(addr1, coins)},
		Outputs: []Output{
			NewOutput(addr2, halfCoins),
			NewOutput(addr3, halfCoins),
		},
	}
	sendMsg3 = MsgSend{
		Inputs: []Input{
			NewInput(addr1, coins),
			NewInput(addr4, coins),
		},
		Outputs: []Output{
			NewOutput(addr2, coins),
			NewOutput(addr3, coins),
		},
	}
	sendMsg4 = MsgSend{
		Inputs: []Input{
			NewInput(addr2, coins),
		},
		Outputs: []Output{
			NewOutput(addr1, coins),
		},
	}
	sendMsg5 = MsgSend{
		Inputs: []Input{
			NewInput(addr1, manyCoins),
		},
		Outputs: []Output{
			NewOutput(addr2, manyCoins),
		},
	}
)

// initialize the mock application for this module
func getMockApp(t *testing.T) *mock.App {
	mapp, err := getBenchmarkMockApp()
	require.NoError(t, err)
	return mapp
}

func TestMsgSendWithAccounts(t *testing.T) {
	mapp := getMockApp(t)
	acc := &auth.BaseAccount{
		Address: addr1,
		Coins:   sdk.Coins{sdk.NewInt64Coin("foocoin", 67)},
	}

	mock.SetGenesis(mapp, []auth.Account{acc})

	ctxCheck := mapp.BaseApp.NewContext(true, abci.Header{})

	res1 := mapp.AccountMapper.GetAccount(ctxCheck, addr1)
	require.NotNil(t, res1)
	require.Equal(t, acc, res1.(*auth.BaseAccount))

	testCases := []appTestCase{
		{
			msgs:     []sdk.Msg{sendMsg1},
			accNums:  []int64{0},
			accSeqs:  []int64{0},
			expPass:  true,
			privKeys: []crypto.PrivKey{priv1},
			expectedBalances: []expectedBalance{
				{addr1, sdk.Coins{sdk.NewInt64Coin("foocoin", 57)}},
				{addr2, sdk.Coins{sdk.NewInt64Coin("foocoin", 10)}},
			},
		},
		{
			msgs:     []sdk.Msg{sendMsg1, sendMsg2},
			accNums:  []int64{0},
			accSeqs:  []int64{0},
			expPass:  false,
			privKeys: []crypto.PrivKey{priv1},
		},
	}

	for _, tc := range testCases {
		mock.SignCheckDeliver(t, mapp.BaseApp, tc.msgs, tc.accNums, tc.accSeqs, tc.expPass, tc.privKeys...)

		for _, eb := range tc.expectedBalances {
			mock.CheckBalance(t, mapp, eb.addr, eb.coins)
		}
	}

	// bumping the tx nonce number without resigning should be an auth error
	mapp.BeginBlock(abci.RequestBeginBlock{})

	tx := mock.GenTx([]sdk.Msg{sendMsg1}, []int64{0}, []int64{0}, priv1)
	tx.Signatures[0].Sequence = 1

	res := mapp.Deliver(tx)
	require.Equal(t, sdk.ToABCICode(sdk.CodespaceRoot, sdk.CodeUnauthorized), res.Code, res.Log)

	// resigning the tx with the bumped sequence should work
	mock.SignCheckDeliver(t, mapp.BaseApp, []sdk.Msg{sendMsg1, sendMsg2}, []int64{0}, []int64{1}, true, priv1)
}

func TestMsgSendMultipleOut(t *testing.T) {
	mapp := getMockApp(t)

	acc1 := &auth.BaseAccount{
		Address: addr1,
		Coins:   sdk.Coins{sdk.NewInt64Coin("foocoin", 42)},
	}
	acc2 := &auth.BaseAccount{
		Address: addr2,
		Coins:   sdk.Coins{sdk.NewInt64Coin("foocoin", 42)},
	}

	mock.SetGenesis(mapp, []auth.Account{acc1, acc2})

	testCases := []appTestCase{
		{
			msgs:     []sdk.Msg{sendMsg2},
			accNums:  []int64{0},
			accSeqs:  []int64{0},
			expPass:  true,
			privKeys: []crypto.PrivKey{priv1},
			expectedBalances: []expectedBalance{
				{addr1, sdk.Coins{sdk.NewInt64Coin("foocoin", 32)}},
				{addr2, sdk.Coins{sdk.NewInt64Coin("foocoin", 47)}},
				{addr3, sdk.Coins{sdk.NewInt64Coin("foocoin", 5)}},
			},
		},
	}

	for _, tc := range testCases {
		mock.SignCheckDeliver(t, mapp.BaseApp, tc.msgs, tc.accNums, tc.accSeqs, tc.expPass, tc.privKeys...)

		for _, eb := range tc.expectedBalances {
			mock.CheckBalance(t, mapp, eb.addr, eb.coins)
		}
	}
}

func TestSengMsgMultipleInOut(t *testing.T) {
	mapp := getMockApp(t)

	acc1 := &auth.BaseAccount{
		Address: addr1,
		Coins:   sdk.Coins{sdk.NewInt64Coin("foocoin", 42)},
	}
	acc2 := &auth.BaseAccount{
		Address: addr2,
		Coins:   sdk.Coins{sdk.NewInt64Coin("foocoin", 42)},
	}
	acc4 := &auth.BaseAccount{
		Address: addr4,
		Coins:   sdk.Coins{sdk.NewInt64Coin("foocoin", 42)},
	}

	mock.SetGenesis(mapp, []auth.Account{acc1, acc2, acc4})

	testCases := []appTestCase{
		{
			msgs:     []sdk.Msg{sendMsg3},
			accNums:  []int64{0, 2},
			accSeqs:  []int64{0, 0},
			expPass:  true,
			privKeys: []crypto.PrivKey{priv1, priv4},
			expectedBalances: []expectedBalance{
				{addr1, sdk.Coins{sdk.NewInt64Coin("foocoin", 32)}},
				{addr4, sdk.Coins{sdk.NewInt64Coin("foocoin", 32)}},
				{addr2, sdk.Coins{sdk.NewInt64Coin("foocoin", 52)}},
				{addr3, sdk.Coins{sdk.NewInt64Coin("foocoin", 10)}},
			},
		},
	}

	for _, tc := range testCases {
		mock.SignCheckDeliver(t, mapp.BaseApp, tc.msgs, tc.accNums, tc.accSeqs, tc.expPass, tc.privKeys...)

		for _, eb := range tc.expectedBalances {
			mock.CheckBalance(t, mapp, eb.addr, eb.coins)
		}
	}
}

func TestMsgSendDependent(t *testing.T) {
	mapp := getMockApp(t)

	acc1 := &auth.BaseAccount{
		Address: addr1,
		Coins:   sdk.Coins{sdk.NewInt64Coin("foocoin", 42)},
	}

	mock.SetGenesis(mapp, []auth.Account{acc1})

	testCases := []appTestCase{
		{
			msgs:     []sdk.Msg{sendMsg1},
			accNums:  []int64{0},
			accSeqs:  []int64{0},
			expPass:  true,
			privKeys: []crypto.PrivKey{priv1},
			expectedBalances: []expectedBalance{
				{addr1, sdk.Coins{sdk.NewInt64Coin("foocoin", 32)}},
				{addr2, sdk.Coins{sdk.NewInt64Coin("foocoin", 10)}},
			},
		},
		{
			msgs:     []sdk.Msg{sendMsg4},
			accNums:  []int64{1},
			accSeqs:  []int64{0},
			expPass:  true,
			privKeys: []crypto.PrivKey{priv2},
			expectedBalances: []expectedBalance{
				{addr1, sdk.Coins{sdk.NewInt64Coin("foocoin", 42)}},
			},
		},
	}

	for _, tc := range testCases {
		mock.SignCheckDeliver(t, mapp.BaseApp, tc.msgs, tc.accNums, tc.accSeqs, tc.expPass, tc.privKeys...)

		for _, eb := range tc.expectedBalances {
			mock.CheckBalance(t, mapp, eb.addr, eb.coins)
		}
	}
}
