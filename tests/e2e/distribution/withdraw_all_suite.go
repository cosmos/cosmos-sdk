package distribution

import (
	"fmt"
	"strings"

	"cosmossdk.io/simapp"
	"github.com/stretchr/testify/suite"

	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	clitestutil "github.com/cosmos/cosmos-sdk/testutil/cli"
	"github.com/cosmos/cosmos-sdk/testutil/network"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/distribution/client/cli"
	stakingcli "github.com/cosmos/cosmos-sdk/x/staking/client/cli"
)

type WithdrawAllTestSuite struct {
	suite.Suite

	cfg     network.Config
	network *network.Network
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
	val := s.network.Validators[0]
	val1 := s.network.Validators[1]
	clientCtx := val.ClientCtx

	info, _, err := val.ClientCtx.Keyring.NewMnemonic("newAccount", keyring.English, sdk.FullFundraiserPath, keyring.DefaultBIP39Passphrase, hd.Secp256k1)
	require.NoError(err)

	pubkey, err := info.GetPubKey()
	require.NoError(err)

	newAddr := sdk.AccAddress(pubkey.Address())
	_, err = clitestutil.MsgSendExec(
		val.ClientCtx,
		val.Address,
		newAddr,
		sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(2000))), fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
		fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
		fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
	)
	require.NoError(err)
	require.NoError(s.network.WaitForNextBlock())

	// delegate 500 tokens to validator1
	args := []string{
		val.ValAddress.String(),
		sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(500)).String(),
		fmt.Sprintf("--%s=%s", flags.FlagFrom, newAddr.String()),
		fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
		fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
		fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
	}
	cmd := stakingcli.NewDelegateCmd()
	_, err = clitestutil.ExecTestCLICmd(clientCtx, cmd, args)
	require.NoError(err)
	require.NoError(s.network.WaitForNextBlock())

	// delegate 500 tokens to validator2
	args = []string{
		val1.ValAddress.String(),
		sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(500)).String(),
		fmt.Sprintf("--%s=%s", flags.FlagFrom, newAddr.String()),
		fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
		fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
		fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
	}
	_, err = clitestutil.ExecTestCLICmd(clientCtx, cmd, args)
	require.NoError(err)
	require.NoError(s.network.WaitForNextBlock())

	args = []string{
		fmt.Sprintf("--%s=%s", flags.FlagFrom, newAddr.String()),
		fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
		fmt.Sprintf("--%s=true", flags.FlagGenerateOnly),
		fmt.Sprintf("--%s=1", cli.FlagMaxMessagesPerTx),
		fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
		fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
	}
	cmd = cli.NewWithdrawAllRewardsCmd()
	out, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, args)
	require.NoError(err)
	// expect 2 transactions in the generated file when --max-msgs in a tx set 1.
	s.Require().Equal(2, len(strings.Split(strings.Trim(out.String(), "\n"), "\n")))

	args = []string{
		fmt.Sprintf("--%s=%s", flags.FlagFrom, newAddr.String()),
		fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
		fmt.Sprintf("--%s=true", flags.FlagGenerateOnly),
		fmt.Sprintf("--%s=2", cli.FlagMaxMessagesPerTx),
		fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
		fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
	}
	cmd = cli.NewWithdrawAllRewardsCmd()
	out, err = clitestutil.ExecTestCLICmd(clientCtx, cmd, args)
	require.NoError(err)
	// expect 1 transaction in the generated file when --max-msgs in a tx set 2, since there are only delegations.
	s.Require().Equal(1, len(strings.Split(strings.Trim(out.String(), "\n"), "\n")))
}
