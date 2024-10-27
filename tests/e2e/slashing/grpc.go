package testutil

import (
	"fmt"
	"time"

	"github.com/cosmos/gogoproto/proto"

	"github.com/cosmos/cosmos-sdk/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	grpctypes "github.com/cosmos/cosmos-sdk/types/grpc"
	"github.com/cosmos/cosmos-sdk/types/query"
	"github.com/cosmos/cosmos-sdk/x/slashing/types"
)

func (s *E2ETestSuite) TestGRPCQueries() {
	val := s.network.Validators[0]
	baseURL := val.APIAddress

	consAddr := sdk.ConsAddress(val.PubKey.Address()).String()

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
					{
						Address:     sdk.ConsAddress(val.PubKey.Address()).String(),
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
			fmt.Sprintf("%s/cosmos/slashing/v1beta1/signing_infos/%s", baseURL, consAddr),
			map[string]string{
				grpctypes.GRPCBlockHeightHeader: "1",
			},
			false,
			&types.QuerySigningInfoResponse{},
			&types.QuerySigningInfoResponse{
				ValSigningInfo: types.ValidatorSigningInfo{
					Address:     sdk.ConsAddress(val.PubKey.Address()).String(),
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
