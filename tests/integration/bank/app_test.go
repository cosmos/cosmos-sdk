package bank_test

import (
	"testing"

	abci "github.com/cometbft/cometbft/api/cometbft/abci/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"cosmossdk.io/core/header"
	"cosmossdk.io/depinject"
	"cosmossdk.io/log"
	sdkmath "cosmossdk.io/math"
	_ "cosmossdk.io/x/accounts"
	bankkeeper "cosmossdk.io/x/bank/keeper"
	"cosmossdk.io/x/bank/testutil"
	"cosmossdk.io/x/bank/types"
	_ "cosmossdk.io/x/consensus"
	_ "cosmossdk.io/x/distribution"
	distrkeeper "cosmossdk.io/x/distribution/keeper"
	_ "cosmossdk.io/x/gov"
	govv1 "cosmossdk.io/x/gov/types/v1"
	_ "cosmossdk.io/x/protocolpool"
	_ "cosmossdk.io/x/staking"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/client"
	cdctestutil "github.com/cosmos/cosmos-sdk/codec/testutil"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/runtime"
	"github.com/cosmos/cosmos-sdk/testutil/configurator"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	sdk "github.com/cosmos/cosmos-sdk/types"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
	_ "github.com/cosmos/cosmos-sdk/x/auth"
	_ "github.com/cosmos/cosmos-sdk/x/auth/tx/config"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
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

	coins     = sdk.Coins{sdk.NewInt64Coin("foocoin", 10)}
	halfCoins = sdk.Coins{sdk.NewInt64Coin("foocoin", 5)}

	sendMsg1 = types.NewMsgSend(addr1.String(), addr2.String(), coins)
)

type suite struct {
	BankKeeper         bankkeeper.Keeper
	AccountKeeper      types.AccountKeeper
	DistributionKeeper distrkeeper.Keeper
	App                *runtime.App
	TxConfig           client.TxConfig
}

func createTestSuite(t *testing.T, genesisAccounts []authtypes.GenesisAccount) suite {
	t.Helper()
	res := suite{}

	var genAccounts []simtestutil.GenesisAccount
	for _, acc := range genesisAccounts {
		genAccounts = append(genAccounts, simtestutil.GenesisAccount{GenesisAccount: acc})
	}

	startupCfg := simtestutil.DefaultStartUpConfig()
	startupCfg.GenesisAccounts = genAccounts

	app, err := simtestutil.SetupWithConfiguration(
		depinject.Configs(
			configurator.NewAppConfig(
				configurator.AccountsModule(),
				configurator.AuthModule(),
				configurator.StakingModule(),
				configurator.TxModule(),
				configurator.ValidateModule(),
				configurator.ConsensusModule(),
				configurator.BankModule(),
				configurator.GovModule(),
				configurator.DistributionModule(),
				configurator.ProtocolPoolModule(),
			),
			depinject.Supply(log.NewNopLogger()),
		),
		startupCfg,
		&res.BankKeeper,
		&res.AccountKeeper,
		&res.DistributionKeeper,
		&res.TxConfig,
	)

	res.App = app
	res.App.SetTxEncoder(res.TxConfig.TxEncoder())
	res.App.SetTxDecoder(res.TxConfig.TxDecoder())

	require.NoError(t, err)
	return res
}

// CheckBalance checks the balance of an account.
func checkBalance(t *testing.T, baseApp *baseapp.BaseApp, addr sdk.AccAddress, balances sdk.Coins, keeper bankkeeper.Keeper) {
	t.Helper()
	ctxCheck := baseApp.NewContext(true)
	keeperBalances := keeper.GetAllBalances(ctxCheck, addr)
	require.True(t, balances.Equal(keeperBalances), balances.String(), keeperBalances.String())
}

func TestSendNotEnoughBalance(t *testing.T) {
	acc := &authtypes.BaseAccount{
		Address: addr1.String(),
	}

	genAccs := []authtypes.GenesisAccount{acc}
	s := createTestSuite(t, genAccs)
	baseApp := s.App.BaseApp
	ctx := baseApp.NewContext(false)

	require.NoError(t, testutil.FundAccount(ctx, s.BankKeeper, addr1, sdk.NewCoins(sdk.NewInt64Coin("foocoin", 67))))
	_, err := baseApp.FinalizeBlock(&abci.FinalizeBlockRequest{Height: baseApp.LastBlockHeight() + 1})
	require.NoError(t, err)
	_, err = baseApp.Commit()
	require.NoError(t, err)

	res1 := s.AccountKeeper.GetAccount(ctx, addr1)
	require.NotNil(t, res1)
	require.Equal(t, acc, res1.(*authtypes.BaseAccount))

	origAccNum := res1.GetAccountNumber()
	origSeq := res1.GetSequence()

	addr1Str, err := s.AccountKeeper.AddressCodec().BytesToString(addr1)
	require.NoError(t, err)
	addr2Str, err := s.AccountKeeper.AddressCodec().BytesToString(addr2)
	require.NoError(t, err)
	sendMsg := types.NewMsgSend(addr1Str, addr2Str, sdk.Coins{sdk.NewInt64Coin("foocoin", 100)})
	header := header.Info{Height: baseApp.LastBlockHeight() + 1}
	txConfig := moduletestutil.MakeTestTxConfig(cdctestutil.CodecOptions{})
	_, _, err = simtestutil.SignCheckDeliver(t, txConfig, baseApp, header, []sdk.Msg{sendMsg}, "", []uint64{origAccNum}, []uint64{origSeq}, false, false, priv1)
	require.Error(t, err)

	checkBalance(t, baseApp, addr1, sdk.Coins{sdk.NewInt64Coin("foocoin", 67)}, s.BankKeeper)

	ctx2 := baseApp.NewContext(true)
	res2 := s.AccountKeeper.GetAccount(ctx2, addr1)
	require.NotNil(t, res2)

	require.Equal(t, origAccNum, res2.GetAccountNumber())
	require.Equal(t, origSeq+1, res2.GetSequence())
}

func TestMsgMultiSendWithAccounts(t *testing.T) {
	addr1Str, err := cdctestutil.CodecOptions{}.GetAddressCodec().BytesToString(addr1)
	require.NoError(t, err)
	acc := &authtypes.BaseAccount{
		Address: addr1Str,
	}

	addr2Str, err := cdctestutil.CodecOptions{}.GetAddressCodec().BytesToString(addr2)
	require.NoError(t, err)

	moduleStrAddr, err := cdctestutil.CodecOptions{}.GetAddressCodec().BytesToString(moduleAccAddr)
	require.NoError(t, err)

	genAccs := []authtypes.GenesisAccount{acc}
	s := createTestSuite(t, genAccs)
	baseApp := s.App.BaseApp
	ctx := baseApp.NewContext(false)

	require.NoError(t, testutil.FundAccount(ctx, s.BankKeeper, addr1, sdk.NewCoins(sdk.NewInt64Coin("foocoin", 67))))
	_, err = baseApp.FinalizeBlock(&abci.FinalizeBlockRequest{Height: baseApp.LastBlockHeight() + 1})
	require.NoError(t, err)
	_, err = baseApp.Commit()
	require.NoError(t, err)

	res1 := s.AccountKeeper.GetAccount(ctx, addr1)
	require.NotNil(t, res1)
	require.Equal(t, acc, res1.(*authtypes.BaseAccount))

	testCases := []appTestCase{
		{
			desc: "make a valid tx",
			msgs: []sdk.Msg{&types.MsgMultiSend{
				Inputs:  []types.Input{types.NewInput(addr1Str, coins)},
				Outputs: []types.Output{types.NewOutput(addr2Str, coins)},
			}},
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
			desc: "wrong accNum should pass Simulate, but not Deliver",
			msgs: []sdk.Msg{&types.MsgMultiSend{
				Inputs:  []types.Input{types.NewInput(addr1Str, coins)},
				Outputs: []types.Output{types.NewOutput(addr2Str, coins)},
			}},
			accNums:    []uint64{1}, // wrong account number
			accSeqs:    []uint64{1},
			expSimPass: true, // doesn't check signature
			expPass:    false,
			privKeys:   []cryptotypes.PrivKey{priv1},
		},
		{
			desc: "wrong accSeq should not pass Simulate",
			msgs: []sdk.Msg{&types.MsgMultiSend{
				Inputs: []types.Input{types.NewInput(addr1Str, coins)},
				Outputs: []types.Output{
					types.NewOutput(moduleStrAddr, coins),
				},
			}},
			accNums:    []uint64{0},
			accSeqs:    []uint64{0}, // wrong account sequence
			expSimPass: false,
			expPass:    false,
			privKeys:   []cryptotypes.PrivKey{priv1},
		},
		{
			desc: "multiple inputs not allowed",
			msgs: []sdk.Msg{&types.MsgMultiSend{
				Inputs:  []types.Input{types.NewInput(addr1Str, coins), types.NewInput(addr2Str, coins)},
				Outputs: []types.Output{},
			}},
			accNums:    []uint64{0},
			accSeqs:    []uint64{0},
			expSimPass: false,
			expPass:    false,
			privKeys:   []cryptotypes.PrivKey{priv1},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			header := header.Info{Height: baseApp.LastBlockHeight() + 1}
			txConfig := moduletestutil.MakeTestTxConfig(cdctestutil.CodecOptions{})
			_, _, err := simtestutil.SignCheckDeliver(t, txConfig, baseApp, header, tc.msgs, "", tc.accNums, tc.accSeqs, tc.expSimPass, tc.expPass, tc.privKeys...)
			if tc.expPass {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}

			for _, eb := range tc.expectedBalances {
				checkBalance(t, baseApp, eb.addr, eb.coins, s.BankKeeper)
			}
		})
	}
}

func TestMsgMultiSendMultipleOut(t *testing.T) {
	ac := cdctestutil.CodecOptions{}.GetAddressCodec()
	addr1Str, err := ac.BytesToString(addr1)
	require.NoError(t, err)
	acc1 := &authtypes.BaseAccount{
		Address: addr1Str,
	}
	addr2Str, err := ac.BytesToString(addr2)
	require.NoError(t, err)
	acc2 := &authtypes.BaseAccount{
		Address: addr2Str,
	}
	addr3Str, err := ac.BytesToString(addr3)
	require.NoError(t, err)

	genAccs := []authtypes.GenesisAccount{acc1, acc2}
	s := createTestSuite(t, genAccs)
	baseApp := s.App.BaseApp
	ctx := baseApp.NewContext(false)

	require.NoError(t, testutil.FundAccount(ctx, s.BankKeeper, addr1, sdk.NewCoins(sdk.NewInt64Coin("foocoin", 42))))
	require.NoError(t, testutil.FundAccount(ctx, s.BankKeeper, addr2, sdk.NewCoins(sdk.NewInt64Coin("foocoin", 42))))
	_, err = baseApp.FinalizeBlock(&abci.FinalizeBlockRequest{Height: baseApp.LastBlockHeight() + 1})
	require.NoError(t, err)
	_, err = baseApp.Commit()
	require.NoError(t, err)

	testCases := []appTestCase{
		{
			msgs: []sdk.Msg{&types.MsgMultiSend{
				Inputs: []types.Input{types.NewInput(addr1Str, coins)},
				Outputs: []types.Output{
					types.NewOutput(addr2Str, halfCoins),
					types.NewOutput(addr3Str, halfCoins),
				},
			}},
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
		header := header.Info{Height: baseApp.LastBlockHeight() + 1}
		txConfig := moduletestutil.MakeTestTxConfig(cdctestutil.CodecOptions{})
		_, _, err := simtestutil.SignCheckDeliver(t, txConfig, baseApp, header, tc.msgs, "", tc.accNums, tc.accSeqs, tc.expSimPass, tc.expPass, tc.privKeys...)
		require.NoError(t, err)

		for _, eb := range tc.expectedBalances {
			checkBalance(t, baseApp, eb.addr, eb.coins, s.BankKeeper)
		}
	}
}

func TestMsgMultiSendDependent(t *testing.T) {
	ac := cdctestutil.CodecOptions{}.GetAddressCodec()
	addr1Str, err := ac.BytesToString(addr1)
	require.NoError(t, err)
	addr2Str, err := ac.BytesToString(addr2)
	require.NoError(t, err)

	acc1 := authtypes.NewBaseAccountWithAddress(addr1)
	acc2 := authtypes.NewBaseAccountWithAddress(addr2)
	err = acc2.SetAccountNumber(1)
	require.NoError(t, err)

	genAccs := []authtypes.GenesisAccount{acc1, acc2}
	s := createTestSuite(t, genAccs)
	baseApp := s.App.BaseApp
	ctx := baseApp.NewContext(false)

	require.NoError(t, testutil.FundAccount(ctx, s.BankKeeper, addr1, sdk.NewCoins(sdk.NewInt64Coin("foocoin", 42))))
	_, err = baseApp.FinalizeBlock(&abci.FinalizeBlockRequest{Height: baseApp.LastBlockHeight() + 1})
	require.NoError(t, err)
	_, err = baseApp.Commit()
	require.NoError(t, err)

	testCases := []appTestCase{
		{
			msgs: []sdk.Msg{&types.MsgMultiSend{
				Inputs:  []types.Input{types.NewInput(addr1Str, coins)},
				Outputs: []types.Output{types.NewOutput(addr2Str, coins)},
			}},
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
			msgs: []sdk.Msg{&types.MsgMultiSend{
				Inputs: []types.Input{types.NewInput(addr2Str, coins)},
				Outputs: []types.Output{
					types.NewOutput(addr1Str, coins),
				},
			}},
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
		header := header.Info{Height: baseApp.LastBlockHeight() + 1}
		txConfig := moduletestutil.MakeTestTxConfig(cdctestutil.CodecOptions{})
		_, _, err := simtestutil.SignCheckDeliver(t, txConfig, baseApp, header, tc.msgs, "", tc.accNums, tc.accSeqs, tc.expSimPass, tc.expPass, tc.privKeys...)
		require.NoError(t, err)

		for _, eb := range tc.expectedBalances {
			checkBalance(t, baseApp, eb.addr, eb.coins, s.BankKeeper)
		}
	}
}

func TestMsgSetSendEnabled(t *testing.T) {
	acc1 := authtypes.NewBaseAccountWithAddress(addr1)

	genAccs := []authtypes.GenesisAccount{acc1}
	s := createTestSuite(t, genAccs)

	ctx := s.App.BaseApp.NewContext(false)
	require.NoError(t, testutil.FundAccount(ctx, s.BankKeeper, addr1, sdk.NewCoins(sdk.NewInt64Coin("foocoin", 101))))
	require.NoError(t, testutil.FundAccount(ctx, s.BankKeeper, addr1, sdk.NewCoins(sdk.NewInt64Coin("stake", 100000))))
	addr1Str := addr1.String()
	govAddr := s.BankKeeper.GetAuthority()
	goodGovProp, err := govv1.NewMsgSubmitProposal(
		[]sdk.Msg{
			types.NewMsgSetSendEnabled(govAddr, nil, nil),
		},
		sdk.Coins{{Denom: "stake", Amount: sdkmath.NewInt(100000)}},
		addr1Str,
		"set default send enabled to true",
		"Change send enabled",
		"Modify send enabled and set to true",
		govv1.ProposalType_PROPOSAL_TYPE_STANDARD,
	)
	require.NoError(t, err, "making goodGovProp")

	testCases := []appTestCase{
		{
			desc:       "wrong authority",
			expSimPass: false,
			expPass:    false,
			msgs: []sdk.Msg{
				types.NewMsgSetSendEnabled(addr1Str, nil, nil),
			},
			accSeqs: []uint64{0},
			expInError: []string{
				"invalid authority",
				"cosmos10d07y265gmmuvt4z0w9aw880jnsr700j6zn9kn",
				addr1Str,
				"expected authority account as only signer for proposal message",
			},
		},
		{
			desc:       "right authority wrong signer",
			expSimPass: false,
			expPass:    false,
			msgs: []sdk.Msg{
				types.NewMsgSetSendEnabled(govAddr, nil, nil),
			},
			accSeqs: []uint64{1}, // wrong signer, so this sequence doesn't actually get used.
			expInError: []string{
				"cannot be claimed by public key with address",
				govAddr,
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
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(tt *testing.T) {
			header := header.Info{Height: s.App.LastBlockHeight() + 1}
			_, _, err = simtestutil.SignCheckDeliver(tt, s.TxConfig, s.App.BaseApp, header, tc.msgs, "", []uint64{0}, tc.accSeqs, tc.expSimPass, tc.expPass, priv1)
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

// TestSendToNonExistingAccount tests sending coins to an account that does not exist, and this account
// must not be created.
func TestSendToNonExistingAccount(t *testing.T) {
	acc1 := authtypes.NewBaseAccountWithAddress(addr1)
	genAccs := []authtypes.GenesisAccount{acc1}
	s := createTestSuite(t, genAccs)
	baseApp := s.App.BaseApp
	ctx := baseApp.NewContext(false)

	require.NoError(t, testutil.FundAccount(ctx, s.BankKeeper, addr1, sdk.NewCoins(sdk.NewInt64Coin("foocoin", 42))))
	_, err := baseApp.FinalizeBlock(&abci.FinalizeBlockRequest{Height: baseApp.LastBlockHeight() + 1})
	require.NoError(t, err)
	_, err = baseApp.Commit()
	require.NoError(t, err)

	addr2Str, err := s.AccountKeeper.AddressCodec().BytesToString(addr2)
	require.NoError(t, err)
	sendMsg := types.NewMsgSend(addr1.String(), addr2Str, coins)
	h := header.Info{Height: baseApp.LastBlockHeight() + 1}
	txConfig := moduletestutil.MakeTestTxConfig(cdctestutil.CodecOptions{})
	_, _, err = simtestutil.SignCheckDeliver(t, txConfig, baseApp, h, []sdk.Msg{sendMsg}, "", []uint64{0}, []uint64{0}, true, true, priv1)
	require.NoError(t, err)

	// Check that the account was not created
	acc2 := s.AccountKeeper.GetAccount(baseApp.NewContext(true), addr2)
	require.Nil(t, acc2)

	// But it does have a balance
	checkBalance(t, baseApp, addr2, coins, s.BankKeeper)

	// Now we send coins back and the account should be created
	sendMsg = types.NewMsgSend(addr2Str, addr1.String(), coins)
	h = header.Info{Height: baseApp.LastBlockHeight() + 1}
	_, _, err = simtestutil.SignCheckDeliver(t, txConfig, baseApp, h, []sdk.Msg{sendMsg}, "", []uint64{0}, []uint64{0}, true, true, priv2)
	require.NoError(t, err)

	// Balance has been reduced
	checkBalance(t, baseApp, addr2, sdk.NewCoins(), s.BankKeeper)

	// Check that the account was created
	acc2 = s.AccountKeeper.GetAccount(baseApp.NewContext(true), addr2)
	require.NotNil(t, acc2, "account should have been created %s", addr2.String())
}
