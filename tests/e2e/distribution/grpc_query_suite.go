package distribution

import (
	"fmt"
	"strings"
	"time"

	"cosmossdk.io/simapp"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/codec/address"
	clitestutil "github.com/cosmos/cosmos-sdk/testutil/cli"
	"github.com/cosmos/cosmos-sdk/x/distribution/client/cli"

	"github.com/cosmos/gogoproto/proto"
	"github.com/stretchr/testify/suite"

	sdktestutil "github.com/cosmos/cosmos-sdk/testutil"
	"github.com/cosmos/cosmos-sdk/testutil/network"
	sdk "github.com/cosmos/cosmos-sdk/types"
	grpctypes "github.com/cosmos/cosmos-sdk/types/grpc"
	"github.com/cosmos/cosmos-sdk/types/query"
	"github.com/cosmos/cosmos-sdk/x/distribution/types"
)

type GRPCQueryTestSuite struct {
	suite.Suite

	cfg     network.Config
	network *network.Network
}

func (s *GRPCQueryTestSuite) SetupSuite() {
	s.T().Log("setting up e2e test suite")

	cfg := network.DefaultConfig(simapp.NewTestNetworkFixture)
	cfg.NumValidators = 1
	s.cfg = cfg

	var err error
	s.network, err = network.New(s.T(), s.T().TempDir(), cfg)
	s.Require().NoError(err)

	s.Require().NoError(s.network.WaitForNextBlock())
}

// TearDownSuite cleans up the curret test network after _each_ test.
func (s *GRPCQueryTestSuite) TearDownSuite() {
	s.T().Log("tearing down e2e test suite1")
	s.network.Cleanup()
}

func (s *GRPCQueryTestSuite) TestQueryParamsGRPC() {
	val := s.network.Validators[0]
	baseURL := val.APIAddress

	testCases := []struct {
		name     string
		url      string
		respType proto.Message
		expected proto.Message
	}{
		{
			"gRPC request params",
			fmt.Sprintf("%s/cosmos/distribution/v1beta1/params", baseURL),
			&types.QueryParamsResponse{},
			&types.QueryParamsResponse{
				Params: types.DefaultParams(),
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		resp, err := sdktestutil.GetRequest(tc.url)
		s.Run(tc.name, func() {
			s.Require().NoError(err)
			s.Require().NoError(val.ClientCtx.Codec.UnmarshalJSON(resp, tc.respType))
			s.Require().Equal(tc.expected, tc.respType)
		})
	}
}

func (s *GRPCQueryTestSuite) TestQueryValidatorDistributionInfoGRPC() {
	val := s.network.Validators[0]
	baseURL := val.APIAddress

	testCases := []struct {
		name     string
		url      string
		expErr   bool
		respType proto.Message
	}{
		{
			"gRPC request with wrong validator address",
			fmt.Sprintf("%s/cosmos/distribution/v1beta1/validators/%s", baseURL, "wrongAddress"),
			true,
			&types.QueryValidatorDistributionInfoResponse{},
		},
		{
			"gRPC request with valid validator address ",
			fmt.Sprintf("%s/cosmos/distribution/v1beta1/validators/%s", baseURL, val.ValAddress.String()),
			false,
			&types.QueryValidatorDistributionInfoResponse{},
		},
	}

	for _, tc := range testCases {
		tc := tc
		resp, err := sdktestutil.GetRequest(tc.url)
		s.Run(tc.name, func() {
			if tc.expErr {
				s.Require().Error(val.ClientCtx.Codec.UnmarshalJSON(resp, tc.respType))
			} else {
				s.Require().NoError(err)
				s.Require().NoError(val.ClientCtx.Codec.UnmarshalJSON(resp, tc.respType))
			}
		})
	}
}

func (s *GRPCQueryTestSuite) TestQueryOutstandingRewardsGRPC() {
	val := s.network.Validators[0]
	baseURL := val.APIAddress

	rewards, err := sdk.ParseDecCoins("19.6stake")
	s.Require().NoError(err)

	testCases := []struct {
		name     string
		url      string
		headers  map[string]string
		expErr   bool
		respType proto.Message
		expected proto.Message
	}{
		{
			"gRPC request params with wrong validator address",
			fmt.Sprintf("%s/cosmos/distribution/v1beta1/validators/%s/outstanding_rewards", baseURL, "wrongAddress"),
			map[string]string{},
			true,
			&types.QueryValidatorOutstandingRewardsResponse{},
			&types.QueryValidatorOutstandingRewardsResponse{},
		},
		{
			"gRPC request params valid address",
			fmt.Sprintf("%s/cosmos/distribution/v1beta1/validators/%s/outstanding_rewards", baseURL, val.ValAddress.String()),
			map[string]string{
				grpctypes.GRPCBlockHeightHeader: "2",
			},
			false,
			&types.QueryValidatorOutstandingRewardsResponse{},
			&types.QueryValidatorOutstandingRewardsResponse{
				Rewards: types.ValidatorOutstandingRewards{
					Rewards: rewards,
				},
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		resp, err := sdktestutil.GetRequestWithHeaders(tc.url, tc.headers)
		s.Run(tc.name, func() {
			if tc.expErr {
				s.Require().Error(val.ClientCtx.Codec.UnmarshalJSON(resp, tc.respType))
			} else {
				s.Require().NoError(err)
				s.Require().NoError(val.ClientCtx.Codec.UnmarshalJSON(resp, tc.respType))
				s.Require().Equal(tc.expected.String(), tc.respType.String())
			}
		})
	}
}

func (s *GRPCQueryTestSuite) TestQueryValidatorCommissionGRPC() {
	val := s.network.Validators[0]
	baseURL := val.APIAddress

	commission, err := sdk.ParseDecCoins("9.8stake")
	s.Require().NoError(err)

	testCases := []struct {
		name     string
		url      string
		headers  map[string]string
		expErr   bool
		respType proto.Message
		expected proto.Message
	}{
		{
			"gRPC request params with wrong validator address",
			fmt.Sprintf("%s/cosmos/distribution/v1beta1/validators/%s/commission", baseURL, "wrongAddress"),
			map[string]string{},
			true,
			&types.QueryValidatorCommissionResponse{},
			&types.QueryValidatorCommissionResponse{},
		},
		{
			"gRPC request params valid address",
			fmt.Sprintf("%s/cosmos/distribution/v1beta1/validators/%s/commission", baseURL, val.ValAddress.String()),
			map[string]string{
				grpctypes.GRPCBlockHeightHeader: "2",
			},
			false,
			&types.QueryValidatorCommissionResponse{},
			&types.QueryValidatorCommissionResponse{
				Commission: types.ValidatorAccumulatedCommission{
					Commission: commission,
				},
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		resp, err := sdktestutil.GetRequestWithHeaders(tc.url, tc.headers)
		s.Run(tc.name, func() {
			if tc.expErr {
				s.Require().Error(val.ClientCtx.Codec.UnmarshalJSON(resp, tc.respType))
			} else {
				s.Require().NoError(err)
				s.Require().NoError(val.ClientCtx.Codec.UnmarshalJSON(resp, tc.respType))
				s.Require().Equal(tc.expected.String(), tc.respType.String())
			}
		})
	}
}

func (s *GRPCQueryTestSuite) TestQuerySlashesGRPC() {
	val := s.network.Validators[0]
	baseURL := val.APIAddress

	testCases := []struct {
		name     string
		url      string
		expErr   bool
		respType proto.Message
		expected proto.Message
	}{
		{
			"invalid validator address",
			fmt.Sprintf("%s/cosmos/distribution/v1beta1/validators/%s/slashes", baseURL, ""),
			true,
			&types.QueryValidatorSlashesResponse{},
			nil,
		},
		{
			"invalid start height",
			fmt.Sprintf("%s/cosmos/distribution/v1beta1/validators/%s/slashes?starting_height=%s&ending_height=%s", baseURL, val.ValAddress.String(), "-1", "3"),
			true,
			&types.QueryValidatorSlashesResponse{},
			nil,
		},
		{
			"invalid start height",
			fmt.Sprintf("%s/cosmos/distribution/v1beta1/validators/%s/slashes?starting_height=%s&ending_height=%s", baseURL, val.ValAddress.String(), "1", "-3"),
			true,
			&types.QueryValidatorSlashesResponse{},
			nil,
		},
		{
			"valid request get slashes",
			fmt.Sprintf("%s/cosmos/distribution/v1beta1/validators/%s/slashes?starting_height=%s&ending_height=%s", baseURL, val.ValAddress.String(), "1", "3"),
			false,
			&types.QueryValidatorSlashesResponse{},
			&types.QueryValidatorSlashesResponse{
				Pagination: &query.PageResponse{},
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		resp, err := sdktestutil.GetRequest(tc.url)

		s.Run(tc.name, func() {
			if tc.expErr {
				s.Require().Error(val.ClientCtx.Codec.UnmarshalJSON(resp, tc.respType))
			} else {
				s.Require().NoError(err)
				s.Require().NoError(val.ClientCtx.Codec.UnmarshalJSON(resp, tc.respType))
				s.Require().Equal(tc.expected.String(), tc.respType.String())
			}
		})
	}
}

func (s *GRPCQueryTestSuite) TestQueryDelegatorRewardsGRPC() {
	val := s.network.Validators[0]
	baseURL := val.APIAddress

	rewards, err := sdk.ParseDecCoins("9.8stake")
	s.Require().NoError(err)

	testCases := []struct {
		name     string
		url      string
		headers  map[string]string
		expErr   bool
		respType proto.Message
		expected proto.Message
	}{
		{
			"wrong delegator address",
			fmt.Sprintf("%s/cosmos/distribution/v1beta1/delegators/%s/rewards", baseURL, "wrongDelegatorAddress"),
			map[string]string{},
			true,
			&types.QueryDelegationTotalRewardsResponse{},
			nil,
		},
		{
			"valid request",
			fmt.Sprintf("%s/cosmos/distribution/v1beta1/delegators/%s/rewards", baseURL, val.Address.String()),
			map[string]string{
				grpctypes.GRPCBlockHeightHeader: "2",
			},
			false,
			&types.QueryDelegationTotalRewardsResponse{},
			&types.QueryDelegationTotalRewardsResponse{
				Rewards: []types.DelegationDelegatorReward{
					types.NewDelegationDelegatorReward(val.ValAddress, rewards),
				},
				Total: rewards,
			},
		},
		{
			"wrong validator address(specific validator rewards)",
			fmt.Sprintf("%s/cosmos/distribution/v1beta1/delegators/%s/rewards/%s", baseURL, val.Address.String(), "wrongValAddress"),
			map[string]string{},
			true,
			&types.QueryDelegationTotalRewardsResponse{},
			nil,
		},
		{
			"valid request(specific validator rewards)",
			fmt.Sprintf("%s/cosmos/distribution/v1beta1/delegators/%s/rewards/%s", baseURL, val.Address.String(), val.ValAddress.String()),
			map[string]string{
				grpctypes.GRPCBlockHeightHeader: "2",
			},
			false,
			&types.QueryDelegationRewardsResponse{},
			&types.QueryDelegationRewardsResponse{
				Rewards: rewards,
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		resp, err := sdktestutil.GetRequestWithHeaders(tc.url, tc.headers)

		s.Run(tc.name, func() {
			if tc.expErr {
				s.Require().Error(val.ClientCtx.Codec.UnmarshalJSON(resp, tc.respType))
			} else {
				s.Require().NoError(err)
				s.Require().NoError(val.ClientCtx.Codec.UnmarshalJSON(resp, tc.respType))
				s.Require().Equal(tc.expected.String(), tc.respType.String())
			}
		})
	}
}

func (s *GRPCQueryTestSuite) TestQueryDelegatorValidatorsGRPC() {
	val := s.network.Validators[0]
	baseURL := val.APIAddress

	testCases := []struct {
		name     string
		url      string
		expErr   bool
		respType proto.Message
		expected proto.Message
	}{
		{
			"empty delegator address",
			fmt.Sprintf("%s/cosmos/distribution/v1beta1/delegators/%s/validators", baseURL, ""),
			true,
			&types.QueryDelegatorValidatorsResponse{},
			nil,
		},
		{
			"wrong delegator address",
			fmt.Sprintf("%s/cosmos/distribution/v1beta1/delegators/%s/validators", baseURL, "wrongDelegatorAddress"),
			true,
			&types.QueryDelegatorValidatorsResponse{},
			nil,
		},
		{
			"valid request",
			fmt.Sprintf("%s/cosmos/distribution/v1beta1/delegators/%s/validators", baseURL, val.Address.String()),
			false,
			&types.QueryDelegatorValidatorsResponse{},
			&types.QueryDelegatorValidatorsResponse{
				Validators: []string{val.ValAddress.String()},
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		resp, err := sdktestutil.GetRequest(tc.url)

		s.Run(tc.name, func() {
			if tc.expErr {
				s.Require().Error(val.ClientCtx.Codec.UnmarshalJSON(resp, tc.respType))
			} else {
				s.Require().NoError(err)
				s.Require().NoError(val.ClientCtx.Codec.UnmarshalJSON(resp, tc.respType))
				s.Require().Equal(tc.expected.String(), tc.respType.String())
			}
		})
	}
}

func (s *GRPCQueryTestSuite) TestQueryWithdrawAddressGRPC() {
	val := s.network.Validators[0]
	baseURL := val.APIAddress

	testCases := []struct {
		name     string
		url      string
		expErr   bool
		respType proto.Message
		expected proto.Message
	}{
		{
			"empty delegator address",
			fmt.Sprintf("%s/cosmos/distribution/v1beta1/delegators/%s/withdraw_address", baseURL, ""),
			true,
			&types.QueryDelegatorWithdrawAddressResponse{},
			nil,
		},
		{
			"wrong delegator address",
			fmt.Sprintf("%s/cosmos/distribution/v1beta1/delegators/%s/withdraw_address", baseURL, "wrongDelegatorAddress"),
			true,
			&types.QueryDelegatorWithdrawAddressResponse{},
			nil,
		},
		{
			"valid request",
			fmt.Sprintf("%s/cosmos/distribution/v1beta1/delegators/%s/withdraw_address", baseURL, val.Address.String()),
			false,
			&types.QueryDelegatorWithdrawAddressResponse{},
			&types.QueryDelegatorWithdrawAddressResponse{
				WithdrawAddress: val.Address.String(),
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		resp, err := sdktestutil.GetRequest(tc.url)

		s.Run(tc.name, func() {
			if tc.expErr {
				s.Require().Error(val.ClientCtx.Codec.UnmarshalJSON(resp, tc.respType))
			} else {
				s.Require().NoError(err)
				s.Require().NoError(val.ClientCtx.Codec.UnmarshalJSON(resp, tc.respType))
				s.Require().Equal(tc.expected.String(), tc.respType.String())
			}
		})
	}
}

func (s *GRPCQueryTestSuite) TestQueryValidatorCommunityPoolGRPC() {
	val := s.network.Validators[0]
	baseURL := val.APIAddress

	communityPool, err := sdk.ParseDecCoins("0.4stake")
	s.Require().NoError(err)

	testCases := []struct {
		name     string
		url      string
		headers  map[string]string
		expErr   bool
		respType proto.Message
		expected proto.Message
	}{
		{
			"gRPC request params with wrong validator address",
			fmt.Sprintf("%s/cosmos/distribution/v1beta1/community_pool", baseURL),
			map[string]string{
				grpctypes.GRPCBlockHeightHeader: "2",
			},
			false,
			&types.QueryCommunityPoolResponse{},
			&types.QueryCommunityPoolResponse{
				Pool: communityPool,
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		resp, err := sdktestutil.GetRequestWithHeaders(tc.url, tc.headers)

		s.Run(tc.name, func() {
			if tc.expErr {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)
				s.Require().NoError(val.ClientCtx.Codec.UnmarshalJSON(resp, tc.respType))
				s.Require().Equal(tc.expected.String(), tc.respType.String())
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
