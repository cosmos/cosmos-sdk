package bank_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/simapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/bank/testutil"
	"github.com/cosmos/cosmos-sdk/x/bank/types"
	govtypesv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
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
		expInError       []string
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
	app := simapp.SetupWithGenesisAccounts(t, genAccs)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})

	require.NoError(t, testutil.FundAccount(app.BankKeeper, ctx, addr1, sdk.NewCoins(sdk.NewInt64Coin("foocoin", 67))))

	app.Commit()

	res1 := app.AccountKeeper.GetAccount(ctx, addr1)
	require.NotNil(t, res1)
	require.Equal(t, acc, res1.(*authtypes.BaseAccount))

	origAccNum := res1.GetAccountNumber()
	origSeq := res1.GetSequence()

	sendMsg := types.NewMsgSend(addr1, addr2, sdk.Coins{sdk.NewInt64Coin("foocoin", 100)})
	header := tmproto.Header{Height: app.LastBlockHeight() + 1}
	txGen := simapp.MakeTestEncodingConfig().TxConfig
	_, _, err := simapp.SignCheckDeliver(t, txGen, app.BaseApp, header, []sdk.Msg{sendMsg}, "", []uint64{origAccNum}, []uint64{origSeq}, false, false, priv1)
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
	app := simapp.SetupWithGenesisAccounts(t, genAccs)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})

	require.NoError(t, testutil.FundAccount(app.BankKeeper, ctx, addr1, sdk.NewCoins(sdk.NewInt64Coin("foocoin", 67))))

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
	app := simapp.SetupWithGenesisAccounts(t, genAccs)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})

	require.NoError(t, testutil.FundAccount(app.BankKeeper, ctx, addr1, sdk.NewCoins(sdk.NewInt64Coin("foocoin", 42))))

	require.NoError(t, testutil.FundAccount(app.BankKeeper, ctx, addr2, sdk.NewCoins(sdk.NewInt64Coin("foocoin", 42))))

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
	app := simapp.SetupWithGenesisAccounts(t, genAccs)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})

	require.NoError(t, testutil.FundAccount(app.BankKeeper, ctx, addr1, sdk.NewCoins(sdk.NewInt64Coin("foocoin", 42))))

	require.NoError(t, testutil.FundAccount(app.BankKeeper, ctx, addr2, sdk.NewCoins(sdk.NewInt64Coin("foocoin", 42))))

	require.NoError(t, testutil.FundAccount(app.BankKeeper, ctx, addr4, sdk.NewCoins(sdk.NewInt64Coin("foocoin", 42))))

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
	app := simapp.SetupWithGenesisAccounts(t, genAccs)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})

	require.NoError(t, testutil.FundAccount(app.BankKeeper, ctx, addr1, sdk.NewCoins(sdk.NewInt64Coin("foocoin", 42))))

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

func TestMsgSetSendEnabled(t *testing.T) {
	acc1 := authtypes.NewBaseAccountWithAddress(addr1)
	genAccs := []authtypes.GenesisAccount{acc1}
	app := simapp.SetupWithGenesisAccounts(t, genAccs)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})
	require.NoError(t, testutil.FundAccount(app.BankKeeper, ctx, addr1, sdk.NewCoins(sdk.NewInt64Coin("foocoin", 101))))
	addr1Str := addr1.String()
	govAddr := app.BankKeeper.GetAuthority()
	goodGovProp, err := govtypesv1.NewMsgSubmitProposal(
		[]sdk.Msg{
			types.NewMsgSetSendEnabled(govAddr, nil, nil, true, true),
		},
		sdk.Coins{{"foocoin", sdk.NewInt(5)}},
		addr1Str,
		"set default send enabled to true",
	)
	require.NoError(t, err, "making goodGovProp")
	badGovProp, err := govtypesv1.NewMsgSubmitProposal(
		[]sdk.Msg{
			types.NewMsgSetSendEnabled(govAddr, []*types.SendEnabled{{"bad coin name!", true}}, nil, true, true),
		},
		sdk.Coins{{"foocoin", sdk.NewInt(5)}},
		addr1Str,
		"set default send enabled to true",
	)
	require.NoError(t, err, "making badGovProp")

	testCases := []appTestCase{
		{
			desc:       "wrong authority",
			expSimPass: false,
			expPass:    false,
			msgs: []sdk.Msg{
				types.NewMsgSetSendEnabled(addr1Str, nil, nil, true, true),
			},
			accSeqs: []uint64{0},
			expInError: []string{
				"incorrect authority",
				`"cosmos10d07y265gmmuvt4z0w9aw880jnsr700j6zn9kn"`,
				`"` + addr1Str + `"`,
				"tx intended signer does not match the given signer",
			},
		},
		{
			desc:       "right authority wrong signer",
			expSimPass: false,
			expPass:    false,
			msgs: []sdk.Msg{
				types.NewMsgSetSendEnabled(govAddr, nil, nil, true, true),
			},
			accSeqs: []uint64{1}, // wrong signer, so this sequence doesn't actually get used.
			expInError: []string{
				"pubKey does not match signer address",
				govAddr,
				"with signer index: 0",
				"invalid pubkey",
			},
		},
		{
			desc:       "submitted good as gov prop",
			expSimPass: true,
			expPass:    true,
			msgs: []sdk.Msg{
				goodGovProp,
			},
			accSeqs:    []uint64{1},
			expInError: nil,
		},
		{
			desc:       "submitted bad as gov prop",
			expSimPass: false,
			expPass:    false,
			msgs: []sdk.Msg{
				badGovProp,
			},
			accSeqs: []uint64{2},
			expInError: []string{
				"invalid denom: bad coin name!",
				"invalid proposal message",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(tt *testing.T) {
			header := tmproto.Header{Height: app.LastBlockHeight() + 1}
			txGen := simapp.MakeTestEncodingConfig().TxConfig
			_, _, err = simapp.SignCheckDeliver(tt, txGen, app.BaseApp, header, tc.msgs, "", []uint64{0}, tc.accSeqs, tc.expSimPass, tc.expPass, priv1)
			if len(tc.expInError) > 0 {
				require.Error(tt, err)
				for _, exp := range tc.expInError {
					assert.ErrorContains(tt, err, exp)
				}
			} else {
				require.NoError(tt, err)
			}
		})
	}
}
