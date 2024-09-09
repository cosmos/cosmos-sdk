package client

import (
	"fmt"

	"github.com/cosmos/gogoproto/proto"
	"github.com/stretchr/testify/suite"

	"cosmossdk.io/core/address"
	"cosmossdk.io/math"
	"cosmossdk.io/x/bank/types"

	addresscodec "github.com/cosmos/cosmos-sdk/codec/address"
	"github.com/cosmos/cosmos-sdk/testutil"
	"github.com/cosmos/cosmos-sdk/testutil/network"
	sdk "github.com/cosmos/cosmos-sdk/types"
	grpctypes "github.com/cosmos/cosmos-sdk/types/grpc"
	"github.com/cosmos/cosmos-sdk/types/query"
)

type E2ETestSuite struct {
	suite.Suite

	cfg     network.Config
	ac      address.Codec
	network network.NetworkI
}

func NewE2ETestSuite(cfg network.Config) *E2ETestSuite {
	return &E2ETestSuite{cfg: cfg}
}

func (s *E2ETestSuite) SetupSuite() {
	s.T().Log("setting up e2e test suite")

	genesisState := s.cfg.GenesisState
	var bankGenesis types.GenesisState
	s.Require().NoError(s.cfg.Codec.UnmarshalJSON(genesisState[types.ModuleName], &bankGenesis))

	bankGenesis.DenomMetadata = []types.Metadata{
		{
			Name:        "Cosmos Hub Atom",
			Symbol:      "ATOM",
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
		{
			Name:        "Ethereum",
			Symbol:      "ETH",
			Description: "Ethereum mainnet token",
			DenomUnits: []*types.DenomUnit{
				{
					Denom:    "wei",
					Exponent: 0,
				},
				{
					Denom:    "eth",
					Exponent: 6,
					Aliases:  []string{"ETH"},
				},
			},
			Base:    "wei",
			Display: "eth",
		},
	}

	bankGenesisBz, err := s.cfg.Codec.MarshalJSON(&bankGenesis)
	s.Require().NoError(err)
	genesisState[types.ModuleName] = bankGenesisBz
	s.cfg.GenesisState = genesisState

	s.network, err = network.New(s.T(), s.T().TempDir(), s.cfg)
	s.Require().NoError(err)
	s.Require().NoError(s.network.WaitForNextBlock())
	s.ac = addresscodec.NewBech32Codec("cosmos")
}

func (s *E2ETestSuite) TearDownSuite() {
	s.T().Log("tearing down e2e test suite")
	s.network.Cleanup()
}

func (s *E2ETestSuite) TestTotalSupplyGRPCHandler() {
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
			"test GRPC total supply",
			fmt.Sprintf("%s/cosmos/bank/v1beta1/supply", baseURL),
			map[string]string{
				grpctypes.GRPCBlockHeightHeader: "1",
			},
			&types.QueryTotalSupplyResponse{},
			&types.QueryTotalSupplyResponse{
				Supply: sdk.NewCoins(
					sdk.NewCoin(fmt.Sprintf("%stoken", val.GetMoniker()), s.cfg.AccountTokens),
					sdk.NewCoin(s.cfg.BondDenom, s.cfg.StakingTokens.Add(math.NewInt(47))),
				),
				Pagination: &query.PageResponse{
					Total: 2,
				},
			},
		},
		{
			"GRPC total supply of a specific denom",
			fmt.Sprintf("%s/cosmos/bank/v1beta1/supply/by_denom?denom=%s", baseURL, s.cfg.BondDenom),
			map[string]string{
				grpctypes.GRPCBlockHeightHeader: "1",
			},
			&types.QuerySupplyOfResponse{},
			&types.QuerySupplyOfResponse{
				Amount: sdk.NewCoin(s.cfg.BondDenom, s.cfg.StakingTokens.Add(math.NewInt(47))),
			},
		},
		{
			"Query for `height` > 1",
			fmt.Sprintf("%s/cosmos/bank/v1beta1/supply/by_denom?denom=%s", baseURL, s.cfg.BondDenom),
			map[string]string{
				grpctypes.GRPCBlockHeightHeader: "2",
			},
			&types.QuerySupplyOfResponse{},
			&types.QuerySupplyOfResponse{
				Amount: sdk.NewCoin(s.cfg.BondDenom, s.cfg.StakingTokens.Add(math.NewInt(47))),
			},
		},
		{
			"Query params shouldn't be considered as height",
			fmt.Sprintf("%s/cosmos/bank/v1beta1/supply/by_denom?denom=%s&height=2", baseURL, s.cfg.BondDenom),
			map[string]string{
				grpctypes.GRPCBlockHeightHeader: "1",
			},
			&types.QuerySupplyOfResponse{},
			&types.QuerySupplyOfResponse{
				Amount: sdk.NewCoin(s.cfg.BondDenom, s.cfg.StakingTokens.Add(math.NewInt(47))),
			},
		},
		{
			"GRPC total supply of a bogus denom",
			fmt.Sprintf("%s/cosmos/bank/v1beta1/supply/by_denom?denom=foobar", baseURL),
			map[string]string{
				grpctypes.GRPCBlockHeightHeader: "1",
			},
			&types.QuerySupplyOfResponse{},
			&types.QuerySupplyOfResponse{
				Amount: sdk.NewCoin("foobar", math.ZeroInt()),
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		s.Run(tc.name, func() {
			resp, err := testutil.GetRequestWithHeaders(tc.url, tc.headers)
			s.Require().NoError(err)

			s.Require().NoError(val.GetClientCtx().Codec.UnmarshalJSON(resp, tc.respType))
			s.Require().Equal(tc.expected.String(), tc.respType.String())
		})
	}
}

func (s *E2ETestSuite) TestDenomMetadataGRPCHandler() {
	val := s.network.GetValidators()[0]
	baseURL := val.GetAPIAddress()

	testCases := []struct {
		name     string
		url      string
		headers  map[string]string
		expErr   bool
		respType proto.Message
		expected proto.Message
	}{
		{
			"test GRPC client metadata",
			fmt.Sprintf("%s/cosmos/bank/v1beta1/denoms_metadata", baseURL),
			map[string]string{
				grpctypes.GRPCBlockHeightHeader: "1",
			},
			false,
			&types.QueryDenomsMetadataResponse{},
			&types.QueryDenomsMetadataResponse{
				Metadatas: []types.Metadata{
					{
						Name:        "Cosmos Hub Atom",
						Symbol:      "ATOM",
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
					{
						Name:        "Ethereum",
						Symbol:      "ETH",
						Description: "Ethereum mainnet token",
						DenomUnits: []*types.DenomUnit{
							{
								Denom:    "wei",
								Exponent: 0,
							},
							{
								Denom:    "eth",
								Exponent: 6,
								Aliases:  []string{"ETH"},
							},
						},
						Base:    "wei",
						Display: "eth",
					},
				},
				Pagination: &query.PageResponse{Total: 2},
			},
		},
		{
			"GRPC client metadata of a specific denom",
			fmt.Sprintf("%s/cosmos/bank/v1beta1/denoms_metadata/uatom", baseURL),
			map[string]string{
				grpctypes.GRPCBlockHeightHeader: "1",
			},
			false,
			&types.QueryDenomMetadataResponse{},
			&types.QueryDenomMetadataResponse{
				Metadata: types.Metadata{
					Name:        "Cosmos Hub Atom",
					Symbol:      "ATOM",
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
			},
		},
		{
			"GRPC client metadata of a bogus denom",
			fmt.Sprintf("%s/cosmos/bank/v1beta1/denoms_metadata/foobar", baseURL),
			map[string]string{
				grpctypes.GRPCBlockHeightHeader: "1",
			},
			true,
			&types.QueryDenomMetadataResponse{},
			&types.QueryDenomMetadataResponse{
				Metadata: types.Metadata{
					DenomUnits: []*types.DenomUnit{},
				},
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		s.Run(tc.name, func() {
			resp, err := testutil.GetRequestWithHeaders(tc.url, tc.headers)
			s.Require().NoError(err)

			if tc.expErr {
				s.Require().Error(val.GetClientCtx().Codec.UnmarshalJSON(resp, tc.respType))
			} else {
				s.Require().NoError(val.GetClientCtx().Codec.UnmarshalJSON(resp, tc.respType))
				s.Require().Equal(tc.expected.String(), tc.respType.String())
			}
		})
	}
}

func (s *E2ETestSuite) TestBalancesGRPCHandler() {
	val := s.network.GetValidators()[0]
	baseURL := val.GetAPIAddress()

	testCases := []struct {
		name     string
		url      string
		respType proto.Message
		expected proto.Message
	}{
		{
			"gRPC total account balance",
			fmt.Sprintf("%s/cosmos/bank/v1beta1/balances/%s", baseURL, val.GetAddress().String()),
			&types.QueryAllBalancesResponse{},
			&types.QueryAllBalancesResponse{
				Balances: sdk.NewCoins(
					sdk.NewCoin(fmt.Sprintf("%stoken", val.GetMoniker()), s.cfg.AccountTokens),
					sdk.NewCoin(s.cfg.BondDenom, s.cfg.StakingTokens.Sub(s.cfg.BondedTokens)),
				),
				Pagination: &query.PageResponse{
					Total: 2,
				},
			},
		},
		{
			"gPRC account balance of a denom",
			fmt.Sprintf("%s/cosmos/bank/v1beta1/balances/%s/by_denom?denom=%s", baseURL, val.GetAddress().String(), s.cfg.BondDenom),
			&types.QueryBalanceResponse{},
			&types.QueryBalanceResponse{
				Balance: &sdk.Coin{
					Denom:  s.cfg.BondDenom,
					Amount: s.cfg.StakingTokens.Sub(s.cfg.BondedTokens),
				},
			},
		},
		{
			"gPRC account balance of a bogus denom",
			fmt.Sprintf("%s/cosmos/bank/v1beta1/balances/%s/by_denom?denom=foobar", baseURL, val.GetAddress().String()),
			&types.QueryBalanceResponse{},
			&types.QueryBalanceResponse{
				Balance: &sdk.Coin{
					Denom:  "foobar",
					Amount: math.NewInt(0),
				},
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		s.Run(tc.name, func() {
			resp, err := testutil.GetRequest(tc.url)
			s.Require().NoError(err)

			s.Require().NoError(val.GetClientCtx().Codec.UnmarshalJSON(resp, tc.respType))
			s.Require().Equal(tc.expected.String(), tc.respType.String())
		})
	}
}
