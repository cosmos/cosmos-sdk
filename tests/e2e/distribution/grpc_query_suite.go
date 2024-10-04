package distribution

import (
	"fmt"

	"github.com/cosmos/gogoproto/proto"
	"github.com/stretchr/testify/suite"

	"cosmossdk.io/simapp"
	"cosmossdk.io/x/distribution/types"

	sdktestutil "github.com/cosmos/cosmos-sdk/testutil"
	"github.com/cosmos/cosmos-sdk/testutil/network"
	sdk "github.com/cosmos/cosmos-sdk/types"
	grpctypes "github.com/cosmos/cosmos-sdk/types/grpc"
	"github.com/cosmos/cosmos-sdk/types/query"
)

type GRPCQueryTestSuite struct {
	suite.Suite

	cfg     network.Config
	network network.NetworkI
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

// TearDownSuite cleans up the current test network after _each_ test.
func (s *GRPCQueryTestSuite) TearDownSuite() {
	s.T().Log("tearing down e2e test suite1")
	s.network.Cleanup()
}

func (s *GRPCQueryTestSuite) TestQueryParamsGRPC() {
	val := s.network.GetValidators()[0]
	baseURL := val.GetAPIAddress()

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

		resp, err := sdktestutil.GetRequest(tc.url)
		s.Run(tc.name, func() {
			s.Require().NoError(err)
			s.Require().NoError(val.GetClientCtx().Codec.UnmarshalJSON(resp, tc.respType))
			s.Require().Equal(tc.expected, tc.respType)
		})
	}
}

func (s *GRPCQueryTestSuite) TestQueryValidatorDistributionInfoGRPC() {
	val := s.network.GetValidators()[0]
	baseURL := val.GetAPIAddress()

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
			fmt.Sprintf("%s/cosmos/distribution/v1beta1/validators/%s", baseURL, val.GetValAddress().String()),
			false,
			&types.QueryValidatorDistributionInfoResponse{},
		},
	}

	for _, tc := range testCases {

		resp, err := sdktestutil.GetRequest(tc.url)
		s.Run(tc.name, func() {
			if tc.expErr {
				s.Require().Error(val.GetClientCtx().Codec.UnmarshalJSON(resp, tc.respType))
			} else {
				s.Require().NoError(err)
				s.Require().NoError(val.GetClientCtx().Codec.UnmarshalJSON(resp, tc.respType))
			}
		})
	}
}

func (s *GRPCQueryTestSuite) TestQueryOutstandingRewardsGRPC() {
	val := s.network.GetValidators()[0]
	baseURL := val.GetAPIAddress()

	rewards, err := sdk.ParseDecCoins("46.06stake")
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
			fmt.Sprintf("%s/cosmos/distribution/v1beta1/validators/%s/outstanding_rewards", baseURL, val.GetValAddress().String()),
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

		resp, err := sdktestutil.GetRequestWithHeaders(tc.url, tc.headers)
		s.Run(tc.name, func() {
			if tc.expErr {
				s.Require().Error(val.GetClientCtx().Codec.UnmarshalJSON(resp, tc.respType))
			} else {
				s.Require().NoError(err)
				s.Require().NoError(val.GetClientCtx().Codec.UnmarshalJSON(resp, tc.respType))
				s.Require().Equal(tc.expected.String(), tc.respType.String())
			}
		})
	}
}

func (s *GRPCQueryTestSuite) TestQueryValidatorCommissionGRPC() {
	val := s.network.GetValidators()[0]
	baseURL := val.GetAPIAddress()

	commission, err := sdk.ParseDecCoins("23.03stake")
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
			fmt.Sprintf("%s/cosmos/distribution/v1beta1/validators/%s/commission", baseURL, val.GetValAddress().String()),
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

		resp, err := sdktestutil.GetRequestWithHeaders(tc.url, tc.headers)
		s.Run(tc.name, func() {
			if tc.expErr {
				s.Require().Error(val.GetClientCtx().Codec.UnmarshalJSON(resp, tc.respType))
			} else {
				s.Require().NoError(err)
				s.Require().NoError(val.GetClientCtx().Codec.UnmarshalJSON(resp, tc.respType))
				s.Require().Equal(tc.expected.String(), tc.respType.String())
			}
		})
	}
}

func (s *GRPCQueryTestSuite) TestQuerySlashesGRPC() {
	val := s.network.GetValidators()[0]
	baseURL := val.GetAPIAddress()

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
			fmt.Sprintf("%s/cosmos/distribution/v1beta1/validators/%s/slashes?starting_height=%s&ending_height=%s", baseURL, val.GetValAddress().String(), "-1", "3"),
			true,
			&types.QueryValidatorSlashesResponse{},
			nil,
		},
		{
			"invalid start height",
			fmt.Sprintf("%s/cosmos/distribution/v1beta1/validators/%s/slashes?starting_height=%s&ending_height=%s", baseURL, val.GetValAddress().String(), "1", "-3"),
			true,
			&types.QueryValidatorSlashesResponse{},
			nil,
		},
		{
			"valid request get slashes",
			fmt.Sprintf("%s/cosmos/distribution/v1beta1/validators/%s/slashes?starting_height=%s&ending_height=%s", baseURL, val.GetValAddress().String(), "1", "3"),
			false,
			&types.QueryValidatorSlashesResponse{},
			&types.QueryValidatorSlashesResponse{
				Pagination: &query.PageResponse{},
			},
		},
	}

	for _, tc := range testCases {

		resp, err := sdktestutil.GetRequest(tc.url)

		s.Run(tc.name, func() {
			if tc.expErr {
				s.Require().Error(val.GetClientCtx().Codec.UnmarshalJSON(resp, tc.respType))
			} else {
				s.Require().NoError(err)
				s.Require().NoError(val.GetClientCtx().Codec.UnmarshalJSON(resp, tc.respType))
				s.Require().Equal(tc.expected.String(), tc.respType.String())
			}
		})
	}
}

func (s *GRPCQueryTestSuite) TestQueryDelegatorRewardsGRPC() {
	val := s.network.GetValidators()[0]
	baseURL := val.GetAPIAddress()

	rewards, err := sdk.ParseDecCoins("23.03stake")
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
			fmt.Sprintf("%s/cosmos/distribution/v1beta1/delegators/%s/rewards", baseURL, val.GetAddress().String()),
			map[string]string{
				grpctypes.GRPCBlockHeightHeader: "2",
			},
			false,
			&types.QueryDelegationTotalRewardsResponse{},
			&types.QueryDelegationTotalRewardsResponse{
				Rewards: []types.DelegationDelegatorReward{
					types.NewDelegationDelegatorReward(val.GetValAddress().String(), rewards),
				},
				Total: rewards,
			},
		},
		{
			"wrong validator address(specific validator rewards)",
			fmt.Sprintf("%s/cosmos/distribution/v1beta1/delegators/%s/rewards/%s", baseURL, val.GetAddress().String(), "wrongValAddress"),
			map[string]string{},
			true,
			&types.QueryDelegationTotalRewardsResponse{},
			nil,
		},
		{
			"valid request(specific validator rewards)",
			fmt.Sprintf("%s/cosmos/distribution/v1beta1/delegators/%s/rewards/%s", baseURL, val.GetAddress().String(), val.GetValAddress().String()),
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

		resp, err := sdktestutil.GetRequestWithHeaders(tc.url, tc.headers)

		s.Run(tc.name, func() {
			if tc.expErr {
				s.Require().Error(val.GetClientCtx().Codec.UnmarshalJSON(resp, tc.respType))
			} else {
				s.Require().NoError(err)
				s.Require().NoError(val.GetClientCtx().Codec.UnmarshalJSON(resp, tc.respType))
				s.Require().Equal(tc.expected.String(), tc.respType.String())
			}
		})
	}
}

func (s *GRPCQueryTestSuite) TestQueryDelegatorValidatorsGRPC() {
	val := s.network.GetValidators()[0]
	baseURL := val.GetAPIAddress()

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
			fmt.Sprintf("%s/cosmos/distribution/v1beta1/delegators/%s/validators", baseURL, val.GetAddress().String()),
			false,
			&types.QueryDelegatorValidatorsResponse{},
			&types.QueryDelegatorValidatorsResponse{
				Validators: []string{val.GetValAddress().String()},
			},
		},
	}

	for _, tc := range testCases {

		resp, err := sdktestutil.GetRequest(tc.url)

		s.Run(tc.name, func() {
			if tc.expErr {
				s.Require().Error(val.GetClientCtx().Codec.UnmarshalJSON(resp, tc.respType))
			} else {
				s.Require().NoError(err)
				s.Require().NoError(val.GetClientCtx().Codec.UnmarshalJSON(resp, tc.respType))
				s.Require().Equal(tc.expected.String(), tc.respType.String())
			}
		})
	}
}

func (s *GRPCQueryTestSuite) TestQueryWithdrawAddressGRPC() {
	val := s.network.GetValidators()[0]
	baseURL := val.GetAPIAddress()

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
			fmt.Sprintf("%s/cosmos/distribution/v1beta1/delegators/%s/withdraw_address", baseURL, val.GetAddress().String()),
			false,
			&types.QueryDelegatorWithdrawAddressResponse{},
			&types.QueryDelegatorWithdrawAddressResponse{
				WithdrawAddress: val.GetAddress().String(),
			},
		},
	}

	for _, tc := range testCases {

		resp, err := sdktestutil.GetRequest(tc.url)

		s.Run(tc.name, func() {
			if tc.expErr {
				s.Require().Error(val.GetClientCtx().Codec.UnmarshalJSON(resp, tc.respType))
			} else {
				s.Require().NoError(err)
				s.Require().NoError(val.GetClientCtx().Codec.UnmarshalJSON(resp, tc.respType))
				s.Require().Equal(tc.expected.String(), tc.respType.String())
			}
		})
	}
}
