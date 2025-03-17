package protocolpool

import (
	"fmt"

	"github.com/cosmos/gogoproto/proto"
	"github.com/stretchr/testify/suite"

	"cosmossdk.io/math"

	sdktestutil "github.com/cosmos/cosmos-sdk/testutil"
	"github.com/cosmos/cosmos-sdk/testutil/network"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/protocolpool/types"
)

type GRPCQueryTestSuite struct {
	suite.Suite
	cfg     network.Config
	network *network.Network
}

func NewGRPCQueryTestSuite() *GRPCQueryTestSuite {
	return &GRPCQueryTestSuite{}
}

func (s *GRPCQueryTestSuite) SetupSuite() {
	s.T().Log("setting up grpc x/protocolpool e2e test suite")

	cfg := initNetworkConfig(s.T())

	cfg.NumValidators = 1
	s.cfg = cfg

	var err error
	s.network, err = network.New(s.T(), s.T().TempDir(), cfg)
	s.Require().NoError(err)

	s.Require().NoError(s.network.WaitForNextBlock())
}

// TearDownSuite cleans up the curret test network after _each_ test.
func (s *GRPCQueryTestSuite) TearDownSuite() {
	s.T().Log("tearing down grpc x/protocolpool e2e test suite")
	s.network.Cleanup()
}

func (s *GRPCQueryTestSuite) TestQueryParamsGRPC() {
	val := s.network.Validators[0]
	baseURL := val.APIAddress
	expectedParams := types.DefaultParams()

	testCases := []struct {
		name     string
		url      string
		expErr   bool
		respType proto.Message
		expected proto.Message
	}{
		{
			name:     "gRPC request params",
			url:      fmt.Sprintf("%s/cosmos/protocolpool/v1/params", baseURL),
			respType: &types.QueryParamsResponse{},
			expected: &types.QueryParamsResponse{
				Params: *expectedParams,
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		resp, err := sdktestutil.GetRequest(tc.url)
		s.Run(tc.name, func() {
			if tc.expErr {
				s.Require().Error(err)
				return
			}
			s.Require().NoError(err)
			s.Require().NoError(val.ClientCtx.Codec.UnmarshalJSON(resp, tc.respType))
			s.Require().Equal(tc.expected, tc.respType)
		})
	}
}

func (s *GRPCQueryTestSuite) TestQueryCommunityPoolGRPC() {
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
			name:     "gRPC request community pool",
			url:      fmt.Sprintf("%s/cosmos/protocolpool/v1/community_pool", baseURL),
			respType: &types.QueryCommunityPoolResponse{},
			expected: &types.QueryCommunityPoolResponse{
				Pool: sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, math.ZeroInt())),
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		resp, err := sdktestutil.GetRequest(tc.url)
		s.Run(tc.name, func() {
			if tc.expErr {
				s.Require().Error(err)
				return
			}
			s.Require().NoError(err)
			s.Require().NoError(val.ClientCtx.Codec.UnmarshalJSON(resp, tc.respType))
			s.Require().Equal(tc.expected, tc.respType)
		})
	}
}
