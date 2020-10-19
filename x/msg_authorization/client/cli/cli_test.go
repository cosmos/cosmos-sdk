// +build norace

package cli_test

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/gogo/protobuf/proto"
	"github.com/spf13/viper"

	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/testutil"
	clitestutil "github.com/cosmos/cosmos-sdk/testutil/cli"
	"github.com/cosmos/cosmos-sdk/testutil/network"
	sdk "github.com/cosmos/cosmos-sdk/types"
	bankcli "github.com/cosmos/cosmos-sdk/x/bank/client/testutil"
	banktestutil "github.com/cosmos/cosmos-sdk/x/bank/client/testutil"
	"github.com/cosmos/cosmos-sdk/x/msg_authorization/client/cli"
	"github.com/cosmos/cosmos-sdk/x/msg_authorization/types"
	"github.com/stretchr/testify/suite"

	"github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	tmcli "github.com/tendermint/tendermint/libs/cli"
)

type IntegrationTestSuite struct {
	suite.Suite

	cfg     network.Config
	network *network.Network
	grantee sdk.AccAddress
}

func (s *IntegrationTestSuite) SetupSuite() {
	s.T().Log("setting up integration test suite")

	cfg := network.DefaultConfig()
	cfg.NumValidators = 1

	s.cfg = cfg
	s.network = network.New(s.T(), cfg)

	val := s.network.Validators[0]

	// Create new account in the keyring.
	info, _, err := val.ClientCtx.Keyring.NewMnemonic("grantee", keyring.English, sdk.FullFundraiserPath, hd.Secp256k1)
	s.Require().NoError(err)
	newAddr := sdk.AccAddress(info.GetPubKey().Address())

	// Send some funds to the new account.
	_, err = banktestutil.MsgSendExec(
		val.ClientCtx,
		val.Address,
		newAddr,
		sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(200))), fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
		fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
		fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
	)
	s.Require().NoError(err)
	s.grantee = newAddr

	_, err = s.network.WaitForHeight(1)
	s.Require().NoError(err)
}

func (s *IntegrationTestSuite) TearDownSuite() {
	s.T().Log("tearing down integration test suite")
	s.network.Cleanup()
}

var typeMsgSend = types.SendAuthorization{}.MethodName()

func (s *IntegrationTestSuite) TestCLITxGrantAuthorization() {
	val := s.network.Validators[0]
	grantee := s.grantee

	twoHours := 60 * 2  // In seconds
	pastHour := 60 * -1 // In seconds

	testCases := []struct {
		name         string
		args         []string
		respType     proto.Message
		expectedCode uint32
		expectErr    bool
		expiration   int
	}{
		{
			"Invalid granter Address",
			[]string{
				"grantee_addr",
				typeMsgSend,
				"100steak",
				fmt.Sprintf("--%s=%s", flags.FlagFrom, "granter"),
				fmt.Sprintf("--%s=true", flags.FlagGenerateOnly),
				fmt.Sprintf("--%s=%d", cli.FlagExpiration, twoHours),
			},
			nil,
			0,
			true,
			twoHours,
		},
		{
			"Invalid grantee Address",
			[]string{
				"grantee_addr",
				typeMsgSend,
				"100steak",
				fmt.Sprintf("--%s=%s", flags.FlagFrom, val.Address.String()),
				fmt.Sprintf("--%s=true", flags.FlagGenerateOnly),
				fmt.Sprintf("--%s=%d", cli.FlagExpiration, twoHours),
			},
			nil, 0,
			true,
			twoHours,
		},
		{
			"Invalid expiration time",
			[]string{
				grantee.String(),
				typeMsgSend,
				"100steak",
				fmt.Sprintf("--%s=%s", flags.FlagFrom, val.Address.String()),
				fmt.Sprintf("--%s=true", flags.FlagGenerateOnly),
				fmt.Sprintf("--%s=%d", cli.FlagExpiration, pastHour),
			},
			nil, 0,
			true,
			pastHour,
		},
		{
			"Valid tx",
			[]string{
				grantee.String(),
				typeMsgSend,
				"100steak",
				fmt.Sprintf("--%s=%s", flags.FlagFrom, val.Address.String()),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%d", cli.FlagExpiration, twoHours),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			&sdk.TxResponse{}, 0,
			false,
			twoHours,
		},
	}

	for _, tc := range testCases {
		tc := tc
		s.Run(tc.name, func() {
			clientCtx := val.ClientCtx
			viper.Set(cli.FlagExpiration, tc.expiration)

			out, err := execGrantSendAuthorization(
				val,
				tc.args,
			)
			if tc.expectErr {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)
				s.Require().NoError(clientCtx.JSONMarshaler.UnmarshalJSON(out.Bytes(), tc.respType), out.String())
				txResp := tc.respType.(*sdk.TxResponse)
				s.Require().Equal(tc.expectedCode, txResp.Code, out.String())
			}
		})
	}

}

func (s *IntegrationTestSuite) TestQueryAuthorizations() {
	val := s.network.Validators[0]

	grantee := s.grantee
	twoHours := 60 * 2 // In seconds
	viper.Set(cli.FlagExpiration, twoHours)

	_, err := execGrantSendAuthorization(
		val,
		[]string{
			grantee.String(),
			typeMsgSend,
			"100steak",
			fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
			fmt.Sprintf("--%s=%s", flags.FlagFrom, val.Address),
			fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
			fmt.Sprintf("--%s=%d", cli.FlagExpiration, twoHours),
			fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
		},
	)
	s.Require().NoError(err)

	testCases := []struct {
		name           string
		args           []string
		expectErr      bool
		expectedOutput string
	}{
		{
			"Error: Invalid grantee",
			[]string{
				"authorization",
				val.Address.String(),
				"invalid grantee",
				typeMsgSend,
				fmt.Sprintf("--%s=json", tmcli.OutputFlag),
			},
			true,
			"",
		},
		{
			"Error: Invalid granter",
			[]string{
				"authorization",
				"invalid granter",
				grantee.String(),
				typeMsgSend,
				fmt.Sprintf("--%s=json", tmcli.OutputFlag),
			},
			true,
			"",
		},
		{
			"no authorization found",
			[]string{
				"authorization",
				val.Address.String(),
				grantee.String(),
				"typeMsgSend",
				fmt.Sprintf("--%s=json", tmcli.OutputFlag),
			},
			true,
			"",
		},
		{
			"Valid txn (json)",
			[]string{
				"authorization",
				val.Address.String(),
				grantee.String(),
				typeMsgSend,
				fmt.Sprintf("--%s=json", tmcli.OutputFlag),
			},
			false,
			`{"@type":"/cosmos.msg_authorization.v1beta1.SendAuthorization","spend_limit":[{"denom":"steak","amount":"100"}]}`,
		},
	}
	for _, tc := range testCases {
		tc := tc

		s.Run(tc.name, func() {
			cmd := cli.GetQueryCmd()
			clientCtx := val.ClientCtx
			out, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, tc.args)
			if tc.expectErr {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)
				s.Require().Equal(tc.expectedOutput, strings.TrimSpace(out.String()))
			}
		})
	}
}

func execGrantSendAuthorization(val *network.Validator, args []string) (testutil.BufferWriter, error) {
	cmd := cli.NewCmdGrantAuthorization()
	clientCtx := val.ClientCtx
	return clitestutil.ExecTestCLICmd(clientCtx, cmd, args)
}

func (s *IntegrationTestSuite) TestCmdRevokeAuthorizations() {
	val := s.network.Validators[0]

	grantee := s.grantee
	twoHours := 60 * 2 // In seconds

	viper.Set(cli.FlagExpiration, twoHours)

	_, err := execGrantSendAuthorization(
		val,
		[]string{
			grantee.String(),
			typeMsgSend,
			"100steak",
			fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
			fmt.Sprintf("--%s=%s", flags.FlagFrom, val.Address),
			fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
			fmt.Sprintf("--%s=%d", cli.FlagExpiration, twoHours),
			fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
		},
	)
	s.Require().NoError(err)
	testCases := []struct {
		name         string
		args         []string
		respType     proto.Message
		expectedCode uint32
		expectErr    bool
	}{
		{
			"invalid grantee address",
			[]string{
				"invlid grantee",
				typeMsgSend,
				fmt.Sprintf("--%s=%s", flags.FlagFrom, val.Address.String()),
				fmt.Sprintf("--%s=true", flags.FlagGenerateOnly),
			},
			nil,
			0,
			true,
		},
		{
			"invalid granter address",
			[]string{
				grantee.String(),
				typeMsgSend,
				fmt.Sprintf("--%s=%s", flags.FlagFrom, "granter"),
				fmt.Sprintf("--%s=true", flags.FlagGenerateOnly),
			},
			nil,
			0,
			true,
		},
		{
			"Valid tx",
			[]string{
				grantee.String(),
				typeMsgSend,
				fmt.Sprintf("--%s=%s", flags.FlagFrom, val.Address.String()),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			&sdk.TxResponse{}, 0,
			false,
		},
	}
	for _, tc := range testCases {
		tc := tc
		s.Run(tc.name, func() {
			cmd := cli.NewCmdRevokeAuthorization()
			clientCtx := val.ClientCtx

			out, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, tc.args)
			if tc.expectErr {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)
				s.Require().NoError(clientCtx.JSONMarshaler.UnmarshalJSON(out.Bytes(), tc.respType), out.String())

				txResp := tc.respType.(*sdk.TxResponse)
				s.Require().Equal(tc.expectedCode, txResp.Code, out.String())
			}
		})
	}
}

func (s *IntegrationTestSuite) TestNewExecAuthorized() {
	val := s.network.Validators[0]

	grantee := s.grantee

	now := time.Now()
	twoHours := now.Add(time.Minute * time.Duration(120)).Unix()

	viper.Set(cli.FlagExpiration, twoHours)
	_, err := execGrantSendAuthorization(
		val,
		[]string{
			grantee.String(),
			typeMsgSend,
			fmt.Sprintf("12%stoken", val.Moniker),
			fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
			fmt.Sprintf("--%s=%s", flags.FlagFrom, val.Address.String()),
			fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
			fmt.Sprintf("--%s=%d", cli.FlagExpiration, twoHours),
			fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
		},
	)
	s.Require().NoError(err)
	tokens := sdk.NewCoins(
		sdk.NewCoin(fmt.Sprintf("%stoken", val.Moniker), sdk.NewInt(12)),
	)
	normalGeneratedTx, err := bankcli.MsgSendExec(
		val.ClientCtx,
		val.Address,
		grantee,
		tokens,
		fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
		fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
		fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
		fmt.Sprintf("--%s=true", flags.FlagGenerateOnly),
	)
	s.Require().NoError(err)
	execMsg, cleanup1 := testutil.WriteToNewTempFile(s.T(), normalGeneratedTx.String())
	defer cleanup1()
	testCases := []struct {
		name         string
		args         []string
		respType     proto.Message
		expectedCode uint32
		expectErr    bool
	}{
		{
			"fail invalid grantee",
			[]string{
				execMsg.Name(),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, "grantee"),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=true", flags.FlagGenerateOnly),
			},
			nil,
			0,
			true,
		},
		{
			"fail invalid json path",
			[]string{
				"/invalid/file.txt",
				fmt.Sprintf("--%s=%s", flags.FlagFrom, grantee.String()),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
			},
			nil,
			0,
			true,
		},
		{
			"valid txn",
			[]string{
				execMsg.Name(),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, grantee.String()),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
			},
			&sdk.TxResponse{},
			0,
			false,
		},
	}

	for _, tc := range testCases {
		tc := tc
		s.Run(tc.name, func() {

			cmd := cli.NewCmdSendAs()
			clientCtx := val.ClientCtx

			out, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, tc.args)
			if tc.expectErr {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)
				s.Require().NoError(clientCtx.JSONMarshaler.UnmarshalJSON(out.Bytes(), tc.respType), out.String())
				txResp := tc.respType.(*sdk.TxResponse)
				s.Require().Equal(tc.expectedCode, txResp.Code, out.String())
			}
		})
	}
}

func TestIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(IntegrationTestSuite))
}
