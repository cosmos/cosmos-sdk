package bank

import (
	"testing"

	"github.com/stretchr/testify/require"
	secp256k1_internal "gitlab.com/yawning/secp256k1-voi"
	"gitlab.com/yawning/secp256k1-voi/secec"

	"cosmossdk.io/depinject"
	"cosmossdk.io/log"
	sdkmath "cosmossdk.io/math"
	_ "cosmossdk.io/x/accounts"
	_ "cosmossdk.io/x/bank"
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
	stakingtypes "cosmossdk.io/x/staking/types"

	"github.com/cosmos/cosmos-sdk/client"
	cdctestutil "github.com/cosmos/cosmos-sdk/codec/testutil"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/tests/integration/v2"
	"github.com/cosmos/cosmos-sdk/testutil/configurator"
	sdk "github.com/cosmos/cosmos-sdk/types"
	_ "github.com/cosmos/cosmos-sdk/x/auth"
	_ "github.com/cosmos/cosmos-sdk/x/auth/tx/config"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
)

var (
	stablePrivateKey, _ = secec.NewPrivateKeyFromScalar(secp256k1_internal.NewScalarFromUint64(100))
	priv1               = &secp256k1.PrivKey{Key: stablePrivateKey.Bytes()}
	addr1               = sdk.AccAddress(priv1.PubKey().Address())
	priv2               = secp256k1.GenPrivKey()
	addr2               = sdk.AccAddress(priv2.PubKey().Address())
	addr3               = sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())
	coins               = sdk.Coins{sdk.NewInt64Coin("foocoin", 10)}
	halfCoins           = sdk.Coins{sdk.NewInt64Coin("foocoin", 5)}
	moduleAccAddr       = authtypes.NewModuleAddress(stakingtypes.BondedPoolName)
)

type suite struct {
	BankKeeper         bankkeeper.Keeper
	AccountKeeper      types.AccountKeeper
	DistributionKeeper distrkeeper.Keeper
	App                *integration.App
	TxConfig           client.TxConfig
}

type expectedBalance struct {
	addr  sdk.AccAddress
	coins sdk.Coins
}

type appTestCase struct {
	desc             string
	msgs             []sdk.Msg
	accNums          []uint64
	accSeqs          []uint64
	privKeys         []cryptotypes.PrivKey
	expectedBalances []expectedBalance
	expInError       []string
}

func createTestSuite(t *testing.T, genesisAccounts []authtypes.GenesisAccount) suite {
	t.Helper()
	res := suite{}

	moduleConfigs := []configurator.ModuleOption{
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
	}
	var err error
	startupCfg := integration.DefaultStartUpConfig(t)
	var genAccounts []integration.GenesisAccount
	for _, acc := range genesisAccounts {
		genAccounts = append(genAccounts, integration.GenesisAccount{GenesisAccount: acc})
	}
	startupCfg.GenesisAccounts = genAccounts
	res.App, err = integration.NewApp(
		depinject.Configs(configurator.NewAppV2Config(moduleConfigs...), depinject.Supply(log.NewNopLogger())),
		startupCfg,
		&res.BankKeeper, &res.AccountKeeper, &res.DistributionKeeper, &res.TxConfig)
	require.NoError(t, err)

	return res
}

func TestSendNotEnoughBalance(t *testing.T) {
	acc := &authtypes.BaseAccount{
		Address: addr1.String(),
	}

	genAccs := []authtypes.GenesisAccount{acc}
	s := createTestSuite(t, genAccs)
	ctx := s.App.StateLatestContext(t)

	err := testutil.FundAccount(
		ctx, s.BankKeeper, addr1,
		sdk.NewCoins(sdk.NewInt64Coin("foocoin", 67)))
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

	// TODO how to auto-advance height with app v2 interface?
	s.App.SignCheckDeliver(
		t, ctx, []sdk.Msg{sendMsg}, "", []uint64{origAccNum}, []uint64{origSeq},
		[]cryptotypes.PrivKey{priv1},
		"spendable balance 67foocoin is smaller than 100foocoin",
	)
	s.App.CheckBalance(t, ctx, addr1, sdk.Coins{sdk.NewInt64Coin("foocoin", 67)}, s.BankKeeper)
	res2 := s.AccountKeeper.GetAccount(ctx, addr1)
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
	ctx := s.App.StateLatestContext(t)

	require.NoError(t, testutil.FundAccount(ctx, s.BankKeeper, addr1, sdk.NewCoins(sdk.NewInt64Coin("foocoin", 67))))

	_, state := s.App.Deliver(t, ctx, nil)
	_, err = s.App.Commit(state)
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
			accNums:  []uint64{0},
			accSeqs:  []uint64{0},
			privKeys: []cryptotypes.PrivKey{priv1},
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
			expInError: []string{"signature verification failed; please verify account number"},
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
			expInError: []string{"account sequence mismatch"},
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
			expInError: []string{"invalid number of signatures"},
			privKeys:   []cryptotypes.PrivKey{priv1},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			var errString string
			if len(tc.expInError) > 0 {
				errString = tc.expInError[0]
			}
			s.App.SignCheckDeliver(t, ctx, tc.msgs, "", tc.accNums, tc.accSeqs, tc.privKeys, errString)

			for _, eb := range tc.expectedBalances {
				s.App.CheckBalance(t, ctx, eb.addr, eb.coins, s.BankKeeper)
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
	ctx := s.App.StateLatestContext(t)

	require.NoError(t, testutil.FundAccount(ctx, s.BankKeeper, addr1, sdk.NewCoins(sdk.NewInt64Coin("foocoin", 42))))
	require.NoError(t, testutil.FundAccount(ctx, s.BankKeeper, addr2, sdk.NewCoins(sdk.NewInt64Coin("foocoin", 42))))
	_, state := s.App.Deliver(t, ctx, nil)
	_, err = s.App.Commit(state)
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
			accNums:  []uint64{0},
			accSeqs:  []uint64{0},
			privKeys: []cryptotypes.PrivKey{priv1},
			expectedBalances: []expectedBalance{
				{addr1, sdk.Coins{sdk.NewInt64Coin("foocoin", 32)}},
				{addr2, sdk.Coins{sdk.NewInt64Coin("foocoin", 47)}},
				{addr3, sdk.Coins{sdk.NewInt64Coin("foocoin", 5)}},
			},
		},
	}

	for _, tc := range testCases {
		s.App.SignCheckDeliver(t, ctx, tc.msgs, "", tc.accNums, tc.accSeqs, tc.privKeys, "")

		for _, eb := range tc.expectedBalances {
			s.App.CheckBalance(t, ctx, eb.addr, eb.coins, s.BankKeeper)
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
	ctx := s.App.StateLatestContext(t)

	require.NoError(t, testutil.FundAccount(ctx, s.BankKeeper, addr1, sdk.NewCoins(sdk.NewInt64Coin("foocoin", 42))))
	_, state := s.App.Deliver(t, ctx, nil)
	_, err = s.App.Commit(state)
	require.NoError(t, err)

	testCases := []appTestCase{
		{
			msgs: []sdk.Msg{&types.MsgMultiSend{
				Inputs:  []types.Input{types.NewInput(addr1Str, coins)},
				Outputs: []types.Output{types.NewOutput(addr2Str, coins)},
			}},
			accNums:  []uint64{0},
			accSeqs:  []uint64{0},
			privKeys: []cryptotypes.PrivKey{priv1},
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
			accNums:  []uint64{1},
			accSeqs:  []uint64{0},
			privKeys: []cryptotypes.PrivKey{priv2},
			expectedBalances: []expectedBalance{
				{addr1, sdk.Coins{sdk.NewInt64Coin("foocoin", 42)}},
			},
		},
	}

	for _, tc := range testCases {
		s.App.SignCheckDeliver(t, ctx, tc.msgs, "", tc.accNums, tc.accSeqs, tc.privKeys, "")

		for _, eb := range tc.expectedBalances {
			s.App.CheckBalance(t, ctx, eb.addr, eb.coins, s.BankKeeper)
		}
	}
}

func TestMsgSetSendEnabled(t *testing.T) {
	acc1 := authtypes.NewBaseAccountWithAddress(addr1)

	genAccs := []authtypes.GenesisAccount{acc1}
	s := createTestSuite(t, genAccs)

	ctx := s.App.StateLatestContext(t)
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
			desc: "wrong authority",
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
			desc: "right authority wrong signer",
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
			desc: "submitted good as gov prop",
			msgs: []sdk.Msg{
				goodGovProp,
			},
			accSeqs:    []uint64{1},
			expInError: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(tt *testing.T) {
			var errString string
			if len(tc.expInError) > 0 {
				errString = tc.expInError[0]
			}
			txResult := s.App.SignCheckDeliver(
				tt, ctx, tc.msgs, "", []uint64{0}, tc.accSeqs, []cryptotypes.PrivKey{priv1}, errString)
			if len(tc.expInError) > 0 {
				require.Error(tt, txResult.Error)
				for _, exp := range tc.expInError {
					require.ErrorContains(tt, txResult.Error, exp)
				}
			} else {
				require.NoError(tt, txResult.Error)
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
	ctx := s.App.StateLatestContext(t)

	require.NoError(t, testutil.FundAccount(ctx, s.BankKeeper, addr1, sdk.NewCoins(sdk.NewInt64Coin("foocoin", 42))))
	_, state := s.App.Deliver(t, ctx, nil)
	_, err := s.App.Commit(state)
	require.NoError(t, err)

	addr2Str, err := s.AccountKeeper.AddressCodec().BytesToString(addr2)
	require.NoError(t, err)
	sendMsg := types.NewMsgSend(addr1.String(), addr2Str, coins)
	res := s.App.SignCheckDeliver(t, ctx, []sdk.Msg{sendMsg}, "", []uint64{0}, []uint64{0}, []cryptotypes.PrivKey{priv1}, "")
	require.NoError(t, res.Error)

	// Check that the account was not created
	acc2 := s.AccountKeeper.GetAccount(ctx, addr2)
	require.Nil(t, acc2)

	// But it does have a balance
	s.App.CheckBalance(t, ctx, addr2, coins, s.BankKeeper)

	// Now we send coins back and the account should be created
	sendMsg = types.NewMsgSend(addr2Str, addr1.String(), coins)
	res = s.App.SignCheckDeliver(t, ctx, []sdk.Msg{sendMsg}, "", []uint64{0}, []uint64{0}, []cryptotypes.PrivKey{priv2}, "")
	require.NoError(t, res.Error)

	// Balance has been reduced
	s.App.CheckBalance(t, ctx, addr2, sdk.NewCoins(), s.BankKeeper)

	// Check that the account was created
	acc2 = s.AccountKeeper.GetAccount(ctx, addr2)
	require.NotNil(t, acc2, "account should have been created %s", addr2.String())
}
