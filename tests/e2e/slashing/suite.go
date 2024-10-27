package testutil

import (
	"fmt"
	"strings"

	"github.com/cosmos/gogoproto/proto"
	"github.com/stretchr/testify/suite"

	"github.com/cosmos/cosmos-sdk/client/flags"
	clitestutil "github.com/cosmos/cosmos-sdk/testutil/cli"
	"github.com/cosmos/cosmos-sdk/testutil/network"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/slashing/client/cli"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
)

type E2ETestSuite struct {
	suite.Suite

	cfg     network.Config
	network *network.Network
}

func NewE2ETestSuite(cfg network.Config) *E2ETestSuite {
	return &E2ETestSuite{cfg: cfg}
}

// SetupSuite executes bootstrapping logic before all the tests, i.e. once before
// the entire suite, start executing.
func (s *E2ETestSuite) SetupSuite() {
	s.T().Log("setting up e2e test suite")

	var err error
	s.network, err = network.New(s.T(), s.T().TempDir(), s.cfg)
	s.Require().NoError(err)

	s.Require().NoError(s.network.WaitForNextBlock())
}

// TearDownSuite performs cleanup logic after all the tests, i.e. once after the
// entire suite, has finished executing.
func (s *E2ETestSuite) TearDownSuite() {
	s.T().Log("tearing down e2e test suite")
	s.network.Cleanup()
}

func (s *E2ETestSuite) TestGetCmdQuerySigningInfo() {
	val := s.network.Validators[0]
	pubKeyBz, err := s.cfg.Codec.MarshalInterfaceJSON(val.PubKey)
	s.Require().NoError(err)
	pubKeyStr := string(pubKeyBz)

	testCases := []struct {
		name           string
		args           []string
		expectErr      bool
		expectedOutput string
	}{
		{"invalid address", []string{"foo"}, true, ``},
		{
			"valid address (json output)",
			[]string{
				pubKeyStr,
				fmt.Sprintf("--%s=json", flags.FlagOutput),
				fmt.Sprintf("--%s=1", flags.FlagHeight),
			},
			false,
			fmt.Sprintf("{\"address\":\"%s\",\"start_height\":\"0\",\"index_offset\":\"0\",\"jailed_until\":\"1970-01-01T00:00:00Z\",\"tombstoned\":false,\"missed_blocks_counter\":\"0\"}", sdk.ConsAddress(val.PubKey.Address())),
		},
		{
			"valid address (text output)",
			[]string{
				pubKeyStr,
				fmt.Sprintf("--%s=text", flags.FlagOutput),
				fmt.Sprintf("--%s=1", flags.FlagHeight),
			},
			false,
			fmt.Sprintf(`address: %s
index_offset: "0"
jailed_until: "1970-01-01T00:00:00Z"
missed_blocks_counter: "0"
start_height: "0"
tombstoned: false`, sdk.ConsAddress(val.PubKey.Address())),
		},
	}

	for _, tc := range testCases {
		tc := tc

		s.Run(tc.name, func() {
			cmd := cli.GetCmdQuerySigningInfo()
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

func (s *E2ETestSuite) TestGetCmdQueryParams() {
	val := s.network.Validators[0]

	testCases := []struct {
		name           string
		args           []string
		expectedOutput string
	}{
		{
			"json output",
			[]string{fmt.Sprintf("--%s=json", flags.FlagOutput)},
			`{"signed_blocks_window":"100","min_signed_per_window":"0.500000000000000000","downtime_jail_duration":"600s","slash_fraction_double_sign":"0.050000000000000000","slash_fraction_downtime":"0.010000000000000000"}`,
		},
		{
			"text output",
			[]string{fmt.Sprintf("--%s=text", flags.FlagOutput)},
			`downtime_jail_duration: 600s
min_signed_per_window: "0.500000000000000000"
signed_blocks_window: "100"
slash_fraction_double_sign: "0.050000000000000000"
slash_fraction_downtime: "0.010000000000000000"`,
		},
	}

	for _, tc := range testCases {
		tc := tc

		s.Run(tc.name, func() {
			cmd := cli.GetCmdQueryParams()
			clientCtx := val.ClientCtx

			out, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, tc.args)
			s.Require().NoError(err)
			s.Require().Equal(tc.expectedOutput, strings.TrimSpace(out.String()))
		})
	}
}

func (s *E2ETestSuite) TestNewUnjailTxCmd() {
	val := s.network.Validators[0]
	testCases := []struct {
		name         string
		args         []string
		expectErr    bool
		expectedCode uint32
		respType     proto.Message
	}{
		{
			"valid transaction",
			[]string{
				fmt.Sprintf("--%s=%s", flags.FlagFrom, val.Address.String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync), // sync mode as there are no funds yet
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			false, slashingtypes.ErrValidatorNotJailed.ABCICode(), &sdk.TxResponse{},
		},
	}

	for _, tc := range testCases {
		tc := tc

		s.Run(tc.name, func() {
			cmd := cli.NewUnjailTxCmd()
			clientCtx := val.ClientCtx

			out, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, tc.args)
			if tc.expectErr {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)
				s.Require().NoError(clientCtx.Codec.UnmarshalJSON(out.Bytes(), tc.respType), out.String())

				txResp := tc.respType.(*sdk.TxResponse)
				s.Require().NoError(clitestutil.CheckTxCode(s.network, clientCtx, txResp.TxHash, tc.expectedCode))
			}
		})
	}
}
