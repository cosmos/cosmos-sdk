package authz

import (
	"fmt"
	"time"

	"github.com/cosmos/gogoproto/proto"
	"github.com/stretchr/testify/suite"

	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/codec/address"
	"github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	"github.com/cosmos/cosmos-sdk/testutil"
	clitestutil "github.com/cosmos/cosmos-sdk/testutil/cli"
	"github.com/cosmos/cosmos-sdk/testutil/network"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authcli "github.com/cosmos/cosmos-sdk/x/auth/client/cli"
	"github.com/cosmos/cosmos-sdk/x/authz"
	"github.com/cosmos/cosmos-sdk/x/authz/client/cli"
	authzclitestutil "github.com/cosmos/cosmos-sdk/x/authz/client/testutil"
	bank "github.com/cosmos/cosmos-sdk/x/bank/types"
	govcli "github.com/cosmos/cosmos-sdk/x/gov/client/cli"
	govtestutil "github.com/cosmos/cosmos-sdk/x/gov/client/testutil"
	govv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	govv1beta1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
	stakingcli "github.com/cosmos/cosmos-sdk/x/staking/client/cli"
)

type E2ETestSuite struct {
	suite.Suite

	cfg     network.Config
	network *network.Network
	grantee []sdk.AccAddress
}

func NewE2ETestSuite(cfg network.Config) *E2ETestSuite {
	return &E2ETestSuite{cfg: cfg}
}

func (s *E2ETestSuite) SetupSuite() {
	s.T().Log("setting up e2e test suite")

	var err error
	s.network, err = network.New(s.T(), s.T().TempDir(), s.cfg)
	s.Require().NoError(err)

	val := s.network.Validators[0]
	s.grantee = make([]sdk.AccAddress, 6)

	// Send some funds to the new account.
	// Create new account in the keyring.
	s.grantee[0] = s.createAccount("grantee1")
	s.msgSendExec(s.grantee[0])

	// create a proposal with deposit
	_, err = govtestutil.MsgSubmitLegacyProposal(val.ClientCtx, val.Address.String(),
		"Text Proposal 1", "Where is the title!?", govv1beta1.ProposalTypeText,
		fmt.Sprintf("--%s=%s", govcli.FlagDeposit, sdk.NewCoin(s.cfg.BondDenom, govv1.DefaultMinDepositTokens).String()))
	s.Require().NoError(err)
	s.Require().NoError(s.network.WaitForNextBlock())

	// Create new account in the keyring.
	s.grantee[1] = s.createAccount("grantee2")
	// Send some funds to the new account.
	s.msgSendExec(s.grantee[1])

	// grant send authorization to grantee2
	out, err := authzclitestutil.CreateGrant(val.ClientCtx, []string{
		s.grantee[1].String(),
		"send",
		fmt.Sprintf("--%s=100stake", cli.FlagSpendLimit),
		fmt.Sprintf("--%s=%s", flags.FlagFrom, val.Address.String()),
		fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
		fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
		fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(10))).String()),
		fmt.Sprintf("--%s=%d", cli.FlagExpiration, time.Now().Add(time.Minute*time.Duration(120)).Unix()),
	})
	s.Require().NoError(err)
	s.Require().NoError(s.network.WaitForNextBlock())
	var response sdk.TxResponse
	s.Require().NoError(val.ClientCtx.Codec.UnmarshalJSON(out.Bytes(), &response), out.String())
	s.Require().NoError(clitestutil.CheckTxCode(s.network, val.ClientCtx, response.TxHash, 0))

	// Create new account in the keyring.
	s.grantee[2] = s.createAccount("grantee3")

	// grant send authorization to grantee3
	_, err = authzclitestutil.CreateGrant(val.ClientCtx, []string{
		s.grantee[2].String(),
		"send",
		fmt.Sprintf("--%s=100stake", cli.FlagSpendLimit),
		fmt.Sprintf("--%s=%s", flags.FlagFrom, val.Address.String()),
		fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
		fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
		fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(10))).String()),
		fmt.Sprintf("--%s=%d", cli.FlagExpiration, time.Now().Add(time.Minute*time.Duration(120)).Unix()),
	})
	s.Require().NoError(err)
	s.Require().NoError(s.network.WaitForNextBlock())

	// Create new accounts in the keyring.
	s.grantee[3] = s.createAccount("grantee4")
	s.msgSendExec(s.grantee[3])

	s.grantee[4] = s.createAccount("grantee5")
	s.grantee[5] = s.createAccount("grantee6")

	// grant send authorization with allow list to grantee4
	out, err = authzclitestutil.CreateGrant(val.ClientCtx,
		[]string{
			s.grantee[3].String(),
			"send",
			fmt.Sprintf("--%s=100stake", cli.FlagSpendLimit),
			fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
			fmt.Sprintf("--%s=%s", flags.FlagFrom, val.Address.String()),
			fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
			fmt.Sprintf("--%s=%d", cli.FlagExpiration, time.Now().Add(time.Minute*time.Duration(120)).Unix()),
			fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			fmt.Sprintf("--%s=%s", cli.FlagAllowList, s.grantee[4]),
		},
	)
	s.Require().NoError(err)
	s.Require().NoError(s.network.WaitForNextBlock())

	s.Require().NoError(val.ClientCtx.Codec.UnmarshalJSON(out.Bytes(), &response), out.String())
	s.Require().NoError(clitestutil.CheckTxCode(s.network, val.ClientCtx, response.TxHash, 0))
}

func (s *E2ETestSuite) createAccount(uid string) sdk.AccAddress {
	val := s.network.Validators[0]
	// Create new account in the keyring.
	k, _, err := val.ClientCtx.Keyring.NewMnemonic(uid, keyring.English, sdk.FullFundraiserPath, keyring.DefaultBIP39Passphrase, hd.Secp256k1)
	s.Require().NoError(err)

	addr, err := k.GetAddress()
	s.Require().NoError(err)

	return addr
}

func (s *E2ETestSuite) msgSendExec(grantee sdk.AccAddress) {
	val := s.network.Validators[0]
	// Send some funds to the new account.
	out, err := clitestutil.MsgSendExec(
		val.ClientCtx,
		val.Address,
		grantee,
		sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(200))), fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
		fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
		fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
	)
	s.Require().NoError(err)
	s.Require().Contains(out.String(), `"code":0`)
	s.Require().NoError(s.network.WaitForNextBlock())
}

func (s *E2ETestSuite) TearDownSuite() {
	s.T().Log("tearing down e2e test suite")
	s.network.Cleanup()
}

var (
	typeMsgSend           = bank.SendAuthorization{}.MsgTypeURL()
	typeMsgVote           = sdk.MsgTypeURL(&govv1.MsgVote{})
	typeMsgSubmitProposal = sdk.MsgTypeURL(&govv1.MsgSubmitProposal{})
)

func execDelegate(val *network.Validator, args []string) (testutil.BufferWriter, error) {
	cmd := stakingcli.NewDelegateCmd()
	clientCtx := val.ClientCtx
	return clitestutil.ExecTestCLICmd(clientCtx, cmd, args)
}

func (s *E2ETestSuite) TestCmdRevokeAuthorizations() {
	val := s.network.Validators[0]

	grantee := s.grantee[0]
	twoHours := time.Now().Add(time.Minute * time.Duration(120)).Unix()

	// send-authorization
	_, err := authzclitestutil.CreateGrant(
		val.ClientCtx,
		[]string{
			grantee.String(),
			"send",
			fmt.Sprintf("--%s=100stake", cli.FlagSpendLimit),
			fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
			fmt.Sprintf("--%s=%s", flags.FlagFrom, val.Address),
			fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
			fmt.Sprintf("--%s=%d", cli.FlagExpiration, twoHours),
			fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
		},
	)
	s.Require().NoError(err)
	s.Require().NoError(s.network.WaitForNextBlock())

	// generic-authorization
	_, err = authzclitestutil.CreateGrant(
		val.ClientCtx,
		[]string{
			grantee.String(),
			"generic",
			fmt.Sprintf("--%s=%s", cli.FlagMsgType, typeMsgVote),
			fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
			fmt.Sprintf("--%s=%s", flags.FlagFrom, val.Address),
			fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
			fmt.Sprintf("--%s=%d", cli.FlagExpiration, twoHours),
			fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
		},
	)
	s.Require().NoError(err)
	s.Require().NoError(s.network.WaitForNextBlock())

	// generic-authorization used for amino testing
	_, err = authzclitestutil.CreateGrant(
		val.ClientCtx,
		[]string{
			grantee.String(),
			"generic",
			fmt.Sprintf("--%s=%s", cli.FlagMsgType, typeMsgSubmitProposal),
			fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
			fmt.Sprintf("--%s=%s", flags.FlagFrom, val.Address),
			fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
			fmt.Sprintf("--%s=%d", cli.FlagExpiration, twoHours),
			fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			fmt.Sprintf("--%s=%s", flags.FlagSignMode, flags.SignModeLegacyAminoJSON),
		},
	)
	s.Require().NoError(err)
	s.Require().NoError(s.network.WaitForNextBlock())

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
				"invalid grantee",
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
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
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
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			&sdk.TxResponse{}, 0,
			false,
		},
		{
			"Valid tx with amino",
			[]string{
				grantee.String(),
				typeMsgSubmitProposal,
				fmt.Sprintf("--%s=%s", flags.FlagFrom, val.Address.String()),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
				fmt.Sprintf("--%s=%s", flags.FlagSignMode, flags.SignModeLegacyAminoJSON),
			},
			&sdk.TxResponse{}, 0,
			false,
		},
	}
	for _, tc := range testCases {
		tc := tc
		s.Run(tc.name, func() {
			cmd := cli.NewCmdRevokeAuthorization(address.NewBech32Codec("cosmos"))
			clientCtx := val.ClientCtx

			out, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, tc.args)
			if tc.expectErr {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)
				s.Require().NoError(clientCtx.Codec.UnmarshalJSON(out.Bytes(), tc.respType), out.String())

				txResp := tc.respType.(*sdk.TxResponse)
				s.Require().NoError(clitestutil.CheckTxCode(s.network, val.ClientCtx, txResp.TxHash, tc.expectedCode))
			}
		})
	}
}

func (s *E2ETestSuite) TestExecAuthorizationWithExpiration() {
	val := s.network.Validators[0]
	grantee := s.grantee[0]
	tenSeconds := time.Now().Add(time.Second * time.Duration(10)).Unix()

	_, err := authzclitestutil.CreateGrant(
		val.ClientCtx,
		[]string{
			grantee.String(),
			"generic",
			fmt.Sprintf("--%s=%s", cli.FlagMsgType, typeMsgVote),
			fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
			fmt.Sprintf("--%s=%s", flags.FlagFrom, val.Address.String()),
			fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
			fmt.Sprintf("--%s=%d", cli.FlagExpiration, tenSeconds),
			fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
		},
	)
	s.Require().NoError(err)
	// msg vote
	voteTx := fmt.Sprintf(`{"body":{"messages":[{"@type":"/cosmos.gov.v1.MsgVote","proposal_id":"1","voter":"%s","option":"VOTE_OPTION_YES"}],"memo":"","timeout_height":"0","extension_options":[],"non_critical_extension_options":[]},"auth_info":{"signer_infos":[],"fee":{"amount":[],"gas_limit":"200000","payer":"","granter":""}},"signatures":[]}`, val.Address.String())
	execMsg := testutil.WriteToNewTempFile(s.T(), voteTx)
	defer execMsg.Close()

	// waiting for authorization to expires
	time.Sleep(12 * time.Second)

	cmd := cli.NewCmdExecAuthorization()
	clientCtx := val.ClientCtx

	out, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, []string{
		execMsg.Name(),
		fmt.Sprintf("--%s=%s", flags.FlagFrom, grantee.String()),
		fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
		fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
		fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
	})
	s.Require().NoError(err)
	var response sdk.TxResponse
	s.Require().NoError(clientCtx.Codec.UnmarshalJSON(out.Bytes(), &response), out.String())
	s.Require().NoError(clitestutil.CheckTxCode(s.network, clientCtx, response.TxHash, authz.ErrNoAuthorizationFound.ABCICode()))
}

func (s *E2ETestSuite) TestNewExecGenericAuthorized() {
	val := s.network.Validators[0]
	grantee := s.grantee[0]
	twoHours := time.Now().Add(time.Minute * time.Duration(120)).Unix()

	_, err := authzclitestutil.CreateGrant(
		val.ClientCtx,
		[]string{
			grantee.String(),
			"generic",
			fmt.Sprintf("--%s=%s", cli.FlagMsgType, typeMsgVote),
			fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
			fmt.Sprintf("--%s=%s", flags.FlagFrom, val.Address.String()),
			fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
			fmt.Sprintf("--%s=%d", cli.FlagExpiration, twoHours),
			fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
		},
	)
	s.Require().NoError(err)
	s.Require().NoError(s.network.WaitForNextBlock())

	// msg vote
	voteTx := fmt.Sprintf(`{"body":{"messages":[{"@type":"/cosmos.gov.v1.MsgVote","proposal_id":"1","voter":"%s","option":"VOTE_OPTION_YES"}],"memo":"","timeout_height":"0","extension_options":[],"non_critical_extension_options":[]},"auth_info":{"signer_infos":[],"fee":{"amount":[],"gas_limit":"200000","payer":"","granter":""}},"signatures":[]}`, val.Address.String())
	execMsg := testutil.WriteToNewTempFile(s.T(), voteTx)
	defer execMsg.Close()

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
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
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
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
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
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
			},
			&sdk.TxResponse{},
			0,
			false,
		},
		{
			"valid tx with amino",
			[]string{
				execMsg.Name(),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, grantee.String()),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagSignMode, flags.SignModeLegacyAminoJSON),
			},
			&sdk.TxResponse{}, 0,
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
				s.Require().NoError(clientCtx.Codec.UnmarshalJSON(out.Bytes(), tc.respType), out.String())
				txResp := tc.respType.(*sdk.TxResponse)
				s.Require().NoError(clitestutil.CheckTxCode(s.network, val.ClientCtx, txResp.TxHash, tc.expectedCode))
			}
		})
	}
}

func (s *E2ETestSuite) TestNewExecGrantAuthorized() {
	val := s.network.Validators[0]
	grantee := s.grantee[0]
	grantee1 := s.grantee[2]
	twoHours := time.Now().Add(time.Minute * time.Duration(120)).Unix()

	_, err := authzclitestutil.CreateGrant(
		val.ClientCtx,
		[]string{
			grantee.String(),
			"send",
			fmt.Sprintf("--%s=12%stoken", cli.FlagSpendLimit, val.Moniker),
			fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
			fmt.Sprintf("--%s=%s", flags.FlagFrom, val.Address.String()),
			fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
			fmt.Sprintf("--%s=%d", cli.FlagExpiration, twoHours),
			fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
		},
	)
	s.Require().NoError(err)
	s.Require().NoError(s.network.WaitForNextBlock())

	tokens := sdk.NewCoins(
		sdk.NewCoin(fmt.Sprintf("%stoken", val.Moniker), sdk.NewInt(12)),
	)
	normalGeneratedTx, err := clitestutil.MsgSendExec(
		val.ClientCtx,
		val.Address,
		grantee,
		tokens,
		fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
		fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
		fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
		fmt.Sprintf("--%s=true", flags.FlagGenerateOnly),
	)
	s.Require().NoError(err)
	execMsg := testutil.WriteToNewTempFile(s.T(), normalGeneratedTx.String())
	defer execMsg.Close()
	testCases := []struct {
		name         string
		args         []string
		expectedCode uint32
		expectErr    bool
		expectErrMsg string
	}{
		{
			"valid txn",
			[]string{
				execMsg.Name(),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, grantee.String()),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
			},
			0,
			false,
			"",
		},
		{
			"error over grantee doesn't exist on chain",
			[]string{
				execMsg.Name(),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, grantee1.String()),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
			},
			0,
			true,
			"insufficient funds", // earlier the error was account not found here.
		},
		{
			"error over spent",
			[]string{
				execMsg.Name(),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, grantee.String()),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
			},
			authz.ErrNoAuthorizationFound.ABCICode(),
			false,
			"",
		},
	}

	for _, tc := range testCases {
		tc := tc
		s.Run(tc.name, func() {
			cmd := cli.NewCmdExecAuthorization()
			clientCtx := val.ClientCtx

			var response sdk.TxResponse
			out, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, tc.args)
			switch {
			case tc.expectErrMsg != "":
				s.Require().NoError(clientCtx.Codec.UnmarshalJSON(out.Bytes(), &response), out.String())
				s.Require().Contains(response.RawLog, tc.expectErrMsg)

			case tc.expectErr:
				s.Require().Error(err)

			default:
				s.Require().NoError(err)
				s.Require().NoError(clientCtx.Codec.UnmarshalJSON(out.Bytes(), &response), out.String())
				s.Require().NoError(clitestutil.CheckTxCode(s.network, val.ClientCtx, response.TxHash, tc.expectedCode))
			}
		})
	}
}

func (s *E2ETestSuite) TestExecSendAuthzWithAllowList() {
	val := s.network.Validators[0]
	grantee := s.grantee[3]
	allowedAddr := s.grantee[4]
	notAllowedAddr := s.grantee[5]
	twoHours := time.Now().Add(time.Minute * time.Duration(120)).Unix()

	_, err := authzclitestutil.CreateGrant(
		val.ClientCtx,
		[]string{
			grantee.String(),
			"send",
			fmt.Sprintf("--%s=100stake", cli.FlagSpendLimit),
			fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
			fmt.Sprintf("--%s=%s", flags.FlagFrom, val.Address.String()),
			fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
			fmt.Sprintf("--%s=%d", cli.FlagExpiration, twoHours),
			fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			fmt.Sprintf("--%s=%s", cli.FlagAllowList, allowedAddr),
		},
	)
	s.Require().NoError(err)
	s.Require().NoError(s.network.WaitForNextBlock())

	tokens := sdk.NewCoins(
		sdk.NewCoin("stake", sdk.NewInt(12)),
	)

	validGeneratedTx, err := clitestutil.MsgSendExec(
		val.ClientCtx,
		val.Address,
		allowedAddr,
		tokens,
		fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
		fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
		fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
		fmt.Sprintf("--%s=true", flags.FlagGenerateOnly),
	)
	s.Require().NoError(err)
	execMsg := testutil.WriteToNewTempFile(s.T(), validGeneratedTx.String())
	defer execMsg.Close()

	invalidGeneratedTx, err := clitestutil.MsgSendExec(
		val.ClientCtx,
		val.Address,
		notAllowedAddr,
		tokens,
		fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
		fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
		fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
		fmt.Sprintf("--%s=true", flags.FlagGenerateOnly),
	)
	s.Require().NoError(err)
	execMsg1 := testutil.WriteToNewTempFile(s.T(), invalidGeneratedTx.String())
	defer execMsg1.Close()

	// test sending to allowed address
	args := []string{
		execMsg.Name(),
		fmt.Sprintf("--%s=%s", flags.FlagFrom, grantee.String()),
		fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
		fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
		fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
	}
	var response sdk.TxResponse
	cmd := cli.NewCmdExecAuthorization()
	out, err := clitestutil.ExecTestCLICmd(val.ClientCtx, cmd, args)
	s.Require().NoError(err)
	s.Require().NoError(val.ClientCtx.Codec.UnmarshalJSON(out.Bytes(), &response), out.String())
	s.Require().NoError(s.network.WaitForNextBlock())

	// test sending to not allowed address
	args = []string{
		execMsg1.Name(),
		fmt.Sprintf("--%s=%s", flags.FlagFrom, grantee.String()),
		fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
		fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
		fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
	}
	out, err = clitestutil.ExecTestCLICmd(val.ClientCtx, cmd, args)
	s.Require().NoError(err)
	s.Require().NoError(val.ClientCtx.Codec.UnmarshalJSON(out.Bytes(), &response), out.String())
	s.Require().NoError(s.network.WaitForNextBlock())

	// query tx and check result
	err = s.network.RetryForBlocks(func() error {
		out, err = clitestutil.ExecTestCLICmd(val.ClientCtx, authcli.QueryTxCmd(), []string{response.TxHash, fmt.Sprintf("--%s=json", flags.FlagOutput)})
		return err
	}, 3)
	s.Require().NoError(err)
	s.Contains(out.String(), fmt.Sprintf("cannot send to %s address", notAllowedAddr))
}

func (s *E2ETestSuite) TestExecDelegateAuthorization() {
	val := s.network.Validators[0]
	grantee := s.grantee[0]
	twoHours := time.Now().Add(time.Minute * time.Duration(120)).Unix()

	_, err := authzclitestutil.CreateGrant(
		val.ClientCtx,
		[]string{
			grantee.String(),
			"delegate",
			fmt.Sprintf("--%s=100stake", cli.FlagSpendLimit),
			fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
			fmt.Sprintf("--%s=%s", flags.FlagFrom, val.Address.String()),
			fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
			fmt.Sprintf("--%s=%d", cli.FlagExpiration, twoHours),
			fmt.Sprintf("--%s=%s", cli.FlagAllowedValidators, val.ValAddress.String()),
			fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
		},
	)
	s.Require().NoError(err)
	s.Require().NoError(s.network.WaitForNextBlock())

	tokens := sdk.NewCoins(
		sdk.NewCoin("stake", sdk.NewInt(50)),
	)

	delegateTx := fmt.Sprintf(`{"body":{"messages":[{"@type":"/cosmos.staking.v1beta1.MsgDelegate","delegator_address":"%s","validator_address":"%s","amount":{"denom":"%s","amount":"%s"}}],"memo":"","timeout_height":"0","extension_options":[],"non_critical_extension_options":[]},"auth_info":{"signer_infos":[],"fee":{"amount":[],"gas_limit":"200000","payer":"","granter":""}},"signatures":[]}`, val.Address.String(), val.ValAddress.String(),
		tokens.GetDenomByIndex(0), tokens[0].Amount)
	execMsg := testutil.WriteToNewTempFile(s.T(), delegateTx)
	defer execMsg.Close()

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
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
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
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
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
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
			},
			authz.ErrNoAuthorizationFound.ABCICode(),
			false,
			authz.ErrNoAuthorizationFound.Error(),
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
				s.Require().NoError(clientCtx.Codec.UnmarshalJSON(out.Bytes(), &response), out.String())
				s.Require().NoError(clitestutil.CheckTxCode(s.network, val.ClientCtx, response.TxHash, tc.expectedCode))
			}
		})
	}

	// test delegate no spend-limit
	_, err = authzclitestutil.CreateGrant(
		val.ClientCtx,
		[]string{
			grantee.String(),
			"delegate",
			fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
			fmt.Sprintf("--%s=%s", flags.FlagFrom, val.Address.String()),
			fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
			fmt.Sprintf("--%s=%d", cli.FlagExpiration, twoHours),
			fmt.Sprintf("--%s=%s", cli.FlagAllowedValidators, val.ValAddress.String()),
			fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
		},
	)
	s.Require().NoError(err)
	s.Require().NoError(s.network.WaitForNextBlock())

	tokens = sdk.NewCoins(
		sdk.NewCoin("stake", sdk.NewInt(50)),
	)

	delegateTx = fmt.Sprintf(`{"body":{"messages":[{"@type":"/cosmos.staking.v1beta1.MsgDelegate","delegator_address":"%s","validator_address":"%s","amount":{"denom":"%s","amount":"%s"}}],"memo":"","timeout_height":"0","extension_options":[],"non_critical_extension_options":[]},"auth_info":{"signer_infos":[],"fee":{"amount":[],"gas_limit":"200000","payer":"","granter":""}},"signatures":[]}`, val.Address.String(), val.ValAddress.String(),
		tokens.GetDenomByIndex(0), tokens[0].Amount)
	execMsg = testutil.WriteToNewTempFile(s.T(), delegateTx)
	defer execMsg.Close()

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
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
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
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
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
				s.Require().NoError(clientCtx.Codec.UnmarshalJSON(out.Bytes(), &response), out.String())
				s.Require().NoError(clitestutil.CheckTxCode(s.network, val.ClientCtx, response.TxHash, tc.expectedCode))
			}
		})
	}

	// test delegating to denied validator
	_, err = authzclitestutil.CreateGrant(
		val.ClientCtx,
		[]string{
			grantee.String(),
			"delegate",
			fmt.Sprintf("--%s=100stake", cli.FlagSpendLimit),
			fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
			fmt.Sprintf("--%s=%s", flags.FlagFrom, val.Address.String()),
			fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
			fmt.Sprintf("--%s=%d", cli.FlagExpiration, twoHours),
			fmt.Sprintf("--%s=%s", cli.FlagDenyValidators, val.ValAddress.String()),
			fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
		},
	)
	s.Require().NoError(err)
	s.Require().NoError(s.network.WaitForNextBlock())

	args := []string{
		execMsg.Name(),
		fmt.Sprintf("--%s=%s", flags.FlagFrom, grantee.String()),
		fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
		fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
		fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
	}
	cmd := cli.NewCmdExecAuthorization()
	out, err := clitestutil.ExecTestCLICmd(val.ClientCtx, cmd, args)
	s.Require().NoError(err)
	s.Require().NoError(s.network.WaitForNextBlock())

	var response sdk.TxResponse
	s.Require().NoError(val.ClientCtx.Codec.UnmarshalJSON(out.Bytes(), &response), out.String())

	// query tx and check result
	err = s.network.RetryForBlocks(func() error {
		out, err = clitestutil.ExecTestCLICmd(val.ClientCtx, authcli.QueryTxCmd(), []string{response.TxHash, fmt.Sprintf("--%s=json", flags.FlagOutput)})
		return err
	}, 3)
	s.Require().NoError(err)
	s.Contains(out.String(), fmt.Sprintf("cannot delegate/undelegate to %s validator", val.ValAddress.String()))
}

func (s *E2ETestSuite) TestExecUndelegateAuthorization() {
	val := s.network.Validators[0]
	grantee := s.grantee[0]
	twoHours := time.Now().Add(time.Minute * time.Duration(120)).Unix()

	// granting undelegate msg authorization
	_, err := authzclitestutil.CreateGrant(
		val.ClientCtx,
		[]string{
			grantee.String(),
			"unbond",
			fmt.Sprintf("--%s=100stake", cli.FlagSpendLimit),
			fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
			fmt.Sprintf("--%s=%s", flags.FlagFrom, val.Address.String()),
			fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
			fmt.Sprintf("--%s=%d", cli.FlagExpiration, twoHours),
			fmt.Sprintf("--%s=%s", cli.FlagAllowedValidators, val.ValAddress.String()),
			fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
		},
	)
	s.Require().NoError(err)
	s.Require().NoError(s.network.WaitForNextBlock())

	// delegating stakes to validator
	_, err = execDelegate(
		val,
		[]string{
			val.ValAddress.String(),
			"100stake",
			fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
			fmt.Sprintf("--%s=%s", flags.FlagFrom, val.Address.String()),
			fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
			fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
		},
	)
	s.Require().NoError(err)

	tokens := sdk.NewCoins(
		sdk.NewCoin("stake", sdk.NewInt(50)),
	)

	undelegateTx := fmt.Sprintf(`{"body":{"messages":[{"@type":"/cosmos.staking.v1beta1.MsgUndelegate","delegator_address":"%s","validator_address":"%s","amount":{"denom":"%s","amount":"%s"}}],"memo":"","timeout_height":"0","extension_options":[],"non_critical_extension_options":[]},"auth_info":{"signer_infos":[],"fee":{"amount":[],"gas_limit":"200000","payer":"","granter":""}},"signatures":[]}`, val.Address.String(), val.ValAddress.String(),
		tokens.GetDenomByIndex(0), tokens[0].Amount)
	execMsg := testutil.WriteToNewTempFile(s.T(), undelegateTx)
	defer execMsg.Close()

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
				fmt.Sprintf("--%s=%s", flags.FlagGas, "250000"),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, grantee.String()),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
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
				fmt.Sprintf("--%s=%s", flags.FlagGas, "250000"),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, grantee.String()),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
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
				fmt.Sprintf("--%s=%s", flags.FlagGas, "250000"),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, grantee.String()),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
			},
			authz.ErrNoAuthorizationFound.ABCICode(),
			false,
			authz.ErrNoAuthorizationFound.Error(),
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
				s.Require().NoError(clientCtx.Codec.UnmarshalJSON(out.Bytes(), &response), out.String())
				s.Require().NoError(clitestutil.CheckTxCode(s.network, val.ClientCtx, response.TxHash, tc.expectedCode))
			}
		})
	}

	// grant undelegate authorization without limit
	_, err = authzclitestutil.CreateGrant(
		val.ClientCtx,
		[]string{
			grantee.String(),
			"unbond",
			fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
			fmt.Sprintf("--%s=%s", flags.FlagFrom, val.Address.String()),
			fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
			fmt.Sprintf("--%s=%d", cli.FlagExpiration, twoHours),
			fmt.Sprintf("--%s=%s", cli.FlagAllowedValidators, val.ValAddress.String()),
			fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
		},
	)
	s.Require().NoError(err)
	s.Require().NoError(s.network.WaitForNextBlock())

	tokens = sdk.NewCoins(
		sdk.NewCoin("stake", sdk.NewInt(50)),
	)

	undelegateTx = fmt.Sprintf(`{"body":{"messages":[{"@type":"/cosmos.staking.v1beta1.MsgUndelegate","delegator_address":"%s","validator_address":"%s","amount":{"denom":"%s","amount":"%s"}}],"memo":"","timeout_height":"0","extension_options":[],"non_critical_extension_options":[]},"auth_info":{"signer_infos":[],"fee":{"amount":[],"gas_limit":"200000","payer":"","granter":""}},"signatures":[]}`, val.Address.String(), val.ValAddress.String(),
		tokens.GetDenomByIndex(0), tokens[0].Amount)
	execMsg = testutil.WriteToNewTempFile(s.T(), undelegateTx)
	defer execMsg.Close()

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
				fmt.Sprintf("--%s=%s", flags.FlagGas, "250000"),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, grantee.String()),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
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
				fmt.Sprintf("--%s=%s", flags.FlagGas, "250000"),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, grantee.String()),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
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
				s.Require().NoError(clientCtx.Codec.UnmarshalJSON(out.Bytes(), &response), out.String())
				s.Require().NoError(clitestutil.CheckTxCode(s.network, val.ClientCtx, response.TxHash, tc.expectedCode))
			}
		})
	}
}
