package bank_test

import (
	"testing"

	abci "github.com/cometbft/cometbft/abci/types"
	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"cosmossdk.io/depinject"
	"cosmossdk.io/log"
	sdkmath "cosmossdk.io/math"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	"github.com/cosmos/cosmos-sdk/runtime"
	"github.com/cosmos/cosmos-sdk/testutil/configurator"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	sdk "github.com/cosmos/cosmos-sdk/types"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
	_ "github.com/cosmos/cosmos-sdk/x/auth"
	_ "github.com/cosmos/cosmos-sdk/x/auth/tx/config"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	"github.com/cosmos/cosmos-sdk/x/bank/testutil"
	"github.com/cosmos/cosmos-sdk/x/bank/types"
	_ "github.com/cosmos/cosmos-sdk/x/consensus"
	_ "github.com/cosmos/cosmos-sdk/x/distribution"
	distrkeeper "github.com/cosmos/cosmos-sdk/x/distribution/keeper"
	_ "github.com/cosmos/cosmos-sdk/x/gov"
	govv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	_ "github.com/cosmos/cosmos-sdk/x/params"
	_ "github.com/cosmos/cosmos-sdk/x/staking"
)

type (
	appTestCase struct {
		desc       string
		expSimPass bool
		expPass    bool
		msgs       []sdk.Msg
		accSeqs    []uint64
		expInError []string
	}
)

var (
	priv1 = secp256k1.GenPrivKey()
	addr1 = sdk.AccAddress(priv1.PubKey().Address())
	priv2 = secp256k1.GenPrivKey()
	addr2 = sdk.AccAddress(priv2.PubKey().Address())

	coins = sdk.Coins{sdk.NewInt64Coin("foocoin", 10)}

	sendMsg1 = types.NewMsgSend(addr1, addr2, coins)
)

type suite struct {
	BankKeeper         bankkeeper.Keeper
	AccountKeeper      types.AccountKeeper
	DistributionKeeper distrkeeper.Keeper
	App                *runtime.App
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
				configurator.ParamsModule(),
				configurator.AuthModule(),
				configurator.StakingModule(),
				configurator.TxModule(),
				configurator.ConsensusModule(),
				configurator.BankModule(),
				configurator.GovModule(),
				configurator.DistributionModule(),
			),
			depinject.Supply(log.NewNopLogger()),
		),
		startupCfg, &res.BankKeeper, &res.AccountKeeper, &res.DistributionKeeper)

	res.App = app

	require.NoError(t, err)
	return res
}

// CheckBalance checks the balance of an account.
func checkBalance(t *testing.T, baseApp *baseapp.BaseApp, addr sdk.AccAddress, balances sdk.Coins, keeper bankkeeper.Keeper) {
	t.Helper()

	ctxCheck := baseApp.NewContext(true)
	keeperBalances := keeper.GetAllBalances(ctxCheck, addr)
	require.True(t, balances.Equal(keeperBalances))
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
	_, err := baseApp.FinalizeBlock(&abci.RequestFinalizeBlock{Height: baseApp.LastBlockHeight() + 1})
	require.NoError(t, err)
	_, err = baseApp.Commit()
	require.NoError(t, err)

	res1 := s.AccountKeeper.GetAccount(ctx, addr1)
	require.NotNil(t, res1)
	require.Equal(t, acc, res1.(*authtypes.BaseAccount))

	origAccNum := res1.GetAccountNumber()
	origSeq := res1.GetSequence()

	sendMsg := types.NewMsgSend(addr1, addr2, sdk.Coins{sdk.NewInt64Coin("foocoin", 100)})
	header := cmtproto.Header{Height: baseApp.LastBlockHeight() + 1}
	txConfig := moduletestutil.MakeTestTxConfig()
	_, _, err = simtestutil.SignCheckDeliver(t, txConfig, baseApp, header, []sdk.Msg{sendMsg}, "", []uint64{origAccNum}, []uint64{origSeq}, false, false, priv1)
	require.Error(t, err)

	checkBalance(t, baseApp, addr1, sdk.Coins{sdk.NewInt64Coin("foocoin", 67)}, s.BankKeeper)

	ctx2 := baseApp.NewContext(true)
	res2 := s.AccountKeeper.GetAccount(ctx2, addr1)
	require.NotNil(t, res2)

	require.Equal(t, origAccNum, res2.GetAccountNumber())
	require.Equal(t, origSeq+1, res2.GetSequence())
}

func TestMsgSetSendEnabled(t *testing.T) {
	acc1 := authtypes.NewBaseAccountWithAddress(addr1)

	genAccs := []authtypes.GenesisAccount{acc1}
	s := createTestSuite(t, genAccs)

	ctx := s.App.NewContext(false)
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
		false,
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
				"expected gov account as only signer for proposal message",
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
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(tt *testing.T) {
			header := cmtproto.Header{Height: s.App.LastBlockHeight() + 1}
			txGen := moduletestutil.MakeTestTxConfig()
			_, _, err = simtestutil.SignCheckDeliver(tt, txGen, s.App.BaseApp, header, tc.msgs, "", []uint64{0}, tc.accSeqs, tc.expSimPass, tc.expPass, priv1)
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
