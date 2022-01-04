// +build norace

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
	genesisState := cfg.GenesisState
	cfg.NumValidators = 1

	var bankGenesis types.GenesisState
	s.Require().NoError(cfg.Codec.UnmarshalJSON(genesisState[types.ModuleName], &bankGenesis))

	bankGenesis.DenomMetadata = []types.Metadata{
		{
			Description: "The native staking token of the Cosmos Hub.",
			DenomUnits: []*types.DenomUnit{
				{
					Denom:    "uatom",
					Exponent: 0,
					Aliases:  []string{"microatom"},
				},
				{
					Denom:    "atom",
					Exponent: 6,
					Aliases:  []string{"ATOM"},
				},
			},
			Base:    "uatom",
			Display: "atom",
		},
	}

	bankGenesisBz, err := cfg.Codec.MarshalJSON(&bankGenesis)
	s.Require().NoError(err)
	genesisState[types.ModuleName] = bankGenesisBz
	cfg.GenesisState = genesisState

	s.cfg = cfg
	s.network = network.New(s.T(), cfg)

	_, err = s.network.WaitForHeight(2)
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
		name      string
		url       string
		expHeight int64
		respType  fmt.Stringer
		expected  fmt.Stringer
	}{
		{
			"total account balance",
			fmt.Sprintf("%s/bank/balances/%s", baseURL, val.Address),
			-1,
			&sdk.Coins{},
			sdk.NewCoins(
				sdk.NewCoin(fmt.Sprintf("%stoken", val.Moniker), s.cfg.AccountTokens),
				sdk.NewCoin(s.cfg.BondDenom, s.cfg.StakingTokens.Sub(s.cfg.BondedTokens)),
			),
		},
		{
			"total account balance with height",
			fmt.Sprintf("%s/bank/balances/%s?height=1", baseURL, val.Address),
			1,
			&sdk.Coins{},
			sdk.NewCoins(
				sdk.NewCoin(fmt.Sprintf("%stoken", val.Moniker), s.cfg.AccountTokens),
				sdk.NewCoin(s.cfg.BondDenom, s.cfg.StakingTokens.Sub(s.cfg.BondedTokens)),
			),
		},
		{
			"total account balance of a specific denom",
			fmt.Sprintf("%s/bank/balances/%s?denom=%s", baseURL, val.Address, s.cfg.BondDenom),
			-1,
			&sdk.Coin{},
			sdk.NewCoin(s.cfg.BondDenom, s.cfg.StakingTokens.Sub(s.cfg.BondedTokens)),
		},
		{
			"total account balance of a bogus denom",
			fmt.Sprintf("%s/bank/balances/%s?denom=foobar", baseURL, val.Address),
			-1,
			&sdk.Coin{},
			sdk.NewCoin("foobar", sdk.ZeroInt()),
		},
	}

	for _, tc := range testCases {
		tc := tc
		s.Run(tc.name, func() {
			respJSON, err := rest.GetRequest(tc.url)
			s.Require().NoError(err)

			var resp = rest.ResponseWithHeight{}
			err = val.ClientCtx.LegacyAmino.UnmarshalJSON(respJSON, &resp)
			s.Require().NoError(err)

			// Check height.
			if tc.expHeight >= 0 {
				s.Require().Equal(resp.Height, tc.expHeight)
			} else {
				// To avoid flakiness, just test that height is positive.
				s.Require().Greater(resp.Height, int64(0))
			}

			// Check result.
			s.Require().NoError(val.ClientCtx.LegacyAmino.UnmarshalJSON(resp.Result, tc.respType))
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

func TestIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(IntegrationTestSuite))
}
