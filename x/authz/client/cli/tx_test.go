// +build norace

package cli_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/gogo/protobuf/proto"

	"github.com/stretchr/testify/suite"

	"github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"

	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/testutil"
	clitestutil "github.com/cosmos/cosmos-sdk/testutil/cli"
	"github.com/cosmos/cosmos-sdk/testutil/network"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/authz/client/cli"
	govcli "github.com/cosmos/cosmos-sdk/x/gov/client/cli"
	govtestutil "github.com/cosmos/cosmos-sdk/x/gov/client/testutil"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	stakingcli "github.com/cosmos/cosmos-sdk/x/staking/client/cli"

	authztestutil "github.com/cosmos/cosmos-sdk/x/authz/client/testutil"
	bankcli "github.com/cosmos/cosmos-sdk/x/bank/client/testutil"
	banktestutil "github.com/cosmos/cosmos-sdk/x/bank/client/testutil"
	bank "github.com/cosmos/cosmos-sdk/x/bank/types"
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
	info, _, err := val.ClientCtx.Keyring.NewMnemonic("grantee", keyring.English, sdk.FullFundraiserPath, keyring.DefaultBIP39Passphrase, hd.Secp256k1)
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

var typeMsgSend = bank.SendAuthorization{}.MethodName()
var typeMsgVote = "/cosmos.gov.v1beta1.Msg/Vote"

var commonFlags = []string{}

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
				"send",
				fmt.Sprintf("--%s=100steak", cli.FlagSpendLimit),
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
				"send",
				fmt.Sprintf("--%s=100steak", cli.FlagSpendLimit),
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
				"send",
				fmt.Sprintf("--%s=100steak", cli.FlagSpendLimit),
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
				"generic",
				fmt.Sprintf("--%s=invalid-msg-type", cli.FlagMsgType),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, val.Address.String()),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
				fmt.Sprintf("--%s=%d", cli.FlagExpiration, twoHours),
			},
			nil, 0,
			true,
		},
		{
			"failed with error both validators not allowed",
			[]string{
				grantee.String(),
				"delegate",
				fmt.Sprintf("--%s=100stake", cli.FlagSpendLimit),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, val.Address.String()),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%d", cli.FlagExpiration, twoHours),
				fmt.Sprintf("--%s=%s", cli.FlagAllowedValidators, fmt.Sprintf("%s", val.ValAddress.String())),
				fmt.Sprintf("--%s=%s", cli.FlagDenyValidators, fmt.Sprintf("%s", val.ValAddress.String())),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			nil, 0,
			true,
		},
		{
			"valid tx delegate authorization allowed validators",
			[]string{
				grantee.String(),
				"delegate",
				fmt.Sprintf("--%s=100stake", cli.FlagSpendLimit),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, val.Address.String()),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%d", cli.FlagExpiration, twoHours),
				fmt.Sprintf("--%s=%s", cli.FlagAllowedValidators, fmt.Sprintf("%s", val.ValAddress.String())),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			&sdk.TxResponse{}, 0,
			false,
		},
		{
			"valid tx delegate authorization deny validators",
			[]string{
				grantee.String(),
				"delegate",
				fmt.Sprintf("--%s=100stake", cli.FlagSpendLimit),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, val.Address.String()),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%d", cli.FlagExpiration, twoHours),
				fmt.Sprintf("--%s=%s", cli.FlagDenyValidators, fmt.Sprintf("%s", val.ValAddress.String())),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			&sdk.TxResponse{}, 0,
			false,
		},
		{
			"valid tx undelegate authorization",
			[]string{
				grantee.String(),
				"unbond",
				fmt.Sprintf("--%s=100stake", cli.FlagSpendLimit),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, val.Address.String()),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%d", cli.FlagExpiration, twoHours),
				fmt.Sprintf("--%s=%s", cli.FlagAllowedValidators, fmt.Sprintf("%s", val.ValAddress.String())),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			&sdk.TxResponse{}, 0,
			false,
		},
		{
			"valid tx redelegate authorization",
			[]string{
				grantee.String(),
				"redelegate",
				fmt.Sprintf("--%s=100stake", cli.FlagSpendLimit),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, val.Address.String()),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%d", cli.FlagExpiration, twoHours),
				fmt.Sprintf("--%s=%s", cli.FlagAllowedValidators, fmt.Sprintf("%s", val.ValAddress.String())),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			&sdk.TxResponse{}, 0,
			false,
		},
		{
			"Valid tx send authorization",
			[]string{
				grantee.String(),
				"send",
				fmt.Sprintf("--%s=100steak", cli.FlagSpendLimit),
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
				"generic",
				fmt.Sprintf("--%s=%s", cli.FlagMsgType, typeMsgVote),
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
			out, err := authztestutil.ExecGrantAuthorization(
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

func execDelegate(val *network.Validator, args []string) (testutil.BufferWriter, error) {
	cmd := stakingcli.NewDelegateCmd()
	clientCtx := val.ClientCtx
	return clitestutil.ExecTestCLICmd(clientCtx, cmd, args)
}

func (s *IntegrationTestSuite) TestCmdRevokeAuthorizations() {
	val := s.network.Validators[0]

	grantee := s.grantee
	twoHours := time.Now().Add(time.Minute * time.Duration(120)).Unix()

	// send-authorization
	_, err := authztestutil.ExecGrantAuthorization(
		val,
		[]string{
			grantee.String(),
			"send",
			fmt.Sprintf("--%s=100steak", cli.FlagSpendLimit),
			fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
			fmt.Sprintf("--%s=%s", flags.FlagFrom, val.Address),
			fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
			fmt.Sprintf("--%s=%d", cli.FlagExpiration, twoHours),
			fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
		},
	)
	s.Require().NoError(err)

	// generic-authorization
	_, err = authztestutil.ExecGrantAuthorization(
		val,
		[]string{
			grantee.String(),
			"generic",
			fmt.Sprintf("--%s=%s", cli.FlagMsgType, typeMsgVote),
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

	_, err := authztestutil.ExecGrantAuthorization(
		val,
		[]string{
			grantee.String(),
			"generic",
			fmt.Sprintf("--%s=%s", cli.FlagMsgType, typeMsgVote),
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
	time.Sleep(12 * time.Second)

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

	_, err := authztestutil.ExecGrantAuthorization(
		val,
		[]string{
			grantee.String(),
			"generic",
			fmt.Sprintf("--%s=%s", cli.FlagMsgType, typeMsgVote),
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

	_, err := authztestutil.ExecGrantAuthorization(
		val,
		[]string{
			grantee.String(),
			"send",
			fmt.Sprintf("--%s=12%stoken", cli.FlagSpendLimit, val.Moniker),
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
	execMsg := testutil.WriteToNewTempFile(s.T(), normalGeneratedTx.String())
	testCases := []struct {
		name         string
		args         []string
		expectedCode uint32
		expectErr    bool
	}{
		{
			"valid txn",
			[]string{
				execMsg.Name(),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, grantee.String()),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
			},
			0,
			false,
		},
		{
			"error over spent",
			[]string{
				execMsg.Name(),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, grantee.String()),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
			},
			4,
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
				var response sdk.TxResponse
				s.Require().NoError(err)
				s.Require().NoError(clientCtx.JSONMarshaler.UnmarshalJSON(out.Bytes(), &response), out.String())
				s.Require().Equal(tc.expectedCode, response.Code, out.String())
			}
		})
	}
}

func (s *IntegrationTestSuite) TestExecDelegateAuthorization() {
	val := s.network.Validators[0]
	grantee := s.grantee
	twoHours := time.Now().Add(time.Minute * time.Duration(120)).Unix()

	_, err := authztestutil.ExecGrantAuthorization(
		val,
		[]string{
			grantee.String(),
			"delegate",
			fmt.Sprintf("--%s=100stake", cli.FlagSpendLimit),
			fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
			fmt.Sprintf("--%s=%s", flags.FlagFrom, val.Address.String()),
			fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
			fmt.Sprintf("--%s=%d", cli.FlagExpiration, twoHours),
			fmt.Sprintf("--%s=%s", cli.FlagAllowedValidators, fmt.Sprintf("%s", val.ValAddress.String())),
			fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
		},
	)
	s.Require().NoError(err)

	tokens := sdk.NewCoins(
		sdk.NewCoin("stake", sdk.NewInt(50)),
	)

	delegateTx := fmt.Sprintf(`{"body":{"messages":[{"@type":"/cosmos.staking.v1beta1.Msg/Delegate","delegator_address":"%s","validator_address":"%s","amount":{"denom":"%s","amount":"%s"}}],"memo":"","timeout_height":"0","extension_options":[],"non_critical_extension_options":[]},"auth_info":{"signer_infos":[],"fee":{"amount":[],"gas_limit":"200000","payer":"","granter":""}},"signatures":[]}`, val.Address.String(), val.ValAddress.String(),
		tokens.GetDenomByIndex(0), tokens[0].Amount)
	execMsg := testutil.WriteToNewTempFile(s.T(), delegateTx)

	testCases := []struct {
		name         string
		args         []string
		expectedCode uint32
		expectErr    bool
		errMsg       string
	}{
		{
			"valid txn: (delegate half tokens)",
			[]string{
				execMsg.Name(),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, grantee.String()),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
			},
			0,
			false,
			"",
		},
		{
			"valid txn: (delegate remaining half tokens)",
			[]string{
				execMsg.Name(),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, grantee.String()),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
			},
			0,
			false,
			"",
		},
		{
			"failed with error no authorization found",
			[]string{
				execMsg.Name(),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, grantee.String()),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
			},
			4,
			false,
			"authorization not found",
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
				s.Require().Contains(err.Error(), tc.errMsg)
			} else {
				var response sdk.TxResponse
				s.Require().NoError(err)
				s.Require().NoError(clientCtx.JSONMarshaler.UnmarshalJSON(out.Bytes(), &response), out.String())
				s.Require().Equal(tc.expectedCode, response.Code, out.String())
			}
		})
	}

	//test delegate no spend-limit
	_, err = authztestutil.ExecGrantAuthorization(
		val,
		[]string{
			grantee.String(),
			"delegate",
			fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
			fmt.Sprintf("--%s=%s", flags.FlagFrom, val.Address.String()),
			fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
			fmt.Sprintf("--%s=%d", cli.FlagExpiration, twoHours),
			fmt.Sprintf("--%s=%s", cli.FlagAllowedValidators, fmt.Sprintf("%s", val.ValAddress.String())),
			fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
		},
	)
	s.Require().NoError(err)
	tokens = sdk.NewCoins(
		sdk.NewCoin("stake", sdk.NewInt(50)),
	)

	delegateTx = fmt.Sprintf(`{"body":{"messages":[{"@type":"/cosmos.staking.v1beta1.Msg/Delegate","delegator_address":"%s","validator_address":"%s","amount":{"denom":"%s","amount":"%s"}}],"memo":"","timeout_height":"0","extension_options":[],"non_critical_extension_options":[]},"auth_info":{"signer_infos":[],"fee":{"amount":[],"gas_limit":"200000","payer":"","granter":""}},"signatures":[]}`, val.Address.String(), val.ValAddress.String(),
		tokens.GetDenomByIndex(0), tokens[0].Amount)
	execMsg = testutil.WriteToNewTempFile(s.T(), delegateTx)

	testCases = []struct {
		name         string
		args         []string
		expectedCode uint32
		expectErr    bool
		errMsg       string
	}{
		{
			"valid txn",
			[]string{
				execMsg.Name(),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, grantee.String()),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
			},
			0,
			false,
			"",
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
			0,
			false,
			"",
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
				s.Require().Contains(err.Error(), tc.errMsg)
			} else {
				var response sdk.TxResponse
				s.Require().NoError(err)
				s.Require().NoError(clientCtx.JSONMarshaler.UnmarshalJSON(out.Bytes(), &response), out.String())
				s.Require().Equal(tc.expectedCode, response.Code, out.String())
			}
		})
	}

	// test delegating to denied validator
	_, err = authztestutil.ExecGrantAuthorization(
		val,
		[]string{
			grantee.String(),
			"delegate",
			fmt.Sprintf("--%s=100stake", cli.FlagSpendLimit),
			fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
			fmt.Sprintf("--%s=%s", flags.FlagFrom, val.Address.String()),
			fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
			fmt.Sprintf("--%s=%d", cli.FlagExpiration, twoHours),
			fmt.Sprintf("--%s=%s", cli.FlagDenyValidators, fmt.Sprintf("%s", val.ValAddress.String())),
			fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
		},
	)
	s.Require().NoError(err)

	args := []string{
		execMsg.Name(),
		fmt.Sprintf("--%s=%s", flags.FlagFrom, grantee.String()),
		fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
		fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
		fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
	}
	cmd := cli.NewCmdExecAuthorization()
	out, err := clitestutil.ExecTestCLICmd(val.ClientCtx, cmd, args)
	s.Require().NoError(err)
	s.Contains(out.String(), fmt.Sprintf("cannot delegate/undelegate to %s validator", val.ValAddress.String()))

}

func (s *IntegrationTestSuite) TestExecUndelegateAuthorization() {
	val := s.network.Validators[0]
	grantee := s.grantee
	twoHours := time.Now().Add(time.Minute * time.Duration(120)).Unix()

	// granting undelegate msg authorization
	_, err := authztestutil.ExecGrantAuthorization(
		val,
		[]string{
			grantee.String(),
			"unbond",
			fmt.Sprintf("--%s=100stake", cli.FlagSpendLimit),
			fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
			fmt.Sprintf("--%s=%s", flags.FlagFrom, val.Address.String()),
			fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
			fmt.Sprintf("--%s=%d", cli.FlagExpiration, twoHours),
			fmt.Sprintf("--%s=%s", cli.FlagAllowedValidators, fmt.Sprintf("%s", val.ValAddress.String())),
			fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
		},
	)
	s.Require().NoError(err)

	// delegating stakes to validator
	_, err = execDelegate(
		val,
		[]string{
			val.ValAddress.String(),
			"100stake",
			fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
			fmt.Sprintf("--%s=%s", flags.FlagFrom, val.Address.String()),
			fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
			fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
		},
	)
	s.Require().NoError(err)

	tokens := sdk.NewCoins(
		sdk.NewCoin("stake", sdk.NewInt(50)),
	)

	undelegateTx := fmt.Sprintf(`{"body":{"messages":[{"@type":"/cosmos.staking.v1beta1.Msg/Undelegate","delegator_address":"%s","validator_address":"%s","amount":{"denom":"%s","amount":"%s"}}],"memo":"","timeout_height":"0","extension_options":[],"non_critical_extension_options":[]},"auth_info":{"signer_infos":[],"fee":{"amount":[],"gas_limit":"200000","payer":"","granter":""}},"signatures":[]}`, val.Address.String(), val.ValAddress.String(),
		tokens.GetDenomByIndex(0), tokens[0].Amount)
	execMsg := testutil.WriteToNewTempFile(s.T(), undelegateTx)

	testCases := []struct {
		name         string
		args         []string
		expectedCode uint32
		expectErr    bool
		errMsg       string
	}{
		{
			"valid txn: (undelegate half tokens)",
			[]string{
				execMsg.Name(),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, grantee.String()),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
			},
			0,
			false,
			"",
		},
		{
			"valid txn: (undelegate remaining half tokens)",
			[]string{
				execMsg.Name(),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, grantee.String()),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
			},
			0,
			false,
			"",
		},
		{
			"failed with error no authorization found",
			[]string{
				execMsg.Name(),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, grantee.String()),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
			},
			4,
			false,
			"authorization not found",
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
				s.Require().Contains(err.Error(), tc.errMsg)
			} else {
				var response sdk.TxResponse
				s.Require().NoError(err)
				s.Require().NoError(clientCtx.JSONMarshaler.UnmarshalJSON(out.Bytes(), &response), out.String())
				s.Require().Equal(tc.expectedCode, response.Code, out.String())
			}
		})
	}

	// grant undelegate authorization without limit
	_, err = authztestutil.ExecGrantAuthorization(
		val,
		[]string{
			grantee.String(),
			"unbond",
			fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
			fmt.Sprintf("--%s=%s", flags.FlagFrom, val.Address.String()),
			fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
			fmt.Sprintf("--%s=%d", cli.FlagExpiration, twoHours),
			fmt.Sprintf("--%s=%s", cli.FlagAllowedValidators, fmt.Sprintf("%s", val.ValAddress.String())),
			fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
		},
	)
	s.Require().NoError(err)
	tokens = sdk.NewCoins(
		sdk.NewCoin("stake", sdk.NewInt(50)),
	)

	undelegateTx = fmt.Sprintf(`{"body":{"messages":[{"@type":"/cosmos.staking.v1beta1.Msg/Undelegate","delegator_address":"%s","validator_address":"%s","amount":{"denom":"%s","amount":"%s"}}],"memo":"","timeout_height":"0","extension_options":[],"non_critical_extension_options":[]},"auth_info":{"signer_infos":[],"fee":{"amount":[],"gas_limit":"200000","payer":"","granter":""}},"signatures":[]}`, val.Address.String(), val.ValAddress.String(),
		tokens.GetDenomByIndex(0), tokens[0].Amount)
	execMsg = testutil.WriteToNewTempFile(s.T(), undelegateTx)

	testCases = []struct {
		name         string
		args         []string
		expectedCode uint32
		expectErr    bool
		errMsg       string
	}{
		{
			"valid txn",
			[]string{
				execMsg.Name(),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, grantee.String()),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
			},
			0,
			false,
			"",
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
			0,
			false,
			"",
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
				s.Require().Contains(err.Error(), tc.errMsg)
			} else {
				var response sdk.TxResponse
				s.Require().NoError(err)
				s.Require().NoError(clientCtx.JSONMarshaler.UnmarshalJSON(out.Bytes(), &response), out.String())
				s.Require().Equal(tc.expectedCode, response.Code, out.String())
			}
		})
	}
}

func TestIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(IntegrationTestSuite))
}
