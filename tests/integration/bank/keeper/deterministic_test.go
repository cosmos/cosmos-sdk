package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/suite"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	"pgregory.net/rapid"

	"github.com/cosmos/cosmos-sdk/baseapp"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/testutil/configurator"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	"github.com/cosmos/cosmos-sdk/x/bank/keeper"
	banktestutil "github.com/cosmos/cosmos-sdk/x/bank/testutil"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
)

type DeterministicTestSuite struct {
	suite.Suite

	ctx        sdk.Context
	bankKeeper keeper.BaseKeeper

	queryClient banktypes.QueryClient
}

var (
	// ITER_COUNT defines the number of iterations to run on each query to test
	// determinism.
	ITER_COUNT = 1000

	addr1        = sdk.MustAccAddressFromBech32("cosmos139f7kncmglres2nf3h4hc4tade85ekfr8sulz5")
	coin1        = sdk.NewCoin("denom", sdk.NewInt(10))
	metadataAtom = banktypes.Metadata{
		Description: "The native staking token of the Cosmos Hub.",
		DenomUnits: []*banktypes.DenomUnit{
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
	}
)

func TestDeterministicTestSuite(t *testing.T) {
	suite.Run(t, new(DeterministicTestSuite))
}

func (suite *DeterministicTestSuite) SetupTest() {
	var interfaceRegistry codectypes.InterfaceRegistry

	app, err := simtestutil.Setup(
		configurator.NewAppConfig(
			configurator.AuthModule(),
			configurator.TxModule(),
			configurator.ParamsModule(),
			configurator.BankModule(),
			configurator.StakingModule(),
		),
		&suite.bankKeeper,
		&interfaceRegistry,
	)
	suite.Require().NoError(err)

	ctx := app.BaseApp.NewContext(false, tmproto.Header{})
	suite.ctx = ctx

	queryHelper := baseapp.NewQueryServerTestHelper(ctx, interfaceRegistry)
	banktypes.RegisterQueryServer(queryHelper, suite.bankKeeper)
	suite.queryClient = banktypes.NewQueryClient(queryHelper)
}

func (suite *DeterministicTestSuite) fundAccount(addr sdk.AccAddress, coin ...sdk.Coin) {
	err := banktestutil.FundAccount(suite.bankKeeper, suite.ctx, addr, sdk.NewCoins(coin...))
	suite.Require().NoError(err)
}

func (suite *DeterministicTestSuite) getCoin(t *rapid.T) sdk.Coin {
	return sdk.NewCoin(
		rapid.StringMatching(sdk.DefaultCoinDenomRegex()).Draw(t, "denom"),
		sdk.NewInt(rapid.Int64Min(1).Draw(t, "amount")),
	)
}

func (suite *DeterministicTestSuite) runQueryBalanceIterations(addr sdk.AccAddress, expected *sdk.Coin) {
	for i := 0; i < ITER_COUNT; i++ {
		res, err := suite.queryClient.Balance(suite.ctx, banktypes.NewQueryBalanceRequest(addr, expected.GetDenom()))
		suite.Require().NoError(err)
		suite.Require().NotNil(res)

		suite.Require().Equal(res.GetBalance(), expected)
	}
}

func (suite *DeterministicTestSuite) TestGRPCQueryBalance() {
	suite.SetupTest() // reset

	rapid.Check(suite.T(), func(t *rapid.T) {
		addr := testdata.AddressGenerator(t).Draw(t, "address")
		coin := suite.getCoin(t)
		suite.fundAccount(addr, coin)

		suite.runQueryBalanceIterations(addr, &coin)
	})

	suite.fundAccount(addr1, coin1)
	suite.runQueryBalanceIterations(addr1, &coin1)
}

func (suite *DeterministicTestSuite) runAllBalancesIterations(addr sdk.AccAddress, expected sdk.Coins) {
	for i := 0; i < ITER_COUNT; i++ {
		res, err := suite.queryClient.AllBalances(suite.ctx, &banktypes.QueryAllBalancesRequest{
			Address: addr.String(),
		})

		suite.Require().NoError(err)
		suite.Require().NotNil(res)
		suite.Require().NotNil(res.Balances)

		suite.Require().Equal(res.GetBalances(), expected)
	}
}

func (suite *DeterministicTestSuite) TestGRPCQueryAllBalances() {
	suite.SetupTest() // reset

	rapid.Check(suite.T(), func(t *rapid.T) {
		addr := testdata.AddressGenerator(t).Draw(t, "address")
		numCoins := rapid.IntRange(1, 10).Draw(t, "num-count")
		coins := make(sdk.Coins, 0, numCoins)

		for i := 0; i < numCoins; i++ {
			coin := suite.getCoin(t)

			// NewCoins sorts the denoms
			coins = sdk.NewCoins(append(coins, coin)...)
		}

		suite.fundAccount(addr, coins...)
		suite.runAllBalancesIterations(addr, coins)
	})

	coins := sdk.NewCoins(
		sdk.NewCoin("stake", sdk.NewInt(10)),
		sdk.NewCoin("denom", sdk.NewInt(100)),
	)

	suite.fundAccount(addr1, coins...)
	suite.runAllBalancesIterations(addr1, coins)
}

func (suite *DeterministicTestSuite) runSpendableBalancesIterations(req *banktypes.QuerySpendableBalancesRequest, expected *banktypes.QuerySpendableBalancesResponse) {
	for i := 0; i < ITER_COUNT; i++ {
		res, err := suite.queryClient.SpendableBalances(suite.ctx, req)

		suite.Require().NoError(err)
		suite.Require().NotNil(res)
		suite.Require().NotNil(res.Balances)

		suite.Require().Equal(res, expected)
	}
}

func (suite *DeterministicTestSuite) TestGRPCQuerySpendableBalances() {
	suite.SetupTest() // reset

	rapid.Check(suite.T(), func(t *rapid.T) {
		addr := testdata.AddressGenerator(t).Draw(t, "address")
		numCoins := rapid.IntRange(1, 10).Draw(t, "num-count")
		coins := make(sdk.Coins, 0, numCoins)

		for i := 0; i < numCoins; i++ {
			coin := sdk.NewCoin(
				rapid.StringMatching(sdk.DefaultCoinDenomRegex()).Draw(t, "denom"),
				sdk.NewInt(rapid.Int64Min(1).Draw(t, "amount")),
			)

			// NewCoins sorts the denoms
			coins = sdk.NewCoins(append(coins, coin)...)
		}

		err := banktestutil.FundAccount(suite.bankKeeper, suite.ctx, addr, coins)
		suite.Require().NoError(err)

		suite.runSpendableBalancesIterations(
			banktypes.NewQuerySpendableBalancesRequest(addr, testdata.PaginationGenerator(t, uint64(numCoins)).Draw(t, "pagination")),
			&banktypes.QuerySpendableBalancesResponse{Balances: coins},
		)
	})

	coins := sdk.NewCoins(
		sdk.NewCoin("stake", sdk.NewInt(10)),
		sdk.NewCoin("denom", sdk.NewInt(100)),
	)

	err := banktestutil.FundAccount(suite.bankKeeper, suite.ctx, addr1, coins)
	suite.Require().NoError(err)

	suite.runSpendableBalancesIterations(
		banktypes.NewQuerySpendableBalancesRequest(addr1, nil),
		&banktypes.QuerySpendableBalancesResponse{Balances: coins, Pagination: &query.PageResponse{Total: 2}},
	)
}

func (suite *DeterministicTestSuite) runTotalSupplyIterations(req *banktypes.QueryTotalSupplyRequest, expected *banktypes.QueryTotalSupplyResponse) {
	for i := 0; i < ITER_COUNT; i++ {
		res, err := suite.queryClient.TotalSupply(suite.ctx, req)

		suite.Require().NoError(err)
		suite.Require().NotNil(res)

		suite.Require().Equal(res, expected)
	}
}

func (suite *DeterministicTestSuite) TestGRPCQueryTotalSupply() {
	suite.SetupTest() // reset

	res, err := suite.queryClient.TotalSupply(suite.ctx, &banktypes.QueryTotalSupplyRequest{})
	suite.Require().NoError(err)
	initialSupply := res.GetSupply()

	rapid.Check(suite.T(), func(t *rapid.T) {
		numCoins := rapid.IntRange(1, 10).Draw(t, "num-count")
		coins := make(sdk.Coins, 0, numCoins)

		for i := 0; i < numCoins; i++ {
			coin := sdk.NewCoin(
				rapid.StringMatching(sdk.DefaultCoinDenomRegex()).Draw(t, "denom"),
				sdk.NewInt(rapid.Int64Min(1).Draw(t, "amount")),
			)

			coins = coins.Add(coin)
		}

		suite.Require().NoError(suite.bankKeeper.MintCoins(suite.ctx, minttypes.ModuleName, coins))

		initialSupply = initialSupply.Add(coins...)

		req := &banktypes.QueryTotalSupplyRequest{
			Pagination: testdata.PaginationGenerator(t, uint64(numCoins)).Draw(t, "pagination"),
		}
		res, err = suite.queryClient.TotalSupply(suite.ctx, req)
		suite.Require().NoError(err)

		suite.runTotalSupplyIterations(req, res)
	})

	res, err = suite.queryClient.TotalSupply(suite.ctx, &banktypes.QueryTotalSupplyRequest{})
	suite.Require().NoError(err)
	suite.Require().NotNil(res)

	coins := sdk.NewCoins(
		sdk.NewCoin("foo", sdk.NewInt(10)),
		sdk.NewCoin("bar", sdk.NewInt(100)),
	)

	suite.Require().NoError(suite.bankKeeper.MintCoins(suite.ctx, minttypes.ModuleName, coins))

	coins = initialSupply.Add(coins...)
	suite.runTotalSupplyIterations(
		&banktypes.QueryTotalSupplyRequest{},
		&banktypes.QueryTotalSupplyResponse{Supply: coins, Pagination: &query.PageResponse{Total: uint64(len(coins))}},
	)
}

func (suite *DeterministicTestSuite) runTotalSupplyOfIterations(denom string, expected sdk.Coin) {
	for i := 0; i < ITER_COUNT; i++ {
		res, err := suite.queryClient.SupplyOf(suite.ctx, &banktypes.QuerySupplyOfRequest{
			Denom: denom,
		})

		suite.Require().NoError(err)
		suite.Require().NotNil(res)

		suite.Require().Equal(res.GetAmount(), expected)
	}
}

func (suite *DeterministicTestSuite) TestGRPCQueryTotalSupplyOf() {
	suite.SetupTest() // reset

	rapid.Check(suite.T(), func(t *rapid.T) {
		coin := sdk.NewCoin(
			rapid.StringMatching(sdk.DefaultCoinDenomRegex()).Draw(t, "denom"),
			sdk.NewInt(rapid.Int64Min(1).Draw(t, "amount")),
		)

		suite.Require().NoError(suite.bankKeeper.MintCoins(suite.ctx, minttypes.ModuleName, sdk.NewCoins(coin)))

		suite.runTotalSupplyOfIterations(coin.Denom, coin)
	})

	coin := sdk.NewCoin("bar", sdk.NewInt(100))

	suite.Require().NoError(suite.bankKeeper.MintCoins(suite.ctx, minttypes.ModuleName, sdk.NewCoins(coin)))

	suite.runTotalSupplyOfIterations(coin.Denom, coin)
}

func (suite *DeterministicTestSuite) runParamsIterations(expected banktypes.Params) {
	for i := 0; i < ITER_COUNT; i++ {
		res, err := suite.queryClient.Params(suite.ctx, &banktypes.QueryParamsRequest{})

		suite.Require().NoError(err)
		suite.Require().NotNil(res)

		suite.Require().Equal(res.GetParams(), expected)
	}
}

func (suite *DeterministicTestSuite) TestGRPCQueryParams() {
	suite.SetupTest() // reset

	rapid.Check(suite.T(), func(t *rapid.T) {
		enabledStatus := banktypes.SendEnabled{
			Denom:   rapid.StringMatching(sdk.DefaultCoinDenomRegex()).Draw(t, "denom"),
			Enabled: rapid.Bool().Draw(t, "status"),
		}

		params := banktypes.Params{
			SendEnabled:        []*banktypes.SendEnabled{&enabledStatus},
			DefaultSendEnabled: rapid.Bool().Draw(t, "send"),
		}

		// SetParams overwrites `SendEnabled` to nil
		suite.bankKeeper.SetParams(suite.ctx, params)
		params.SendEnabled = nil
		suite.runParamsIterations(params)
	})

	enabledStatus := banktypes.SendEnabled{
		Denom:   "denom",
		Enabled: true,
	}

	params := banktypes.Params{
		SendEnabled:        []*banktypes.SendEnabled{&enabledStatus},
		DefaultSendEnabled: false,
	}

	// SetParams overwrites `SendEnabled` to nil
	suite.bankKeeper.SetParams(suite.ctx, params)
	params.SendEnabled = nil
	suite.runParamsIterations(params)
}

func (suite *DeterministicTestSuite) createAndReturnMetadatas(t *rapid.T, count int) []banktypes.Metadata {
	denomsMetadata := make([]banktypes.Metadata, 0, count)
	for i := 0; i < count; i++ {

		denom := rapid.StringMatching(sdk.DefaultCoinDenomRegex()).Draw(t, "denom")

		aliases := rapid.SliceOf(rapid.String()).Draw(t, "aliases")
		// In the GRPC server code, empty arrays are returned as nil
		if len(aliases) == 0 {
			aliases = nil
		}

		metadata := banktypes.Metadata{
			Description: rapid.StringN(1, 100, 100).Draw(t, "desc"),
			DenomUnits: []*banktypes.DenomUnit{
				{
					Denom:    denom,
					Exponent: rapid.Uint32().Draw(t, "exponent"),
					Aliases:  aliases,
				},
			},
			Base:    denom,
			Display: denom,
			Name:    rapid.String().Draw(t, "name"),
			Symbol:  rapid.String().Draw(t, "symbol"),
			URI:     rapid.String().Draw(t, "uri"),
			URIHash: rapid.String().Draw(t, "uri-hash"),
		}

		denomsMetadata = append(denomsMetadata, metadata)
	}

	return denomsMetadata
}

func (suite *DeterministicTestSuite) runDenomsMetadataIterations(req *banktypes.QueryDenomsMetadataRequest, expected *banktypes.QueryDenomsMetadataResponse) {
	for i := 0; i < ITER_COUNT; i++ {
		res, err := suite.queryClient.DenomsMetadata(suite.ctx, req)
		suite.Require().NoError(err)
		suite.Require().NotNil(res)

		suite.Require().Equal(res, expected)
	}
}

func (suite *DeterministicTestSuite) TestGRPCDenomsMetadata() {
	suite.SetupTest() // reset

	rapid.Check(suite.T(), func(t *rapid.T) {
		count := rapid.IntRange(1, 5).Draw(t, "count")
		denomsMetadata := suite.createAndReturnMetadatas(t, count)
		suite.Require().Len(denomsMetadata, count)

		for i := 0; i < count; i++ {
			suite.bankKeeper.SetDenomMetaData(suite.ctx, denomsMetadata[i])
		}

		req := &banktypes.QueryDenomsMetadataRequest{
			Pagination: testdata.PaginationGenerator(t, uint64(count)).Draw(t, "pagination"),
		}
		res, err := suite.queryClient.DenomsMetadata(suite.ctx, req)
		suite.Require().NoError(err)
		suite.Require().NotNil(res)

		suite.runDenomsMetadataIterations(req, res)
	})

	suite.SetupTest() // reset

	suite.bankKeeper.SetDenomMetaData(suite.ctx, metadataAtom)
	suite.runDenomsMetadataIterations(
		&banktypes.QueryDenomsMetadataRequest{},
		&banktypes.QueryDenomsMetadataResponse{Metadatas: []banktypes.Metadata{metadataAtom}, Pagination: &query.PageResponse{Total: 1}},
	)
}

func (suite *DeterministicTestSuite) runDenomMetadataIterations(denom string, expected banktypes.Metadata) {
	for i := 0; i < ITER_COUNT; i++ {
		res, err := suite.queryClient.DenomMetadata(suite.ctx, &banktypes.QueryDenomMetadataRequest{
			Denom: denom,
		})

		suite.Require().NoError(err)
		suite.Require().NotNil(res)

		suite.Require().Equal(res.GetMetadata(), expected)
	}
}

func (suite *DeterministicTestSuite) TestGRPCDenomMetadata() {
	suite.SetupTest() // reset

	rapid.Check(suite.T(), func(t *rapid.T) {
		denomMetadata := suite.createAndReturnMetadatas(t, 1)
		suite.Require().Len(denomMetadata, 1)
		suite.bankKeeper.SetDenomMetaData(suite.ctx, denomMetadata[0])
		suite.runDenomMetadataIterations(denomMetadata[0].Base, denomMetadata[0])
	})

	suite.bankKeeper.SetDenomMetaData(suite.ctx, metadataAtom)
	suite.runDenomMetadataIterations(metadataAtom.Base, metadataAtom)
}

func (suite *DeterministicTestSuite) runSendEnabledIterations(req *banktypes.QuerySendEnabledRequest, expected *banktypes.QuerySendEnabledResponse) {
	for i := 0; i < ITER_COUNT; i++ {
		res, err := suite.queryClient.SendEnabled(suite.ctx, req)
		suite.Require().NoError(err)
		suite.Require().NotNil(res)

		suite.Require().Equal(expected, res)
	}
}

func (suite *DeterministicTestSuite) TestGRPCSendEnabled() {
	suite.SetupTest() // reset

	rapid.Check(suite.T(), func(t *rapid.T) {
		count := rapid.IntRange(1, 10).Draw(t, "count")
		denoms := make([]string, 0, count)

		for i := 0; i < count; i++ {
			coin := banktypes.SendEnabled{
				Denom:   rapid.StringMatching(sdk.DefaultCoinDenomRegex()).Draw(t, "denom"),
				Enabled: rapid.Bool().Draw(t, "enabled-status"),
			}

			suite.bankKeeper.SetSendEnabled(suite.ctx, coin.Denom, coin.Enabled)
			denoms = append(denoms, coin.Denom)
		}

		req := &banktypes.QuerySendEnabledRequest{
			Denoms:     denoms,
			Pagination: testdata.PaginationGenerator(t, uint64(count)).Draw(t, "pagination"),
		}
		res, err := suite.queryClient.SendEnabled(suite.ctx, req)
		suite.Require().NoError(err)

		suite.runSendEnabledIterations(req, res)
	})

	coin1 := banktypes.SendEnabled{
		Denom:   "falsecoin",
		Enabled: false,
	}
	coin2 := banktypes.SendEnabled{
		Denom:   "truecoin",
		Enabled: true,
	}

	suite.bankKeeper.SetSendEnabled(suite.ctx, coin1.Denom, false)
	suite.bankKeeper.SetSendEnabled(suite.ctx, coin2.Denom, true)

	suite.runSendEnabledIterations(
		&banktypes.QuerySendEnabledRequest{Denoms: []string{coin1.Denom, coin2.Denom}},
		&banktypes.QuerySendEnabledResponse{SendEnabled: []*banktypes.SendEnabled{&coin1, &coin2}},
	)
}

func (suite *DeterministicTestSuite) runDenomOwnerIterations(req *banktypes.QueryDenomOwnersRequest, expected *banktypes.QueryDenomOwnersResponse) {
	for i := 0; i < ITER_COUNT; i++ {
		res, err := suite.queryClient.DenomOwners(suite.ctx, req)

		suite.Require().NoError(err)
		suite.Require().Equal(expected, res)
	}
}

func (suite *DeterministicTestSuite) TestGRPCDenomOwners() {
	suite.SetupTest() // reset

	rapid.Check(suite.T(), func(t *rapid.T) {
		denom := rapid.StringMatching(sdk.DefaultCoinDenomRegex()).Draw(t, "denom")
		numAddr := rapid.IntRange(1, 10).Draw(t, "number-address")
		for i := 0; i < numAddr; i++ {
			addr := testdata.AddressGenerator(t).Draw(t, "address")

			coin := sdk.NewCoin(
				denom,
				sdk.NewInt(rapid.Int64Min(1).Draw(t, "amount")),
			)

			err := banktestutil.FundAccount(suite.bankKeeper, suite.ctx, addr, sdk.NewCoins(coin))
			suite.Require().NoError(err)
		}

		req := &banktypes.QueryDenomOwnersRequest{
			Denom:      denom,
			Pagination: testdata.PaginationGenerator(t, uint64(numAddr)).Draw(t, "pagination"),
		}
		res, err := suite.queryClient.DenomOwners(suite.ctx, req)
		suite.Require().NoError(err)

		suite.runDenomOwnerIterations(req, res)
	})

	denomOwners := []*banktypes.DenomOwner{
		{
			Address: "cosmos1qg65a9q6k2sqq7l3ycp428sqqpmqcucgzze299",
			Balance: coin1,
		},
		{
			Address: "cosmos1qglnsqgpq48l7qqzgs8qdshr6fh3gqq9ej3qut",
			Balance: coin1,
		},
	}

	for i := 0; i < len(denomOwners); i++ {
		addr, err := sdk.AccAddressFromBech32(denomOwners[i].Address)
		suite.Require().NoError(err)

		err = banktestutil.FundAccount(suite.bankKeeper, suite.ctx, addr, sdk.NewCoins(coin1))
		suite.Require().NoError(err)
	}

	suite.runDenomOwnerIterations(
		&banktypes.QueryDenomOwnersRequest{Denom: coin1.Denom},
		&banktypes.QueryDenomOwnersResponse{DenomOwners: denomOwners, Pagination: &query.PageResponse{Total: 2}},
	)
}
