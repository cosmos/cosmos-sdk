package testutil

import (
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"github.com/gogo/protobuf/proto"
	"github.com/stretchr/testify/suite"
	tmcli "github.com/tendermint/tendermint/libs/cli"

	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/testutil"
	clitestutil "github.com/cosmos/cosmos-sdk/testutil/cli"
	"github.com/cosmos/cosmos-sdk/testutil/network"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/distribution/client/cli"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
)

type IntegrationTestSuite struct {
	suite.Suite

	cfg     network.Config
	network *network.Network
}

func NewIntegrationTestSuite(cfg network.Config) *IntegrationTestSuite {
	return &IntegrationTestSuite{cfg: cfg}
}

// SetupSuite creates a new network for _each_ integration test. We create a new
// network for each test because there are some state modifications that are
// needed to be made in order to make useful queries. However, we don't want
// these state changes to be present in other tests.
func (s *IntegrationTestSuite) SetupSuite() {
	s.T().Log("setting up integration test suite")

	cfg := network.DefaultConfig()
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

	_, err = s.network.WaitForHeight(1)
	s.Require().NoError(err)
}

// TearDownSuite cleans up the curret test network after _each_ test.
func (s *IntegrationTestSuite) TearDownSuite() {
	s.T().Log("tearing down integration test suite1")
	s.network.Cleanup()
}

func (s *IntegrationTestSuite) TestGetCmdQueryParams() {
	val := s.network.Validators[0]

	testCases := []struct {
		name           string
		args           []string
		expectedOutput string
	}{
		{
			"json output",
			[]string{fmt.Sprintf("--%s=json", tmcli.OutputFlag)},
			`{"community_tax":"0.020000000000000000","base_proposer_reward":"0.010000000000000000","bonus_proposer_reward":"0.040000000000000000","withdraw_addr_enabled":true}`,
		},
		{
			"text output",
			[]string{fmt.Sprintf("--%s=text", tmcli.OutputFlag)},
			`base_proposer_reward: "0.010000000000000000"
bonus_proposer_reward: "0.040000000000000000"
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

func (s *IntegrationTestSuite) TestGetCmdQueryValidatorDistributionInfo() {
	val := s.network.Validators[0]

	testCases := []struct {
		name   string
		args   []string
		expErr bool
	}{
		{
			"invalid val address",
			[]string{"invalid address", fmt.Sprintf("--%s=json", tmcli.OutputFlag)},
			true,
		},
		{
			"json output",
			[]string{val.ValAddress.String(), fmt.Sprintf("--%s=json", tmcli.OutputFlag)},
			false,
		},
		{
			"text output",
			[]string{val.ValAddress.String(), fmt.Sprintf("--%s=text", tmcli.OutputFlag)},
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

func (s *IntegrationTestSuite) TestGetCmdQueryValidatorOutstandingRewards() {
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
				fmt.Sprintf("--%s=json", tmcli.OutputFlag),
			},
			false,
			`{"rewards":[{"denom":"stake","amount":"232.260000000000000000"}]}`,
		},
		{
			"text output",
			[]string{
				fmt.Sprintf("--%s=text", tmcli.OutputFlag),
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

func (s *IntegrationTestSuite) TestGetCmdQueryValidatorCommission() {
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
				fmt.Sprintf("--%s=json", tmcli.OutputFlag),
			},
			false,
			`{"commission":[{"denom":"stake","amount":"116.130000000000000000"}]}`,
		},
		{
			"text output",
			[]string{
				fmt.Sprintf("--%s=text", tmcli.OutputFlag),
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

func (s *IntegrationTestSuite) TestGetCmdQueryValidatorSlashes() {
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
				fmt.Sprintf("--%s=json", tmcli.OutputFlag),
			},
			false,
			"{\"slashes\":[],\"pagination\":{\"next_key\":null,\"total\":\"0\"}}",
		},
		{
			"text output",
			[]string{
				fmt.Sprintf("--%s=text", tmcli.OutputFlag),
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

func (s *IntegrationTestSuite) TestGetCmdQueryDelegatorRewards() {
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
				fmt.Sprintf("--%s=json", tmcli.OutputFlag),
			},
			false,
			fmt.Sprintf(`{"rewards":[{"validator_address":"%s","reward":[{"denom":"stake","amount":"193.550000000000000000"}]}],"total":[{"denom":"stake","amount":"193.550000000000000000"}]}`, valAddr.String()),
		},
		{
			"json output (specific validator)",
			[]string{
				fmt.Sprintf("--%s=5", flags.FlagHeight),
				addr.String(), valAddr.String(),
				fmt.Sprintf("--%s=json", tmcli.OutputFlag),
			},
			false,
			`{"rewards":[{"denom":"stake","amount":"193.550000000000000000"}]}`,
		},
		{
			"text output",
			[]string{
				fmt.Sprintf("--%s=text", tmcli.OutputFlag),
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
				fmt.Sprintf("--%s=text", tmcli.OutputFlag),
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
			cmd := cli.GetCmdQueryDelegatorRewards()
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

func (s *IntegrationTestSuite) TestGetCmdQueryCommunityPool() {
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
			[]string{fmt.Sprintf("--%s=3", flags.FlagHeight), fmt.Sprintf("--%s=json", tmcli.OutputFlag)},
			`{"pool":[{"denom":"stake","amount":"4.740000000000000000"}]}`,
		},
		{
			"text output",
			[]string{fmt.Sprintf("--%s=text", tmcli.OutputFlag), fmt.Sprintf("--%s=3", flags.FlagHeight)},
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

func (s *IntegrationTestSuite) TestNewWithdrawRewardsCmd() {
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
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
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
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
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
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
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
			bz, err := MsgWithdrawDelegatorRewardExec(clientCtx, tc.valAddr, tc.args...)
			if tc.expectErr {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)
				s.Require().NoError(clientCtx.Codec.UnmarshalJSON(bz, tc.respType), string(bz))

				txResp := tc.respType.(*sdk.TxResponse)
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
						s.Require().True(resp.Amount.IsAllGT(sdk.NewCoins(sdk.NewCoin("stake", sdk.OneInt()))),
							fmt.Sprintf("expected a positive coin value, got %v", resp.Amount))
					case "/cosmos.distribution.v1beta1.MsgWithdrawValidatorCommissionResponse":
						var resp distrtypes.MsgWithdrawValidatorCommissionResponse
						// can't use unpackAny as response types are not registered.
						err = s.cfg.Codec.Unmarshal(msgResponse.Value, &resp)
						s.Require().NoError(err)
						s.Require().True(resp.Amount.IsAllGT(sdk.NewCoins(sdk.NewCoin("stake", sdk.OneInt()))),
							fmt.Sprintf("expected a positive coin value, got %v", resp.Amount))
					}
				}
			}
		})
	}
}

func (s *IntegrationTestSuite) TestNewWithdrawAllRewardsCmd() {
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
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
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
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
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

				txResp := tc.respType.(*sdk.TxResponse)
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
						s.Require().True(resp.Amount.IsAllGT(sdk.NewCoins(sdk.NewCoin("stake", sdk.OneInt()))),
							fmt.Sprintf("expected a positive coin value, got %v", resp.Amount))
					case "/cosmos.distribution.v1beta1.MsgWithdrawValidatorCommissionResponse":
						var resp distrtypes.MsgWithdrawValidatorCommissionResponse
						// can't use unpackAny as response types are not registered.
						err = s.cfg.Codec.Unmarshal(msgResponse.Value, &resp)
						s.Require().NoError(err)
						s.Require().True(resp.Amount.IsAllGT(sdk.NewCoins(sdk.NewCoin("stake", sdk.OneInt()))),
							fmt.Sprintf("expected a positive coin value, got %v", resp.Amount))
					}
				}
			}
		})
	}
}

func (s *IntegrationTestSuite) TestNewSetWithdrawAddrCmd() {
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
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
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
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			false, 0, &sdk.TxResponse{},
		},
	}

	for _, tc := range testCases {
		tc := tc

		s.Run(tc.name, func() {
			cmd := cli.NewSetWithdrawAddrCmd()
			clientCtx := val.ClientCtx

			out, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, tc.args)
			if tc.expectErr {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)
				s.Require().NoError(clientCtx.Codec.UnmarshalJSON(out.Bytes(), tc.respType), out.String())

				txResp := tc.respType.(*sdk.TxResponse)
				s.Require().Equal(tc.expectedCode, txResp.Code)
			}
		})
	}
}

func (s *IntegrationTestSuite) TestNewFundCommunityPoolCmd() {
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
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
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
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
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
				s.Require().Equal(tc.expectedCode, txResp.Code)
			}
		})
	}
}

func (s *IntegrationTestSuite) TestGetCmdSubmitProposal() {
	val := s.network.Validators[0]
	invalidProp := `{
  "title": "",
  "description": "Pay me some Atoms!",
  "recipient": "foo",
  "amount": "-343foocoin",
  "deposit": -324foocoin
}`

	// fund some tokens to the community pool
	args := []string{
		sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(5431))).String(),
		fmt.Sprintf("--%s=%s", flags.FlagFrom, val.Address.String()),
		fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
		fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
		fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
	}

	invalidPropFile := testutil.WriteToNewTempFile(s.T(), invalidProp)
	cmd := cli.NewFundCommunityPoolCmd()
	out, err := clitestutil.ExecTestCLICmd(val.ClientCtx, cmd, args)
	s.Require().NoError(err)

	var txResp sdk.TxResponse
	s.Require().NoError(val.ClientCtx.Codec.UnmarshalJSON(out.Bytes(), &txResp), out.String())
	s.Require().Equal(uint32(0), txResp.Code)

	validProp := fmt.Sprintf(`{
  "title": "Community Pool Spend",
  "description": "Pay me some Atoms!",
  "recipient": "%s",
  "amount": "%s",
  "deposit": "%s"
}`, val.Address.String(), sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(5431)), sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(5431)))

	validPropFile := testutil.WriteToNewTempFile(s.T(), validProp)
	testCases := []struct {
		name         string
		args         []string
		expectErr    bool
		expectedCode uint32
		respType     proto.Message
	}{
		{
			"invalid proposal",
			[]string{
				invalidPropFile.Name(),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, val.Address.String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			true, 0, nil,
		},
		{
			"valid transaction",
			[]string{
				validPropFile.Name(),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, val.Address.String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			false, 0, &sdk.TxResponse{},
		},
	}

	for _, tc := range testCases {
		tc := tc

		s.Run(tc.name, func() {
			cmd := cli.GetCmdSubmitProposal()
			clientCtx := val.ClientCtx
			flags.AddTxFlagsToCmd(cmd)

			out, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, tc.args)
			if tc.expectErr {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)
				s.Require().NoError(clientCtx.Codec.UnmarshalJSON(out.Bytes(), tc.respType), out.String())

				txResp := tc.respType.(*sdk.TxResponse)
				s.Require().Equal(tc.expectedCode, txResp.Code, out.String())
			}
		})
	}
}
