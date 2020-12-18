package cli_test

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/gogo/protobuf/proto"

	"github.com/stretchr/testify/suite"

	"github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	tmcli "github.com/tendermint/tendermint/libs/cli"

	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/testutil"
	clitestutil "github.com/cosmos/cosmos-sdk/testutil/cli"
	"github.com/cosmos/cosmos-sdk/testutil/network"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/authz/client/cli"
	govcli "github.com/cosmos/cosmos-sdk/x/gov/client/cli"
	govtestutil "github.com/cosmos/cosmos-sdk/x/gov/client/testutil"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"

	"github.com/cosmos/cosmos-sdk/x/authz/types"
	bankcli "github.com/cosmos/cosmos-sdk/x/bank/client/testutil"
	banktestutil "github.com/cosmos/cosmos-sdk/x/bank/client/testutil"
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

	// create a proposal with deposit
	_, err = govtestutil.MsgSubmitProposal(val.ClientCtx, val.Address.String(),
		"Text Proposal 1", "Where is the title!?", govtypes.ProposalTypeText,
		fmt.Sprintf("--%s=%s", govcli.FlagDeposit, sdk.NewCoin(s.cfg.BondDenom, govtypes.DefaultMinDepositTokens).String()))
	s.Require().NoError(err)

	_, err = s.network.WaitForHeight(1)
	s.Require().NoError(err)
}

func (s *IntegrationTestSuite) TearDownSuite() {
	s.T().Log("tearing down integration test suite")
	s.network.Cleanup()
}

var typeMsgSend = types.SendAuthorization{}.MethodName()
var typeMsgVote = "/cosmos.gov.v1beta1.Msg/Vote"

func (s *IntegrationTestSuite) TestQueryAuthorizations() {
	val := s.network.Validators[0]

	grantee := s.grantee
	twoHours := time.Now().Add(time.Minute * time.Duration(120)).Unix()

	_, err := execGrantAuthorization(
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
		name      string
		args      []string
		expectErr bool
		expErrMsg string
	}{
		{
			"Error: Invalid grantee",
			[]string{
				val.Address.String(),
				"invalid grantee",
				fmt.Sprintf("--%s=json", tmcli.OutputFlag),
			},
			true,
			"decoding bech32 failed: invalid character in string: ' '",
		},
		{
			"Error: Invalid granter",
			[]string{
				"invalid granter",
				grantee.String(),
				fmt.Sprintf("--%s=json", tmcli.OutputFlag),
			},
			true,
			"decoding bech32 failed: invalid character in string: ' '",
		},
		{
			"Valid txn (json)",
			[]string{
				val.Address.String(),
				grantee.String(),
				fmt.Sprintf("--%s=json", tmcli.OutputFlag),
			},
			false,
			``,
		},
	}
	for _, tc := range testCases {
		tc := tc

		s.Run(tc.name, func() {
			cmd := cli.GetCmdQueryAuthorizations()
			clientCtx := val.ClientCtx
			resp, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, tc.args)
			if tc.expectErr {
				s.Require().Error(err)
				s.Require().Contains(string(resp.Bytes()), tc.expErrMsg)
			} else {
				s.Require().NoError(err)
				var grants types.QueryAuthorizationsResponse
				err = val.ClientCtx.JSONMarshaler.UnmarshalJSON(resp.Bytes(), &grants)
				s.Require().NoError(err)
			}
		})
	}
}

func (s *IntegrationTestSuite) TestQueryAuthorization() {
	val := s.network.Validators[0]

	grantee := s.grantee
	twoHours := time.Now().Add(time.Minute * time.Duration(120)).Unix()

	_, err := execGrantAuthorization(
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
				val.Address.String(),
				grantee.String(),
				typeMsgSend,
				fmt.Sprintf("--%s=json", tmcli.OutputFlag),
			},
			false,
			`{"@type":"/cosmos.authz.v1beta1.SendAuthorization","spend_limit":[{"denom":"steak","amount":"100"}]}`,
		},
	}
	for _, tc := range testCases {
		tc := tc

		s.Run(tc.name, func() {
			cmd := cli.GetCmdQueryAuthorization()
			clientCtx := val.ClientCtx
			out, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, tc.args)
			if tc.expectErr {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)
				s.Require().Contains(strings.TrimSpace(out.String()), tc.expectedOutput)
			}
		})
	}
}

func (s *IntegrationTestSuite) TestCLITxGrantAuthorization() {
	val := s.network.Validators[0]
	grantee := s.grantee

	twoHours := time.Now().Add(time.Minute * time.Duration(120)).Unix()
	pastHour := time.Now().Add(time.Minute * time.Duration(-60)).Unix()

	testCases := []struct {
		name         string
		args         []string
		respType     proto.Message
		expectedCode uint32
		expectErr    bool
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
		},
		{
			"fail with error invalid msg-type",
			[]string{
				grantee.String(),
				"invalid-msg-type",
				fmt.Sprintf("--%s=%s", flags.FlagFrom, val.Address.String()),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%d", cli.FlagExpiration, twoHours),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			&sdk.TxResponse{}, 29,
			false,
		},
		{
			"Valid tx send authorization",
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
		},
		{
			"Valid tx generic authorization",
			[]string{
				grantee.String(),
				typeMsgVote,
				fmt.Sprintf("--%s=%s", flags.FlagFrom, val.Address.String()),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%d", cli.FlagExpiration, twoHours),
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
			clientCtx := val.ClientCtx
			out, err := execGrantAuthorization(
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

func execGrantAuthorization(val *network.Validator, args []string) (testutil.BufferWriter, error) {
	cmd := cli.NewCmdGrantAuthorization()
	clientCtx := val.ClientCtx
	return clitestutil.ExecTestCLICmd(clientCtx, cmd, args)
}

func (s *IntegrationTestSuite) TestCmdRevokeAuthorizations() {
	val := s.network.Validators[0]

	grantee := s.grantee
	twoHours := time.Now().Add(time.Minute * time.Duration(120)).Unix()

	// send-authorization
	_, err := execGrantAuthorization(
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

	// generic-authorization
	_, err = execGrantAuthorization(
		val,
		[]string{
			grantee.String(),
			typeMsgVote,
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
			"Valid tx send authorization",
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
		{
			"Valid tx generic authorization",
			[]string{
				grantee.String(),
				typeMsgVote,
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

func (s *IntegrationTestSuite) TestExecAuthorizationWithExpiration() {
	val := s.network.Validators[0]
	grantee := s.grantee
	tenSeconds := time.Now().Add(time.Second * time.Duration(10)).Unix()

	_, err := execGrantAuthorization(
		val,
		[]string{
			grantee.String(),
			typeMsgVote,
			fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
			fmt.Sprintf("--%s=%s", flags.FlagFrom, val.Address.String()),
			fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
			fmt.Sprintf("--%s=%d", cli.FlagExpiration, tenSeconds),
			fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
		},
	)
	s.Require().NoError(err)
	// msg vote
	voteTx := fmt.Sprintf(`{"body":{"messages":[{"@type":"/cosmos.gov.v1beta1.Msg/Vote","proposal_id":"1","voter":"%s","option":"VOTE_OPTION_YES"}],"memo":"","timeout_height":"0","extension_options":[],"non_critical_extension_options":[]},"auth_info":{"signer_infos":[],"fee":{"amount":[],"gas_limit":"200000","payer":"","granter":""}},"signatures":[]}`, val.Address.String())
	execMsg := testutil.WriteToNewTempFile(s.T(), voteTx)

	// waiting for authorization to expires
	time.Sleep(15 * time.Second)

	cmd := cli.NewCmdExecAuthorization()
	clientCtx := val.ClientCtx

	res, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, []string{
		execMsg.Name(),
		fmt.Sprintf("--%s=%s", flags.FlagFrom, grantee.String()),
		fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
		fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
		fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
	})
	s.Require().NoError(err)
	s.Require().Contains(res.String(), "authorization not found")

}

func (s *IntegrationTestSuite) TestNewExecGenericAuthorized() {
	val := s.network.Validators[0]
	grantee := s.grantee
	twoHours := time.Now().Add(time.Minute * time.Duration(120)).Unix()

	_, err := execGrantAuthorization(
		val,
		[]string{
			grantee.String(),
			typeMsgVote,
			fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
			fmt.Sprintf("--%s=%s", flags.FlagFrom, val.Address.String()),
			fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
			fmt.Sprintf("--%s=%d", cli.FlagExpiration, twoHours),
			fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
		},
	)
	s.Require().NoError(err)

	// msg vote
	voteTx := fmt.Sprintf(`{"body":{"messages":[{"@type":"/cosmos.gov.v1beta1.Msg/Vote","proposal_id":"1","voter":"%s","option":"VOTE_OPTION_YES"}],"memo":"","timeout_height":"0","extension_options":[],"non_critical_extension_options":[]},"auth_info":{"signer_infos":[],"fee":{"amount":[],"gas_limit":"200000","payer":"","granter":""}},"signatures":[]}`, val.Address.String())
	execMsg := testutil.WriteToNewTempFile(s.T(), voteTx)

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

			cmd := cli.NewCmdExecAuthorization()
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

func (s *IntegrationTestSuite) TestNewExecGrantAuthorized() {
	val := s.network.Validators[0]

	grantee := s.grantee

	twoHours := time.Now().Add(time.Minute * time.Duration(120)).Unix()

	_, err := execGrantAuthorization(
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
	normalGeneratedTx, err := bankcli.ServiceMsgSendExec(
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
	execMsg := testutil.WriteToNewTempFile(s.T(), normalGeneratedTx.String())
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

			cmd := cli.NewCmdExecAuthorization()
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
