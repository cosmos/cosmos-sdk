package distribution

import (
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"cosmossdk.io/math"
	"cosmossdk.io/simapp"
	"github.com/cosmos/gogoproto/proto"
	"github.com/stretchr/testify/suite"

	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/codec/address"
	clitestutil "github.com/cosmos/cosmos-sdk/testutil/cli"
	"github.com/cosmos/cosmos-sdk/testutil/network"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/distribution/client/cli"
	distrclitestutil "github.com/cosmos/cosmos-sdk/x/distribution/client/testutil"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
)

type E2ETestSuite struct {
	suite.Suite

	cfg     network.Config
	network *network.Network
}

func NewE2ETestSuite(cfg network.Config) *E2ETestSuite {
	return &E2ETestSuite{cfg: cfg}
}

// SetupSuite creates a new network for _each_ e2e test. We create a new
// network for each test because there are some state modifications that are
// needed to be made in order to make useful queries. However, we don't want
// these state changes to be present in other tests.
func (s *E2ETestSuite) SetupSuite() {
	s.T().Log("setting up e2e test suite")

	cfg := network.DefaultConfig(simapp.NewTestNetworkFixture)
	cfg.NumValidators = 1
	s.cfg = cfg

	genesisState := s.cfg.GenesisState
	var mintData minttypes.GenesisState
	s.Require().NoError(s.cfg.Codec.UnmarshalJSON(genesisState[minttypes.ModuleName], &mintData))

	inflation := sdk.MustNewDecFromStr("1.0")
	mintData.Minter.Inflation = inflation
	mintData.Params.InflationMin = inflation
	mintData.Params.InflationMax = inflation

	mintDataBz, err := s.cfg.Codec.MarshalJSON(&mintData)
	s.Require().NoError(err)
	genesisState[minttypes.ModuleName] = mintDataBz
	s.cfg.GenesisState = genesisState

	s.network, err = network.New(s.T(), s.T().TempDir(), s.cfg)
	s.Require().NoError(err)

	s.Require().NoError(s.network.WaitForNextBlock())
}

// TearDownSuite cleans up the curret test network after _each_ test.
func (s *E2ETestSuite) TearDownSuite() {
	s.T().Log("tearing down e2e test suite1")
	s.network.Cleanup()
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
			`{"community_tax":"0.020000000000000000","base_proposer_reward":"0.000000000000000000","bonus_proposer_reward":"0.000000000000000000","withdraw_addr_enabled":true}`,
		},
		{
			"text output",
			[]string{fmt.Sprintf("--%s=text", flags.FlagOutput)},
			`base_proposer_reward: "0.000000000000000000"
bonus_proposer_reward: "0.000000000000000000"
community_tax: "0.020000000000000000"
withdraw_addr_enabled: true`,
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

func (s *E2ETestSuite) TestGetCmdQueryValidatorDistributionInfo() {
	val := s.network.Validators[0]

	testCases := []struct {
		name   string
		args   []string
		expErr bool
	}{
		{
			"invalid val address",
			[]string{"invalid address", fmt.Sprintf("--%s=json", flags.FlagOutput)},
			true,
		},
		{
			"json output",
			[]string{val.ValAddress.String(), fmt.Sprintf("--%s=json", flags.FlagOutput)},
			false,
		},
		{
			"text output",
			[]string{val.ValAddress.String(), fmt.Sprintf("--%s=text", flags.FlagOutput)},
			false,
		},
	}

	for _, tc := range testCases {
		tc := tc

		s.Run(tc.name, func() {
			cmd := cli.GetCmdQueryValidatorDistributionInfo()
			clientCtx := val.ClientCtx

			_, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, tc.args)
			if tc.expErr {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)
			}
		})
	}
}

func (s *E2ETestSuite) TestGetCmdQueryValidatorOutstandingRewards() {
	val := s.network.Validators[0]

	_, err := s.network.WaitForHeight(4)
	s.Require().NoError(err)

	testCases := []struct {
		name           string
		args           []string
		expectErr      bool
		expectedOutput string
	}{
		{
			"invalid validator address",
			[]string{
				fmt.Sprintf("--%s=3", flags.FlagHeight),
				"foo",
			},
			true,
			"",
		},
		{
			"json output",
			[]string{
				fmt.Sprintf("--%s=3", flags.FlagHeight),
				sdk.ValAddress(val.Address).String(),
				fmt.Sprintf("--%s=json", flags.FlagOutput),
			},
			false,
			`{"rewards":[{"denom":"stake","amount":"232.260000000000000000"}]}`,
		},
		{
			"text output",
			[]string{
				fmt.Sprintf("--%s=text", flags.FlagOutput),
				fmt.Sprintf("--%s=3", flags.FlagHeight),
				sdk.ValAddress(val.Address).String(),
			},
			false,
			`rewards:
- amount: "232.260000000000000000"
  denom: stake`,
		},
	}

	for _, tc := range testCases {
		tc := tc

		s.Run(tc.name, func() {
			cmd := cli.GetCmdQueryValidatorOutstandingRewards()
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

func (s *E2ETestSuite) TestGetCmdQueryValidatorCommission() {
	val := s.network.Validators[0]

	_, err := s.network.WaitForHeight(4)
	s.Require().NoError(err)

	testCases := []struct {
		name           string
		args           []string
		expectErr      bool
		expectedOutput string
	}{
		{
			"invalid validator address",
			[]string{
				fmt.Sprintf("--%s=3", flags.FlagHeight),
				"foo",
			},
			true,
			"",
		},
		{
			"json output",
			[]string{
				fmt.Sprintf("--%s=3", flags.FlagHeight),
				sdk.ValAddress(val.Address).String(),
				fmt.Sprintf("--%s=json", flags.FlagOutput),
			},
			false,
			`{"commission":[{"denom":"stake","amount":"116.130000000000000000"}]}`,
		},
		{
			"text output",
			[]string{
				fmt.Sprintf("--%s=text", flags.FlagOutput),
				fmt.Sprintf("--%s=3", flags.FlagHeight),
				sdk.ValAddress(val.Address).String(),
			},
			false,
			`commission:
- amount: "116.130000000000000000"
  denom: stake`,
		},
	}

	for _, tc := range testCases {
		tc := tc

		s.Run(tc.name, func() {
			cmd := cli.GetCmdQueryValidatorCommission()
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

func (s *E2ETestSuite) TestGetCmdQueryValidatorSlashes() {
	val := s.network.Validators[0]

	_, err := s.network.WaitForHeight(4)
	s.Require().NoError(err)

	testCases := []struct {
		name           string
		args           []string
		expectErr      bool
		expectedOutput string
	}{
		{
			"invalid validator address",
			[]string{
				fmt.Sprintf("--%s=3", flags.FlagHeight),
				"foo", "1", "3",
			},
			true,
			"",
		},
		{
			"invalid start height",
			[]string{
				fmt.Sprintf("--%s=3", flags.FlagHeight),
				sdk.ValAddress(val.Address).String(), "-1", "3",
			},
			true,
			"",
		},
		{
			"invalid end height",
			[]string{
				fmt.Sprintf("--%s=3", flags.FlagHeight),
				sdk.ValAddress(val.Address).String(), "1", "-3",
			},
			true,
			"",
		},
		{
			"json output",
			[]string{
				fmt.Sprintf("--%s=3", flags.FlagHeight),
				sdk.ValAddress(val.Address).String(), "1", "3",
				fmt.Sprintf("--%s=json", flags.FlagOutput),
			},
			false,
			"{\"slashes\":[],\"pagination\":{\"next_key\":null,\"total\":\"0\"}}",
		},
		{
			"text output",
			[]string{
				fmt.Sprintf("--%s=text", flags.FlagOutput),
				fmt.Sprintf("--%s=3", flags.FlagHeight),
				sdk.ValAddress(val.Address).String(), "1", "3",
			},
			false,
			"pagination:\n  next_key: null\n  total: \"0\"\nslashes: []",
		},
	}

	for _, tc := range testCases {
		tc := tc

		s.Run(tc.name, func() {
			cmd := cli.GetCmdQueryValidatorSlashes()
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

func (s *E2ETestSuite) TestGetCmdQueryDelegatorRewards() {
	val := s.network.Validators[0]
	addr := val.Address
	valAddr := sdk.ValAddress(addr)

	_, err := s.network.WaitForHeightWithTimeout(11, time.Minute)
	s.Require().NoError(err)

	testCases := []struct {
		name           string
		args           []string
		expectErr      bool
		expectedOutput string
	}{
		{
			"invalid delegator address",
			[]string{
				fmt.Sprintf("--%s=5", flags.FlagHeight),
				"foo", valAddr.String(),
			},
			true,
			"",
		},
		{
			"invalid validator address",
			[]string{
				fmt.Sprintf("--%s=5", flags.FlagHeight),
				addr.String(), "foo",
			},
			true,
			"",
		},
		{
			"json output",
			[]string{
				fmt.Sprintf("--%s=5", flags.FlagHeight),
				addr.String(),
				fmt.Sprintf("--%s=json", flags.FlagOutput),
			},
			false,
			fmt.Sprintf(`{"rewards":[{"validator_address":"%s","reward":[{"denom":"stake","amount":"193.550000000000000000"}]}],"total":[{"denom":"stake","amount":"193.550000000000000000"}]}`, valAddr.String()),
		},
		{
			"json output (specific validator)",
			[]string{
				fmt.Sprintf("--%s=5", flags.FlagHeight),
				addr.String(), valAddr.String(),
				fmt.Sprintf("--%s=json", flags.FlagOutput),
			},
			false,
			`{"rewards":[{"denom":"stake","amount":"193.550000000000000000"}]}`,
		},
		{
			"text output",
			[]string{
				fmt.Sprintf("--%s=text", flags.FlagOutput),
				fmt.Sprintf("--%s=5", flags.FlagHeight),
				addr.String(),
			},
			false,
			fmt.Sprintf(`rewards:
- reward:
  - amount: "193.550000000000000000"
    denom: stake
  validator_address: %s
total:
- amount: "193.550000000000000000"
  denom: stake`, valAddr.String()),
		},
		{
			"text output (specific validator)",
			[]string{
				fmt.Sprintf("--%s=text", flags.FlagOutput),
				fmt.Sprintf("--%s=5", flags.FlagHeight),
				addr.String(), valAddr.String(),
			},
			false,
			`rewards:
- amount: "193.550000000000000000"
  denom: stake`,
		},
	}

	for _, tc := range testCases {
		tc := tc

		s.Run(tc.name, func() {
			cmd := cli.GetCmdQueryDelegatorRewards(address.NewBech32Codec("cosmos"))
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

func (s *E2ETestSuite) TestGetCmdQueryCommunityPool() {
	val := s.network.Validators[0]

	_, err := s.network.WaitForHeight(4)
	s.Require().NoError(err)

	testCases := []struct {
		name           string
		args           []string
		expectedOutput string
	}{
		{
			"json output",
			[]string{fmt.Sprintf("--%s=3", flags.FlagHeight), fmt.Sprintf("--%s=json", flags.FlagOutput)},
			`{"pool":[{"denom":"stake","amount":"4.740000000000000000"}]}`,
		},
		{
			"text output",
			[]string{fmt.Sprintf("--%s=text", flags.FlagOutput), fmt.Sprintf("--%s=3", flags.FlagHeight)},
			`pool:
- amount: "4.740000000000000000"
  denom: stake`,
		},
	}

	for _, tc := range testCases {
		tc := tc

		s.Run(tc.name, func() {
			cmd := cli.GetCmdQueryCommunityPool()
			clientCtx := val.ClientCtx

			out, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, tc.args)
			s.Require().NoError(err)
			s.Require().Equal(tc.expectedOutput, strings.TrimSpace(out.String()))
		})
	}
}

func (s *E2ETestSuite) TestNewWithdrawRewardsCmd() {
	val := s.network.Validators[0]

	testCases := []struct {
		name                 string
		valAddr              fmt.Stringer
		args                 []string
		expectErr            bool
		expectedCode         uint32
		respType             proto.Message
		expectedResponseType []string
	}{
		{
			"invalid validator address",
			val.Address,
			[]string{
				fmt.Sprintf("--%s=%s", flags.FlagFrom, val.Address.String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			true, 0, nil,
			[]string{},
		},
		{
			"valid transaction",
			sdk.ValAddress(val.Address),
			[]string{
				fmt.Sprintf("--%s=%s", flags.FlagFrom, val.Address.String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			false, 0, &sdk.TxResponse{},
			[]string{
				"/cosmos.distribution.v1beta1.MsgWithdrawDelegatorRewardResponse",
			},
		},
		{
			"valid transaction (with commission)",
			sdk.ValAddress(val.Address),
			[]string{
				fmt.Sprintf("--%s=%s", flags.FlagFrom, val.Address.String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=true", cli.FlagCommission),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			false, 0, &sdk.TxResponse{},
			[]string{
				"/cosmos.distribution.v1beta1.MsgWithdrawDelegatorRewardResponse",
				"/cosmos.distribution.v1beta1.MsgWithdrawValidatorCommissionResponse",
			},
		},
	}

	for _, tc := range testCases {
		tc := tc

		s.Run(tc.name, func() {
			clientCtx := val.ClientCtx

			_, _ = s.network.WaitForHeightWithTimeout(10, time.Minute)
			bz, err := distrclitestutil.MsgWithdrawDelegatorRewardExec(clientCtx, tc.valAddr, tc.args...)
			if tc.expectErr {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)
				s.Require().NoError(clientCtx.Codec.UnmarshalJSON(bz, tc.respType), string(bz))
				s.Require().NoError(s.network.WaitForNextBlock())

				txResp, err := clitestutil.GetTxResponse(s.network, clientCtx, tc.respType.(*sdk.TxResponse).TxHash)
				s.Require().NoError(err)
				s.Require().Equal(tc.expectedCode, txResp.Code)

				data, err := hex.DecodeString(txResp.Data)
				s.Require().NoError(err)

				txMsgData := sdk.TxMsgData{}
				err = s.cfg.Codec.Unmarshal(data, &txMsgData)
				s.Require().NoError(err)
				for responseIdx, msgResponse := range txMsgData.MsgResponses {
					s.Require().Equal(tc.expectedResponseType[responseIdx], msgResponse.TypeUrl)
					switch msgResponse.TypeUrl {
					case "/cosmos.distribution.v1beta1.MsgWithdrawDelegatorRewardResponse":
						var resp distrtypes.MsgWithdrawDelegatorRewardResponse
						// can't use unpackAny as response types are not registered.
						err = s.cfg.Codec.Unmarshal(msgResponse.Value, &resp)
						s.Require().NoError(err)
						s.Require().True(resp.Amount.IsAllGT(sdk.NewCoins(sdk.NewCoin("stake", math.OneInt()))),
							fmt.Sprintf("expected a positive coin value, got %v", resp.Amount))
					case "/cosmos.distribution.v1beta1.MsgWithdrawValidatorCommissionResponse":
						var resp distrtypes.MsgWithdrawValidatorCommissionResponse
						// can't use unpackAny as response types are not registered.
						err = s.cfg.Codec.Unmarshal(msgResponse.Value, &resp)
						s.Require().NoError(err)
						s.Require().True(resp.Amount.IsAllGT(sdk.NewCoins(sdk.NewCoin("stake", math.OneInt()))),
							fmt.Sprintf("expected a positive coin value, got %v", resp.Amount))
					}
				}
			}
		})
	}
}

func (s *E2ETestSuite) TestNewWithdrawAllRewardsCmd() {
	val := s.network.Validators[0]

	testCases := []struct {
		name                 string
		args                 []string
		expectErr            bool
		expectedCode         uint32
		respType             proto.Message
		expectedResponseType []string
	}{
		{
			"valid transaction (offline)",
			[]string{
				fmt.Sprintf("--%s=%s", flags.FlagFrom, val.Address.String()),
				fmt.Sprintf("--%s=true", flags.FlagOffline),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			true, 0, nil,
			[]string{},
		},
		{
			"valid transaction",
			[]string{
				fmt.Sprintf("--%s=%s", flags.FlagFrom, val.Address.String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			false, 0, &sdk.TxResponse{},
			[]string{
				"/cosmos.distribution.v1beta1.MsgWithdrawDelegatorRewardResponse",
			},
		},
	}

	for _, tc := range testCases {
		tc := tc

		s.Run(tc.name, func() {
			cmd := cli.NewWithdrawAllRewardsCmd()
			clientCtx := val.ClientCtx

			_, _ = s.network.WaitForHeightWithTimeout(10, time.Minute)

			out, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, tc.args)
			if tc.expectErr {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)
				s.Require().NoError(clientCtx.Codec.UnmarshalJSON(out.Bytes(), tc.respType), out.String())
				s.Require().NoError(s.network.WaitForNextBlock())

				txResp, err := clitestutil.GetTxResponse(s.network, clientCtx, tc.respType.(*sdk.TxResponse).TxHash)
				s.Require().NoError(err)
				s.Require().Equal(tc.expectedCode, txResp.Code)

				data, err := hex.DecodeString(txResp.Data)
				s.Require().NoError(err)

				txMsgData := sdk.TxMsgData{}
				err = s.cfg.Codec.Unmarshal(data, &txMsgData)
				s.Require().NoError(err)
				for responseIdx, msgResponse := range txMsgData.MsgResponses {
					s.Require().Equal(tc.expectedResponseType[responseIdx], msgResponse.TypeUrl)
					switch msgResponse.TypeUrl {
					case "/cosmos.distribution.v1beta1.MsgWithdrawDelegatorRewardResponse":
						var resp distrtypes.MsgWithdrawDelegatorRewardResponse
						// can't use unpackAny as response types are not registered.
						err = s.cfg.Codec.Unmarshal(msgResponse.Value, &resp)
						s.Require().NoError(err)
						s.Require().True(resp.Amount.IsAllGT(sdk.NewCoins(sdk.NewCoin("stake", math.OneInt()))),
							fmt.Sprintf("expected a positive coin value, got %v", resp.Amount))
					case "/cosmos.distribution.v1beta1.MsgWithdrawValidatorCommissionResponse":
						var resp distrtypes.MsgWithdrawValidatorCommissionResponse
						// can't use unpackAny as response types are not registered.
						err = s.cfg.Codec.Unmarshal(msgResponse.Value, &resp)
						s.Require().NoError(err)
						s.Require().True(resp.Amount.IsAllGT(sdk.NewCoins(sdk.NewCoin("stake", math.OneInt()))),
							fmt.Sprintf("expected a positive coin value, got %v", resp.Amount))
					}
				}
			}
		})
	}
}

func (s *E2ETestSuite) TestNewSetWithdrawAddrCmd() {
	val := s.network.Validators[0]

	testCases := []struct {
		name         string
		args         []string
		expectErr    bool
		expectedCode uint32
		respType     proto.Message
	}{
		{
			"invalid withdraw address",
			[]string{
				"foo",
				fmt.Sprintf("--%s=%s", flags.FlagFrom, val.Address.String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			true, 0, nil,
		},
		{
			"valid transaction",
			[]string{
				val.Address.String(),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, val.Address.String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			false, 0, &sdk.TxResponse{},
		},
	}

	for _, tc := range testCases {
		tc := tc

		s.Run(tc.name, func() {
			cmd := cli.NewSetWithdrawAddrCmd(address.NewBech32Codec("cosmos"))
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

func (s *E2ETestSuite) TestNewFundCommunityPoolCmd() {
	val := s.network.Validators[0]

	testCases := []struct {
		name         string
		args         []string
		expectErr    bool
		expectedCode uint32
		respType     proto.Message
	}{
		{
			"invalid funding amount",
			[]string{
				"-43foocoin",
				fmt.Sprintf("--%s=%s", flags.FlagFrom, val.Address.String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			true, 0, nil,
		},
		{
			"valid transaction",
			[]string{
				sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(5431))).String(),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, val.Address.String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			false, 0, &sdk.TxResponse{},
		},
	}

	for _, tc := range testCases {
		tc := tc

		s.Run(tc.name, func() {
			cmd := cli.NewFundCommunityPoolCmd()
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
