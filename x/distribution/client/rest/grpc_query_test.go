package rest_test

import (
	"encoding/base64"
	"fmt"
	"testing"

	"github.com/cosmos/cosmos-sdk/testutil"
	"github.com/cosmos/cosmos-sdk/testutil/network"
	sdk "github.com/cosmos/cosmos-sdk/types"
	grpctypes "github.com/cosmos/cosmos-sdk/types/grpc"
	"github.com/cosmos/cosmos-sdk/types/rest"
	"github.com/cosmos/cosmos-sdk/x/distribution/types"
	"github.com/gogo/protobuf/proto"
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

	_, err := s.network.WaitForHeight(1)
	s.Require().NoError(err)
}

func (s *IntegrationTestSuite) TestQueryParamsGRPC() {
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
		resp, err := rest.GetRequest(tc.url)
		s.Run(tc.name, func() {
			s.Require().NoError(err)
			s.Require().NoError(val.ClientCtx.JSONMarshaler.UnmarshalJSON(resp, tc.respType))
			s.Require().Equal(tc.expected, tc.respType)
		})
	}
}

func (s *IntegrationTestSuite) TestQueryOutstandingRewardsGRPC() {
	val := s.network.Validators[0]
	baseURL := val.APIAddress

	rewards, _ := sdk.ParseDecCoins("19.6stake")

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
			map[string]string{
				grpctypes.GRPCBlockHeightHeader: "1",
			},
			false,
			&types.QueryValidatorOutstandingRewardsResponse{},
			&types.QueryValidatorOutstandingRewardsResponse{},
		},
		{
			"gRPC request params valid address",
			fmt.Sprintf("%s/cosmos/distribution/v1beta1/validators/%s/outstanding_rewards", baseURL, base64.URLEncoding.EncodeToString(val.ValAddress)),
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
		resp, err := testutil.GetRequestWithHeaders(tc.url, tc.headers)
		s.Run(tc.name, func() {
			if tc.expErr {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)
				s.Require().NoError(val.ClientCtx.JSONMarshaler.UnmarshalJSON(resp, tc.respType))
				s.Require().Equal(tc.expected.String(), tc.respType.String())
			}
		})
	}
}

func TestIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(IntegrationTestSuite))
}
