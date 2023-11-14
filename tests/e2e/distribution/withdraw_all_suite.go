package distribution

import (
	"fmt"
	"strings"

	"github.com/stretchr/testify/suite"

	"cosmossdk.io/math"
	"cosmossdk.io/simapp"
	banktypes "cosmossdk.io/x/bank/types"
	"cosmossdk.io/x/distribution/client/cli"
	stakingtypes "cosmossdk.io/x/staking/types"

	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	clitestutil "github.com/cosmos/cosmos-sdk/testutil/cli"
	"github.com/cosmos/cosmos-sdk/testutil/network"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type WithdrawAllTestSuite struct {
	suite.Suite

	cfg     network.Config
	network network.NetworkI
}

func (s *WithdrawAllTestSuite) SetupSuite() {
	cfg := network.DefaultConfig(simapp.NewTestNetworkFixture)
	cfg.NumValidators = 2
	s.cfg = cfg

	s.T().Log("setting up e2e test suite")
	network, err := network.New(s.T(), s.T().TempDir(), s.cfg)
	s.Require().NoError(err)
	s.network = network

	s.Require().NoError(s.network.WaitForNextBlock())
}

// TearDownSuite cleans up the curret test network after _each_ test.
func (s *WithdrawAllTestSuite) TearDownSuite() {
	s.T().Log("tearing down e2e test suite")
	s.network.Cleanup()
}

// This test requires multiple validators, if I add this test to `E2ETestSuite` by increasing
// `NumValidators` the existing tests are leading to non-determnism so created new suite for this test.
func (s *WithdrawAllTestSuite) TestNewWithdrawAllRewardsGenerateOnly() {
	require := s.Require()
	val := s.network.GetValidators()[0]
	val1 := s.network.GetValidators()[1]
	clientCtx := val.GetClientCtx()

	info, _, err := val.GetClientCtx().Keyring.NewMnemonic("newAccount", keyring.English, sdk.FullFundraiserPath, keyring.DefaultBIP39Passphrase, hd.Secp256k1)
	require.NoError(err)

	pubkey, err := info.GetPubKey()
	require.NoError(err)

	newAddr := sdk.AccAddress(pubkey.Address())

	msgSend := &banktypes.MsgSend{
		FromAddress: val.GetAddress().String(),
		ToAddress:   newAddr.String(),
		Amount:      sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, math.NewInt(2000))),
	}
	_, err = clitestutil.SubmitTestTx(
		val.GetClientCtx(),
		msgSend,
		val.GetAddress(),
		clitestutil.TestTxConfig{},
	)

	require.NoError(err)
	require.NoError(s.network.WaitForNextBlock())

	// delegate 500 tokens to validator1
	msg := &stakingtypes.MsgDelegate{
		DelegatorAddress: newAddr.String(),
		ValidatorAddress: val.GetValAddress().String(),
		Amount:           sdk.NewCoin("stake", math.NewInt(500)),
	}

	_, err = clitestutil.SubmitTestTx(val.GetClientCtx(), msg, newAddr, clitestutil.TestTxConfig{})
	require.NoError(err)
	require.NoError(s.network.WaitForNextBlock())

	// delegate 500 tokens to validator2
	msg2 := &stakingtypes.MsgDelegate{
		DelegatorAddress: newAddr.String(),
		ValidatorAddress: val1.GetValAddress().String(),
		Amount:           sdk.NewCoin("stake", math.NewInt(500)),
	}

	_, err = clitestutil.SubmitTestTx(val.GetClientCtx(), msg2, newAddr, clitestutil.TestTxConfig{})
	require.NoError(err)
	require.NoError(s.network.WaitForNextBlock())

	err = s.network.RetryForBlocks(func() error {
		args := []string{
			fmt.Sprintf("--%s=%s", flags.FlagFrom, newAddr.String()),
			fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
			fmt.Sprintf("--%s=true", flags.FlagGenerateOnly),
			fmt.Sprintf("--%s=1", cli.FlagMaxMessagesPerTx),
			fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
			fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, math.NewInt(10))).String()),
		}
		cmd := cli.NewWithdrawAllRewardsCmd()
		out, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, args)
		if err != nil {
			return err
		}

		// expect 2 transactions in the generated file when --max-msgs in a tx set 1.
		txLen := len(strings.Split(strings.Trim(out.String(), "\n"), "\n"))
		if txLen != 2 {
			return fmt.Errorf("expected 2 transactions in the generated file, got %d", txLen)
		}
		return nil
	}, 3)
	require.NoError(err)

	args := []string{
		fmt.Sprintf("--%s=%s", flags.FlagFrom, newAddr.String()),
		fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
		fmt.Sprintf("--%s=true", flags.FlagGenerateOnly),
		fmt.Sprintf("--%s=2", cli.FlagMaxMessagesPerTx),
		fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
		fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, math.NewInt(10))).String()),
	}
	cmd := cli.NewWithdrawAllRewardsCmd()
	out, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, args)
	require.NoError(err)
	// expect 1 transaction in the generated file when --max-msgs in a tx set 2, since there are only delegations.
	s.Require().Equal(1, len(strings.Split(strings.Trim(out.String(), "\n"), "\n")))
}
