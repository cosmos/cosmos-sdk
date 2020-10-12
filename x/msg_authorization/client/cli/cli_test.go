package cli_test

import (
	"fmt"
	"strings"
	"testing"
	"time"

	tmcli "github.com/tendermint/tendermint/libs/cli"

	clitestutil "github.com/cosmos/cosmos-sdk/testutil/cli"
	"github.com/gogo/protobuf/proto"
	"github.com/spf13/viper"

	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	"github.com/cosmos/cosmos-sdk/testutil"
	"github.com/cosmos/cosmos-sdk/testutil/network"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/msg_authorization/client/cli"
	msgauthcli "github.com/cosmos/cosmos-sdk/x/msg_authorization/client/test_util"
	"github.com/cosmos/cosmos-sdk/x/msg_authorization/types"
	"github.com/stretchr/testify/suite"
)

type IntegrationTestSuite struct {
	suite.Suite

	cfg     network.Config
	network *network.Network
}

func (s *IntegrationTestSuite) SetupSuite() {
	s.T().Log("setting up integration test suite")

	cfg := network.DefaultConfig()
	cfg.NumValidators = 1

	s.cfg = cfg
	s.network = network.New(s.T(), cfg)

	kb := s.network.Validators[0].ClientCtx.Keyring
	_, _, err := kb.NewMnemonic("newAccount", keyring.English, sdk.FullFundraiserPath, hd.Secp256k1)
	s.Require().NoError(err)

	_, _, err = kb.NewMnemonic("grantee", keyring.English, sdk.FullFundraiserPath, hd.Secp256k1)
	s.Require().NoError(err)

	_, err = s.network.WaitForHeight(1)
	s.Require().NoError(err)
}

func (s *IntegrationTestSuite) TearDownSuite() {
	s.T().Log("tearing down integration test suite")
	s.network.Cleanup()
}

var typeMsgSend = types.SendAuthorization{}.MsgType()

func (s *IntegrationTestSuite) TestCLITxGrantAuthorization() {
	val := s.network.Validators[0]
	grantee, err := val.ClientCtx.Keyring.Key("grantee")
	s.Require().NoError(err)

	sendAuth := `{"spend_limit":[{"denom":"steak","amount":"100"}]}`

	sendAuthFile, cleanup := testutil.WriteToNewTempFile(s.T(), sendAuth)
	defer cleanup()
	now := time.Now()
	twoHours := now.Add(time.Minute * time.Duration(120)).Unix()
	pastHour := now.Add(time.Minute * time.Duration(-60)).Unix()

	testCases := []struct {
		name          string
		grantee       string
		msgType       string
		authorization string
		args          []string
		respType      proto.Message
		expectedCode  uint32
		expectErr     bool
		expiration    int64
	}{
		{
			"Invalid granter Address",
			"grantee_addr",
			typeMsgSend,
			sendAuthFile.Name(),
			[]string{
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
			"grantee_addr",
			typeMsgSend,
			sendAuthFile.Name(),
			[]string{
				fmt.Sprintf("--%s=%s", flags.FlagFrom, val.Address.String()),
				fmt.Sprintf("--%s=true", flags.FlagGenerateOnly),
				fmt.Sprintf("--%s=%d", cli.FlagExpiration, twoHours),
			},
			nil, 0,
			true,
			twoHours,
		},
		{
			"Invalid authorization path",
			grantee.GetAddress().String(),
			typeMsgSend,
			"/tmp/invalid/file/path",
			[]string{
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
			grantee.GetAddress().String(),
			typeMsgSend,
			sendAuthFile.Name(),
			[]string{
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
			grantee.GetAddress().String(),
			typeMsgSend,
			sendAuthFile.Name(),
			[]string{
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

			out, err := msgauthcli.MsgGrantSendAuthorizationExec(clientCtx, tc.grantee,
				tc.msgType, tc.authorization, tc.args...)
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

	grantee, err := val.ClientCtx.Keyring.Key("grantee")
	s.Require().NoError(err)
	s.T().Log(grantee.GetAddress().String())

	sendAuth := `{"spend_limit":[{"denom":"steak","amount":"100"}]}`

	sendAuthFile, cleanup := testutil.WriteToNewTempFile(s.T(), sendAuth)
	defer cleanup()
	now := time.Now()
	twoHours := now.Add(time.Minute * time.Duration(120)).Unix()

	viper.Set(cli.FlagExpiration, twoHours)
	_, err = msgauthcli.MsgGrantSendAuthorizationExec(
		val.ClientCtx,
		grantee.GetAddress().String(),
		typeMsgSend,
		sendAuthFile.Name(),
		fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
		fmt.Sprintf("--%s=%s", flags.FlagFrom, val.Address),
		fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
		fmt.Sprintf("--%s=%d", cli.FlagExpiration, twoHours),
		fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
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
				grantee.GetAddress().String(),
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
				grantee.GetAddress().String(),
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
				grantee.GetAddress().String(),
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
			cmd := cli.GetQueryCmd(types.StoreKey)
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

func (s *IntegrationTestSuite) TestCmdRevokeAuthorizations() {
	val := s.network.Validators[0]

	grantee, err := val.ClientCtx.Keyring.Key("grantee")
	s.Require().NoError(err)
	s.T().Log(grantee.GetAddress().String())

	sendAuth := `{"spend_limit":[{"denom":"steak","amount":"100"}]}`

	sendAuthFile, cleanup := testutil.WriteToNewTempFile(s.T(), sendAuth)
	defer cleanup()
	now := time.Now()
	twoHours := now.Add(time.Minute * time.Duration(120)).Unix()

	viper.Set(cli.FlagExpiration, twoHours)
	_, err = msgauthcli.MsgGrantSendAuthorizationExec(
		val.ClientCtx,
		grantee.GetAddress().String(),
		typeMsgSend,
		sendAuthFile.Name(),
		fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
		fmt.Sprintf("--%s=%s", flags.FlagFrom, val.Address),
		fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
		fmt.Sprintf("--%s=%d", cli.FlagExpiration, twoHours),
		fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
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
				grantee.GetAddress().String(),
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
				grantee.GetAddress().String(),
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
			cmd := cli.GetCmdRevokeAuthorization(types.StoreKey)
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
