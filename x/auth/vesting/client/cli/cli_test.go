package cli_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/gogo/protobuf/proto"

	"github.com/stretchr/testify/suite"

	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	"github.com/cosmos/cosmos-sdk/testutil"
	clitestutil "github.com/cosmos/cosmos-sdk/testutil/cli"

	"github.com/cosmos/cosmos-sdk/testutil/network"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/vesting/client/cli"
	banktestutil "github.com/cosmos/cosmos-sdk/x/bank/client/testutil"
)

type IntegrationTestSuite struct {
	suite.Suite

	cfg     network.Config
	network *network.Network
}

func (s *IntegrationTestSuite) SetupSuite() {
	s.T().Log("setting up integration test suite")

	s.cfg = network.DefaultConfig()
	s.network = network.New(s.T(), s.cfg)

	_, err := s.network.WaitForHeight(1)
	s.Require().NoError(err)

	_, err = s.network.WaitForHeight(1)
	s.Require().NoError(err)
}

func (s *IntegrationTestSuite) TestCreateClawbackVestingAccountCmd() {
	val := s.network.Validators[0]

	info, _, err := val.ClientCtx.Keyring.NewMnemonic("NewCreatePoolAddr",
		keyring.English, sdk.FullFundraiserPath, keyring.DefaultBIP39Passphrase, hd.Secp256k1)
	s.Require().NoError(err)
	newAddr := sdk.AccAddress(info.GetPubKey().Address())
	_, err = banktestutil.MsgSendExec(
		val.ClientCtx,
		val.Address,
		newAddr,
		sdk.NewCoins(sdk.NewInt64Coin(s.cfg.BondDenom, 200000000), sdk.NewInt64Coin("node0token", 20000)), fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
		fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
	)
	s.Require().NoError(err)

	testCases := []struct {
		name         string
		lockupJson   string
		vestingJson  string
		expectErr    bool
		respType     proto.Message
		expectedCode uint32
	}{
		{
			"create vesting account",
			fmt.Sprintf(`
			{
				"start_time": %d,
				"periods":[
				   {
						"length":400,
						"coins":"20stake"
				   }
				]
			 }
			`, time.Now().Unix()),
			fmt.Sprintf(`
			{
				"start_time":%d,
				"periods":[
				   {
					  "coins":"10stake",
					  "length":300
				   },
				   {
					  "coins":"10stake",
					  "length":300
				   }
				]
			 }
			`, time.Now().Unix()),
			false, &sdk.TxResponse{}, 0,
		},
	}

	for _, tc := range testCases {
		tc := tc

		s.Run(tc.name, func() {
			cmd := cli.NewMsgCreateClawbackVestingAccountCmd()
			clientCtx := val.ClientCtx

			lockupJsonFile := testutil.WriteToNewTempFile(s.T(), tc.lockupJson)
			vestingJsonFile := testutil.WriteToNewTempFile(s.T(), tc.vestingJson)

			args := []string{
				newAddr.String(),
				fmt.Sprintf("--%s=%s", cli.FlagLockup, lockupJsonFile.Name()),
				fmt.Sprintf("--%s=%s", cli.FlagVesting, vestingJsonFile.Name()),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, val.Address.String()),
				// common args
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			}

			out, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, args)
			if tc.expectErr {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err, out.String())
				err = clientCtx.Codec.UnmarshalJSON(out.Bytes(), tc.respType)
				s.Require().NoError(err, out.String())

				txResp := tc.respType.(*sdk.TxResponse)
				s.Require().Equal(tc.expectedCode, txResp.Code, out.String())
			}

			cmd = cli.NewMsgClawbackCmd()
			args = []string{
				newAddr.String(),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, val.Address.String()),
				// common args
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			}

			out, err = clitestutil.ExecTestCLICmd(clientCtx, cmd, args)
			s.Require().NoError(err)
			s.Require().NoError(err, out.String)
			err = clientCtx.Codec.UnmarshalJSON(out.Bytes(), tc.respType)
			s.Require().NoError(err, out.String())
			txResp := tc.respType.(*sdk.TxResponse)
			s.Require().Equal(tc.expectedCode, txResp.Code, out.String())
		})
	}
}

func TestIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(IntegrationTestSuite))
}
