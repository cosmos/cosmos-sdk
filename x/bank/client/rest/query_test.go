package rest_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/cosmos/cosmos-sdk/testutil/network"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/rest"
	"github.com/cosmos/cosmos-sdk/x/bank/types"
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

func (s *IntegrationTestSuite) TestQueryBalancesRequestHandlerFn() {
	val := s.network.Validators[0]
	baseURL := val.APIAddress

	testCases := []struct {
		name     string
		url      string
		respType fmt.Stringer
		expected fmt.Stringer
	}{
		{
			"total account balance",
			fmt.Sprintf("%s/bank/balances/%s?height=1", baseURL, val.Address),
			&sdk.Coins{},
			sdk.NewCoins(
				sdk.NewCoin(fmt.Sprintf("%stoken", val.Moniker), s.cfg.AccountTokens),
				sdk.NewCoin(s.cfg.BondDenom, s.cfg.StakingTokens.Sub(s.cfg.BondedTokens)),
			),
		},
		{
			"total account balance of a specific denom",
			fmt.Sprintf("%s/bank/balances/%s?height=1&denom=%s", baseURL, val.Address, s.cfg.BondDenom),
			&sdk.Coin{},
			sdk.NewCoin(s.cfg.BondDenom, s.cfg.StakingTokens.Sub(s.cfg.BondedTokens)),
		},
		{
			"total account balance of a bogus denom",
			fmt.Sprintf("%s/bank/balances/%s?height=1&denom=foobar", baseURL, val.Address),
			&sdk.Coin{},
			sdk.NewCoin("foobar", sdk.ZeroInt()),
		},
	}

	for _, tc := range testCases {
		tc := tc
		s.Run(tc.name, func() {
			resp, err := rest.GetRequest(tc.url)
			s.Require().NoError(err)

			bz, err := rest.ParseResponseWithHeight(val.ClientCtx.LegacyAmino, resp)
			s.Require().NoError(err)
			s.Require().NoError(val.ClientCtx.LegacyAmino.UnmarshalJSON(bz, tc.respType))
			s.Require().Equal(tc.expected.String(), tc.respType.String())
		})
	}
}

func (s *IntegrationTestSuite) TestTotalSupplyHandlerFn() {
	val := s.network.Validators[0]
	baseURL := val.APIAddress

	testCases := []struct {
		name     string
		url      string
		respType fmt.Stringer
		expected fmt.Stringer
	}{
		{
			"total supply",
			fmt.Sprintf("%s/bank/total?height=1", baseURL),
			&sdk.Coins{},
			sdk.NewCoins(
				sdk.NewCoin(fmt.Sprintf("%stoken", val.Moniker), s.cfg.AccountTokens),
				sdk.NewCoin(s.cfg.BondDenom, s.cfg.StakingTokens.Add(sdk.NewInt(10))),
			),
		},
		{
			"total supply of a specific denom",
			fmt.Sprintf("%s/bank/total/%s?height=1", baseURL, s.cfg.BondDenom),
			&sdk.Coin{},
			sdk.NewCoin(s.cfg.BondDenom, s.cfg.StakingTokens.Add(sdk.NewInt(10))),
		},
		{
			"total supply of a bogus denom",
			fmt.Sprintf("%s/bank/total/foobar?height=1", baseURL),
			&sdk.Coin{},
			sdk.NewCoin("foobar", sdk.ZeroInt()),
		},
	}

	for _, tc := range testCases {
		tc := tc
		s.Run(tc.name, func() {
			resp, err := rest.GetRequest(tc.url)
			s.Require().NoError(err)

			bz, err := rest.ParseResponseWithHeight(val.ClientCtx.LegacyAmino, resp)
			s.Require().NoError(err)
			s.Require().NoError(val.ClientCtx.LegacyAmino.UnmarshalJSON(bz, tc.respType))
			s.Require().Equal(tc.expected.String(), tc.respType.String())
		})
	}
}

func (s *IntegrationTestSuite) TestTotalSupplyGRPCHandler() {
	val := s.network.Validators[0]
	baseURL := val.APIAddress

	testCases := []struct {
		name     string
		url      string
		respType fmt.Stringer
		expected fmt.Stringer
	}{
		{
			"test GRPC total supply",
			fmt.Sprintf("%s/cosmos/bank/v1beta1/supply", baseURL),
			&types.QueryTotalSupplyResponse{},
			&types.QueryTotalSupplyResponse{
				Supply: sdk.NewCoins(
					sdk.NewCoin(fmt.Sprintf("%stoken", val.Moniker), s.cfg.AccountTokens),
					sdk.NewCoin(s.cfg.BondDenom, s.cfg.StakingTokens.Add(sdk.NewInt50))),
				),
			},
		},
		{
			"GRPC total supply of a specific denom",
			fmt.Sprintf("%s/cosmos/bank/v1beta1/supply/%s", baseURL, s.cfg.BondDenom),
			&types.QuerySupplyOfResponse{},
			&types.QuerySupplyOfResponse{
				Amount: sdk.NewCoin(s.cfg.BondDenom, s.cfg.StakingTokens.Add(sdk.NewInt(50))),
			},
		},
		{
			"GRPC total supply of a bogus denom",
			fmt.Sprintf("%s/cosmos/bank/v1beta1/supply/foobar", baseURL),
			&types.QuerySupplyOfResponse{},
			&types.QuerySupplyOfResponse{
				Amount: sdk.NewCoin("foobar", sdk.ZeroInt()),
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		s.Run(tc.name, func() {
			resp, err := rest.GetRequest(tc.url)
			s.Require().NoError(err)

			s.Require().NoError(val.ClientCtx.LegacyAmino.UnmarshalJSON(resp, tc.respType))
			s.Require().Equal(tc.expected.String(), tc.respType.String())
		})
	}
}

func TestIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(IntegrationTestSuite))
}
