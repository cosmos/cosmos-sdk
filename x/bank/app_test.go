package bank_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/simapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/bank/types"
)

type (
	expectedBalance struct {
		addr  sdk.AccAddress
		coins sdk.Coins
	}

	appTestCase struct {
		desc             string
		expSimPass       bool
		expPass          bool
		msgs             []sdk.Msg
		accNums          []uint64
		accSeqs          []uint64
		privKeys         []cryptotypes.PrivKey
		expectedBalances []expectedBalance
	}
)

var (
	priv1 = secp256k1.GenPrivKey()
	addr1 = sdk.AccAddress(priv1.PubKey().Address())
	priv2 = secp256k1.GenPrivKey()
	addr2 = sdk.AccAddress(priv2.PubKey().Address())
	addr3 = sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())
	priv4 = secp256k1.GenPrivKey()
	addr4 = sdk.AccAddress(priv4.PubKey().Address())

	coins     = sdk.Coins{sdk.NewInt64Coin("foocoin", 10)}
	halfCoins = sdk.Coins{sdk.NewInt64Coin("foocoin", 5)}

	sendMsg1 = types.NewMsgSend(addr1, addr2, coins)

	multiSendMsg1 = &types.MsgMultiSend{
		Inputs:  []types.Input{types.NewInput(addr1, coins)},
		Outputs: []types.Output{types.NewOutput(addr2, coins)},
	}
	multiSendMsg2 = &types.MsgMultiSend{
		Inputs: []types.Input{types.NewInput(addr1, coins)},
		Outputs: []types.Output{
			types.NewOutput(addr2, halfCoins),
			types.NewOutput(addr3, halfCoins),
		},
	}
	multiSendMsg3 = &types.MsgMultiSend{
		Inputs: []types.Input{
			types.NewInput(addr1, coins),
			types.NewInput(addr4, coins),
		},
		Outputs: []types.Output{
			types.NewOutput(addr2, coins),
			types.NewOutput(addr3, coins),
		},
	}
	multiSendMsg4 = &types.MsgMultiSend{
		Inputs: []types.Input{
			types.NewInput(addr2, coins),
		},
		Outputs: []types.Output{
			types.NewOutput(addr1, coins),
		},
	}
	multiSendMsg5 = &types.MsgMultiSend{
		Inputs: []types.Input{
			types.NewInput(addr1, coins),
		},
		Outputs: []types.Output{
			types.NewOutput(moduleAccAddr, coins),
		},
	}
)

func TestSendNotEnoughBalance(t *testing.T) {
	acc := &authtypes.BaseAccount{
		Address: addr1.String(),
	}

	genAccs := []authtypes.GenesisAccount{acc}
	app := simapp.SetupWithGenesisAccounts(genAccs)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})

	err := app.BankKeeper.SetBalances(ctx, addr1, sdk.NewCoins(sdk.NewInt64Coin("foocoin", 67)))
	require.NoError(t, err)

	app.Commit()

	res1 := app.AccountKeeper.GetAccount(ctx, addr1)
	require.NotNil(t, res1)
	require.Equal(t, acc, res1.(*authtypes.BaseAccount))

	origAccNum := res1.GetAccountNumber()
	origSeq := res1.GetSequence()

	sendMsg := types.NewMsgSend(addr1, addr2, sdk.Coins{sdk.NewInt64Coin("foocoin", 100)})
	header := tmproto.Header{Height: app.LastBlockHeight() + 1}
	txGen := simapp.MakeTestEncodingConfig().TxConfig
	_, _, err = simapp.SignCheckDeliver(t, txGen, app.BaseApp, header, []sdk.Msg{sendMsg}, "", []uint64{origAccNum}, []uint64{origSeq}, false, false, priv1)
	require.Error(t, err)

	simapp.CheckBalance(t, app, addr1, sdk.Coins{sdk.NewInt64Coin("foocoin", 67)})

	res2 := app.AccountKeeper.GetAccount(app.NewContext(true, tmproto.Header{}), addr1)
	require.NotNil(t, res2)

	require.Equal(t, res2.GetAccountNumber(), origAccNum)
	require.Equal(t, res2.GetSequence(), origSeq+1)
}

func TestMsgMultiSendWithAccounts(t *testing.T) {
	acc := &authtypes.BaseAccount{
		Address: addr1.String(),
	}

	genAccs := []authtypes.GenesisAccount{acc}
	app := simapp.SetupWithGenesisAccounts(genAccs)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})

	err := app.BankKeeper.SetBalances(ctx, addr1, sdk.NewCoins(sdk.NewInt64Coin("foocoin", 67)))
	require.NoError(t, err)

	app.Commit()

	res1 := app.AccountKeeper.GetAccount(ctx, addr1)
	require.NotNil(t, res1)
	require.Equal(t, acc, res1.(*authtypes.BaseAccount))

	testCases := []appTestCase{
		{
			desc:       "make a valid tx",
			msgs:       []sdk.Msg{multiSendMsg1},
			accNums:    []uint64{0},
			accSeqs:    []uint64{0},
			expSimPass: true,
			expPass:    true,
			privKeys:   []cryptotypes.PrivKey{priv1},
			expectedBalances: []expectedBalance{
				{addr1, sdk.Coins{sdk.NewInt64Coin("foocoin", 57)}},
				{addr2, sdk.Coins{sdk.NewInt64Coin("foocoin", 10)}},
			},
		},
		{
			desc:       "wrong accNum should pass Simulate, but not Deliver",
			msgs:       []sdk.Msg{multiSendMsg1, multiSendMsg2},
			accNums:    []uint64{1}, // wrong account number
			accSeqs:    []uint64{1},
			expSimPass: true, // doesn't check signature
			expPass:    false,
			privKeys:   []cryptotypes.PrivKey{priv1},
		},
		{
			desc:       "wrong accSeq should not pass Simulate",
			msgs:       []sdk.Msg{multiSendMsg5},
			accNums:    []uint64{0},
			accSeqs:    []uint64{0}, // wrong account sequence
			expSimPass: false,
			expPass:    false,
			privKeys:   []cryptotypes.PrivKey{priv1},
		},
	}

	for _, tc := range testCases {
		header := tmproto.Header{Height: app.LastBlockHeight() + 1}
		txGen := simapp.MakeTestEncodingConfig().TxConfig
		_, _, err := simapp.SignCheckDeliver(t, txGen, app.BaseApp, header, tc.msgs, "", tc.accNums, tc.accSeqs, tc.expSimPass, tc.expPass, tc.privKeys...)
		if tc.expPass {
			require.NoError(t, err)
		} else {
			require.Error(t, err)
		}

		for _, eb := range tc.expectedBalances {
			simapp.CheckBalance(t, app, eb.addr, eb.coins)
		}
	}
}

func TestMsgMultiSendMultipleOut(t *testing.T) {
	acc1 := &authtypes.BaseAccount{
		Address: addr1.String(),
	}
	acc2 := &authtypes.BaseAccount{
		Address: addr2.String(),
	}

	genAccs := []authtypes.GenesisAccount{acc1, acc2}
	app := simapp.SetupWithGenesisAccounts(genAccs)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})

	err := app.BankKeeper.SetBalances(ctx, addr1, sdk.NewCoins(sdk.NewInt64Coin("foocoin", 42)))
	require.NoError(t, err)

	err = app.BankKeeper.SetBalances(ctx, addr2, sdk.NewCoins(sdk.NewInt64Coin("foocoin", 42)))
	require.NoError(t, err)

	app.Commit()

	testCases := []appTestCase{
		{
			msgs:       []sdk.Msg{multiSendMsg2},
			accNums:    []uint64{0},
			accSeqs:    []uint64{0},
			expSimPass: true,
			expPass:    true,
			privKeys:   []cryptotypes.PrivKey{priv1},
			expectedBalances: []expectedBalance{
				{addr1, sdk.Coins{sdk.NewInt64Coin("foocoin", 32)}},
				{addr2, sdk.Coins{sdk.NewInt64Coin("foocoin", 47)}},
				{addr3, sdk.Coins{sdk.NewInt64Coin("foocoin", 5)}},
			},
		},
	}

	for _, tc := range testCases {
		header := tmproto.Header{Height: app.LastBlockHeight() + 1}
		txGen := simapp.MakeTestEncodingConfig().TxConfig
		_, _, err := simapp.SignCheckDeliver(t, txGen, app.BaseApp, header, tc.msgs, "", tc.accNums, tc.accSeqs, tc.expSimPass, tc.expPass, tc.privKeys...)
		require.NoError(t, err)

		for _, eb := range tc.expectedBalances {
			simapp.CheckBalance(t, app, eb.addr, eb.coins)
		}
	}
}

func TestMsgMultiSendMultipleInOut(t *testing.T) {
	acc1 := &authtypes.BaseAccount{
		Address: addr1.String(),
	}
	acc2 := &authtypes.BaseAccount{
		Address: addr2.String(),
	}
	acc4 := &authtypes.BaseAccount{
		Address: addr4.String(),
	}

	genAccs := []authtypes.GenesisAccount{acc1, acc2, acc4}
	app := simapp.SetupWithGenesisAccounts(genAccs)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})

	err := app.BankKeeper.SetBalances(ctx, addr1, sdk.NewCoins(sdk.NewInt64Coin("foocoin", 42)))
	require.NoError(t, err)

	err = app.BankKeeper.SetBalances(ctx, addr2, sdk.NewCoins(sdk.NewInt64Coin("foocoin", 42)))
	require.NoError(t, err)

	err = app.BankKeeper.SetBalances(ctx, addr4, sdk.NewCoins(sdk.NewInt64Coin("foocoin", 42)))
	require.NoError(t, err)

	app.Commit()

	testCases := []appTestCase{
		{
			msgs:       []sdk.Msg{multiSendMsg3},
			accNums:    []uint64{0, 2},
			accSeqs:    []uint64{0, 0},
			expSimPass: true,
			expPass:    true,
			privKeys:   []cryptotypes.PrivKey{priv1, priv4},
			expectedBalances: []expectedBalance{
				{addr1, sdk.Coins{sdk.NewInt64Coin("foocoin", 32)}},
				{addr4, sdk.Coins{sdk.NewInt64Coin("foocoin", 32)}},
				{addr2, sdk.Coins{sdk.NewInt64Coin("foocoin", 52)}},
				{addr3, sdk.Coins{sdk.NewInt64Coin("foocoin", 10)}},
			},
		},
	}

	for _, tc := range testCases {
		header := tmproto.Header{Height: app.LastBlockHeight() + 1}
		txGen := simapp.MakeTestEncodingConfig().TxConfig
		_, _, err := simapp.SignCheckDeliver(t, txGen, app.BaseApp, header, tc.msgs, "", tc.accNums, tc.accSeqs, tc.expSimPass, tc.expPass, tc.privKeys...)
		require.NoError(t, err)

		for _, eb := range tc.expectedBalances {
			simapp.CheckBalance(t, app, eb.addr, eb.coins)
		}
	}
}

func TestMsgMultiSendDependent(t *testing.T) {
	acc1 := authtypes.NewBaseAccountWithAddress(addr1)
	acc2 := authtypes.NewBaseAccountWithAddress(addr2)
	err := acc2.SetAccountNumber(1)
	require.NoError(t, err)

	genAccs := []authtypes.GenesisAccount{acc1, acc2}
	app := simapp.SetupWithGenesisAccounts(genAccs)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})

	err = app.BankKeeper.SetBalances(ctx, addr1, sdk.NewCoins(sdk.NewInt64Coin("foocoin", 42)))
	require.NoError(t, err)

	app.Commit()

	testCases := []appTestCase{
		{
			msgs:       []sdk.Msg{multiSendMsg1},
			accNums:    []uint64{0},
			accSeqs:    []uint64{0},
			expSimPass: true,
			expPass:    true,
			privKeys:   []cryptotypes.PrivKey{priv1},
			expectedBalances: []expectedBalance{
				{addr1, sdk.Coins{sdk.NewInt64Coin("foocoin", 32)}},
				{addr2, sdk.Coins{sdk.NewInt64Coin("foocoin", 10)}},
			},
		},
		{
			msgs:       []sdk.Msg{multiSendMsg4},
			accNums:    []uint64{1},
			accSeqs:    []uint64{0},
			expSimPass: true,
			expPass:    true,
			privKeys:   []cryptotypes.PrivKey{priv2},
			expectedBalances: []expectedBalance{
				{addr1, sdk.Coins{sdk.NewInt64Coin("foocoin", 42)}},
			},
		},
	}

	for _, tc := range testCases {
		header := tmproto.Header{Height: app.LastBlockHeight() + 1}
		txGen := simapp.MakeTestEncodingConfig().TxConfig
		_, _, err := simapp.SignCheckDeliver(t, txGen, app.BaseApp, header, tc.msgs, "", tc.accNums, tc.accSeqs, tc.expSimPass, tc.expPass, tc.privKeys...)
		require.NoError(t, err)

		for _, eb := range tc.expectedBalances {
			simapp.CheckBalance(t, app, eb.addr, eb.coins)
		}
	}
}
