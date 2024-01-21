package mint

import (
	"fmt"

	"github.com/cosmos/gogoproto/proto"

	"cosmossdk.io/math"
	minttypes "cosmossdk.io/x/mint/types"

	"github.com/cosmos/cosmos-sdk/testutil"
	grpctypes "github.com/cosmos/cosmos-sdk/types/grpc"
)

func (s *E2ETestSuite) TestQueryGRPC() {
	val := s.network.GetValidators()[0]
	baseURL := val.GetAPIAddress()
	testCases := []struct {
		name     string
		url      string
		headers  map[string]string
		respType proto.Message
		expected proto.Message
	}{
		{
			"gRPC request inflation",
			fmt.Sprintf("%s/cosmos/mint/v1beta1/inflation", baseURL),
			map[string]string{},
			&minttypes.QueryInflationResponse{},
			&minttypes.QueryInflationResponse{
				Inflation: math.LegacyNewDec(1),
			},
		},
		{
			"gRPC request annual provisions",
			fmt.Sprintf("%s/cosmos/mint/v1beta1/annual_provisions", baseURL),
			map[string]string{
				grpctypes.GRPCBlockHeightHeader: "1",
			},
			&minttypes.QueryAnnualProvisionsResponse{},
			&minttypes.QueryAnnualProvisionsResponse{
				AnnualProvisions: math.LegacyNewDec(500000000),
			},
		},
	}
	for _, tc := range testCases {
		resp, err := testutil.GetRequestWithHeaders(tc.url, tc.headers)
		s.Run(tc.name, func() {
			s.Require().NoError(err)
			s.Require().NoError(val.GetClientCtx().Codec.UnmarshalJSON(resp, tc.respType))
			s.Require().Equal(tc.expected.String(), tc.respType.String())
		})
	}
}
