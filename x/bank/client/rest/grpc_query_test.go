package rest_test

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	grpctypes "github.com/cosmos/cosmos-sdk/types/grpc"
	"github.com/cosmos/cosmos-sdk/types/rest"
	"github.com/cosmos/cosmos-sdk/x/bank/types"
)

func (s *IntegrationTestSuite) TestTotalSupplyGRPCHandler() {
	val := s.network.Validators[0]
	baseURL := val.APIAddress

	testCases := []struct {
		name     string
		url      string
		headers  map[string]string
		respType fmt.Stringer
		expected fmt.Stringer
	}{
		{
			"test GRPC total supply",
			fmt.Sprintf("%s/cosmos/bank/v1beta1/supply?height=1", baseURL),
			map[string]string{
				grpctypes.GRPCBlockHeightHeader: "1",
			},
			&types.QueryTotalSupplyResponse{},
			&types.QueryTotalSupplyResponse{
				Supply: sdk.NewCoins(
					sdk.NewCoin(fmt.Sprintf("%stoken", val.Moniker), s.cfg.AccountTokens),
					sdk.NewCoin(s.cfg.BondDenom, s.cfg.StakingTokens.Add(sdk.NewInt(10))),
				),
			},
		},
		{
			"GRPC total supply of a specific denom",
			fmt.Sprintf("%s/cosmos/bank/v1beta1/supply/%s?height=1", baseURL, s.cfg.BondDenom),
			map[string]string{
				grpctypes.GRPCBlockHeightHeader: "1",
			},
			&types.QuerySupplyOfResponse{},
			&types.QuerySupplyOfResponse{
				Amount: sdk.NewCoin(s.cfg.BondDenom, s.cfg.StakingTokens.Add(sdk.NewInt(10))),
			},
		},
		{
			"GRPC total supply of a bogus denom",
			fmt.Sprintf("%s/cosmos/bank/v1beta1/supply/foobar?height=1", baseURL),
			map[string]string{
				grpctypes.GRPCBlockHeightHeader: "1",
			},
			&types.QuerySupplyOfResponse{},
			&types.QuerySupplyOfResponse{
				Amount: sdk.NewCoin("foobar", sdk.ZeroInt()),
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		s.Run(tc.name, func() {
			resp, err := rest.GetRequestWithHeaders(tc.url, tc.headers)
			s.Require().NoError(err)

			s.Require().NoError(val.ClientCtx.LegacyAmino.UnmarshalJSON(resp, tc.respType))
			s.Require().Equal(tc.expected.String(), tc.respType.String())
		})
	}
}
