package testutil

import (
	"fmt"

	"github.com/gogo/protobuf/proto"

	"github.com/cosmos/cosmos-sdk/testutil"
	"github.com/cosmos/cosmos-sdk/x/params/types/proposal"
)

func (s *IntegrationTestSuite) TestQueryParamsGRPC() {
	val := s.network.Validators[0]
	baseURL := val.APIAddress

	testCases := []struct {
		name     string
		url      string
		headers  map[string]string
		expErr   bool
		respType proto.Message
		expected proto.Message
	}{
		{
			"with no subspace, key",
			fmt.Sprintf("%s/cosmos/params/v1beta1/params?subspace=%s&key=%s", baseURL, "", ""),
			map[string]string{},
			true,
			&proposal.QueryParamsResponse{},
			&proposal.QueryParamsResponse{
				Param: proposal.ParamChange{
					Subspace: "staking",
					Key:      "MaxValidators",
					Value:    "100",
				},
			},
		},
		{
			"with wrong subspace",
			fmt.Sprintf("%s/cosmos/params/v1beta1/params?subspace=%s&key=%s", baseURL, "wrongSubspace", "MaxValidators"),
			map[string]string{},
			true,
			&proposal.QueryParamsResponse{},
			&proposal.QueryParamsResponse{
				Param: proposal.ParamChange{
					Subspace: "staking",
					Key:      "MaxValidators",
					Value:    "100",
				},
			},
		},
		{
			"with wrong key",
			fmt.Sprintf("%s/cosmos/params/v1beta1/params?subspace=%s&key=%s", baseURL, "staking", "wrongKey"),
			map[string]string{},
			false,
			&proposal.QueryParamsResponse{},
			&proposal.QueryParamsResponse{
				Param: proposal.ParamChange{
					Subspace: "staking",
					Key:      "wrongKey",
					Value:    "",
				},
			},
		},
		{
			"params",
			fmt.Sprintf("%s/cosmos/params/v1beta1/params?subspace=%s&key=%s", baseURL, "staking", "MaxValidators"),
			map[string]string{},
			false,
			&proposal.QueryParamsResponse{},
			&proposal.QueryParamsResponse{
				Param: proposal.ParamChange{
					Subspace: "staking",
					Key:      "MaxValidators",
					Value:    "100",
				},
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		s.Run(tc.name, func() {
			resp, err := testutil.GetRequestWithHeaders(tc.url, tc.headers)
			s.Require().NoError(err)

			err = val.ClientCtx.Codec.UnmarshalJSON(resp, tc.respType)

			if tc.expErr {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)
				s.Require().Equal(tc.expected.String(), tc.respType.String())
			}
		})
	}
}
