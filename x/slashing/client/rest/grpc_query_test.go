package rest_test

import (
	"encoding/base64"
	"fmt"
	"testing"
	"time"

	"github.com/cosmos/cosmos-sdk/testutil"
	"github.com/cosmos/cosmos-sdk/testutil/network"
	sdk "github.com/cosmos/cosmos-sdk/types"
	grpctypes "github.com/cosmos/cosmos-sdk/types/grpc"
	"github.com/cosmos/cosmos-sdk/types/query"
	"github.com/cosmos/cosmos-sdk/x/slashing/types"
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

func (s *IntegrationTestSuite) TearDownSuite() {
	s.T().Log("tearing down integration test suite")
	s.network.Cleanup()
}

func (s *IntegrationTestSuite) TestGRPCQueries() {
	val := s.network.Validators[0]
	baseURL := val.APIAddress

	// TODO: need to pass bech32 string instead of base64 encoding string
	// ref: https://github.com/cosmos/cosmos-sdk/issues/7195
	consAddrBase64 := base64.URLEncoding.EncodeToString(sdk.ConsAddress(val.PubKey.Address()))

	testCases := []struct {
		name     string
		url      string
		headers  map[string]string
		expErr   bool
		respType proto.Message
		expected proto.Message
	}{
		{
			"get signing infos (height specific)",
			fmt.Sprintf("%s/cosmos/slashing/v1beta1/signing_infos", baseURL),
			map[string]string{
				grpctypes.GRPCBlockHeightHeader: "1",
			},
			false,
			&types.QuerySigningInfosResponse{},
			&types.QuerySigningInfosResponse{
				Info: []types.ValidatorSigningInfo{
					types.ValidatorSigningInfo{
						Address:     sdk.ConsAddress(val.PubKey.Address()),
						JailedUntil: time.Unix(0, 0),
					},
				},
				Pagination: &query.PageResponse{
					Total: uint64(1),
				},
			},
		},
		{
			"get signing info (height specific)",
			fmt.Sprintf("%s/cosmos/slashing/v1beta1/signing_infos/%s", baseURL, consAddrBase64),
			map[string]string{
				grpctypes.GRPCBlockHeightHeader: "1",
			},
			false,
			&types.QuerySigningInfoResponse{},
			&types.QuerySigningInfoResponse{
				ValSigningInfo: types.ValidatorSigningInfo{
					Address:     sdk.ConsAddress(val.PubKey.Address()),
					JailedUntil: time.Unix(0, 0),
				},
			},
		},
		{
			"get signing info wrong address",
			fmt.Sprintf("%s/cosmos/slashing/v1beta1/signing_infos/%s", baseURL, "wrongAddress"),
			map[string]string{},
			true,
			&types.QuerySigningInfoResponse{},
			nil,
		},
		{
			"params",
			fmt.Sprintf("%s/cosmos/slashing/v1beta1/params", baseURL),
			map[string]string{},
			false,
			&types.QueryParamsResponse{},
			&types.QueryParamsResponse{
				Params: types.DefaultParams(),
			},
		},
	}

	for _, tc := range testCases {
		tc := tc

		s.Run(tc.name, func() {
			resp, err := testutil.GetRequestWithHeaders(tc.url, tc.headers)
			s.Require().NoError(err)

			err = val.ClientCtx.JSONMarshaler.UnmarshalJSON(resp, tc.respType)

			if tc.expErr {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)
				s.Require().Equal(tc.expected.String(), tc.respType.String())
			}
		})
	}
}

func TestIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(IntegrationTestSuite))
}
