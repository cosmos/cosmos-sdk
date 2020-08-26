package rest_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/cosmos/cosmos-sdk/testutil"
	"github.com/cosmos/cosmos-sdk/testutil/network"
	sdk "github.com/cosmos/cosmos-sdk/types"
	grpctypes "github.com/cosmos/cosmos-sdk/types/grpc"
	"github.com/cosmos/cosmos-sdk/x/staking/types"
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

func (s *IntegrationTestSuite) TearDownSuite() {
	s.T().Log("tearing down integration test suite")
	s.network.Cleanup()
}

func (s *IntegrationTestSuite) TestQueryValidatorsGRPCHandler() {
	val := s.network.Validators[0]
	baseURL := val.APIAddress

	testCases := []struct {
		name    string
		url     string
		headers map[string]string
		error   bool
	}{
		{
			"test query validators gRPC route with invalid status",
			fmt.Sprintf("%s/cosmos/staking/v1beta1/validators?status=active", baseURL),
			map[string]string{
				grpctypes.GRPCBlockHeightHeader: "1",
			},
			true,
		},
		{
			"test query validators gRPC route without status query param",
			fmt.Sprintf("%s/cosmos/staking/v1beta1/validators", baseURL),
			map[string]string{
				grpctypes.GRPCBlockHeightHeader: "1",
			},
			false,
		},
		{
			"test query validators gRPC route with valid status",
			fmt.Sprintf("%s/cosmos/staking/v1beta1/validators?status=%s", baseURL, sdk.Bonded.String()),
			map[string]string{
				grpctypes.GRPCBlockHeightHeader: "1",
			},
			false,
		},
	}

	for _, tc := range testCases {
		tc := tc
		s.Run(tc.name, func() {
			resp, err := testutil.GetRequestWithHeaders(tc.url, tc.headers)
			s.Require().NoError(err)

			var valRes types.QueryValidatorsResponse
			err = val.ClientCtx.JSONMarshaler.UnmarshalJSON(resp, &valRes)

			if tc.error {
				s.Require().Error(err)
				s.Require().Nil(valRes.Validators)
				s.Require().Equal(0, len(valRes.Validators))
			} else {
				s.Require().NoError(err)
				s.Require().NotNil(valRes.Validators)
				s.Require().Equal(len(s.network.Validators), len(valRes.Validators))
			}
		})
	}
}

func TestIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(IntegrationTestSuite))
}
