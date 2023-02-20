package keeper_test

import (
	"testing"

	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	"github.com/stretchr/testify/suite"
	"pgregory.net/rapid"

	"github.com/cosmos/cosmos-sdk/baseapp"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/testutil/configurator"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/bank/keeper"
	banktestutil "github.com/cosmos/cosmos-sdk/x/bank/testutil"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"

	_ "github.com/cosmos/cosmos-sdk/x/auth"
	_ "github.com/cosmos/cosmos-sdk/x/auth/tx/config"
	_ "github.com/cosmos/cosmos-sdk/x/bank"
	_ "github.com/cosmos/cosmos-sdk/x/consensus"
	_ "github.com/cosmos/cosmos-sdk/x/params"
	_ "github.com/cosmos/cosmos-sdk/x/staking"
)

type DeterministicTestSuite struct {
	suite.Suite

	ctx        sdk.Context
	bankKeeper keeper.BaseKeeper

	queryClient banktypes.QueryClient
}

var (
	denomRegex   = sdk.DefaultCoinDenomRegex()
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
			configurator.ConsensusModule(),
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
		rapid.StringMatching(denomRegex).Draw(t, "denom"),
		sdk.NewInt(rapid.Int64Min(1).Draw(t, "amount")),
	)
}

func (suite *DeterministicTestSuite) TestGRPCQueryBalance() {
	rapid.Check(suite.T(), func(t *rapid.T) {
		addr := testdata.AddressGenerator(t).Draw(t, "address")
		coin := suite.getCoin(t)
		suite.fundAccount(addr, coin)

		req := banktypes.NewQueryBalanceRequest(addr, coin.GetDenom())
		testdata.DeterministicIterations(suite.ctx, suite.Require(), req, suite.queryClient.Balance, 0, true)
	})

	suite.fundAccount(addr1, coin1)
	req := banktypes.NewQueryBalanceRequest(addr1, coin1.GetDenom())
	testdata.DeterministicIterations(suite.ctx, suite.Require(), req, suite.queryClient.Balance, 1087, false)
}

func (suite *DeterministicTestSuite) TestGRPCQueryAllBalances() {
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

		req := banktypes.NewQueryAllBalancesRequest(addr, testdata.PaginationGenerator(t, uint64(numCoins)).Draw(t, "pagination"))
		testdata.DeterministicIterations(suite.ctx, suite.Require(), req, suite.queryClient.AllBalances, 0, true)
	})

	coins := sdk.NewCoins(
		sdk.NewCoin("stake", sdk.NewInt(10)),
		sdk.NewCoin("denom", sdk.NewInt(100)),
	)

	suite.fundAccount(addr1, coins...)
	req := banktypes.NewQueryAllBalancesRequest(addr1, nil)

	testdata.DeterministicIterations(suite.ctx, suite.Require(), req, suite.queryClient.AllBalances, 357, false)
}

func (suite *DeterministicTestSuite) TestGRPCQuerySpendableBalances() {
	rapid.Check(suite.T(), func(t *rapid.T) {
		addr := testdata.AddressGenerator(t).Draw(t, "address")
		numCoins := rapid.IntRange(1, 10).Draw(t, "num-count")
		coins := make(sdk.Coins, 0, numCoins)

		for i := 0; i < numCoins; i++ {
			coin := sdk.NewCoin(
				rapid.StringMatching(denomRegex).Draw(t, "denom"),
				sdk.NewInt(rapid.Int64Min(1).Draw(t, "amount")),
			)

			// NewCoins sorts the denoms
			coins = sdk.NewCoins(append(coins, coin)...)
		}

		err := banktestutil.FundAccount(suite.bankKeeper, suite.ctx, addr, coins)
		suite.Require().NoError(err)

		req := banktypes.NewQuerySpendableBalancesRequest(addr, testdata.PaginationGenerator(t, uint64(numCoins)).Draw(t, "pagination"))
		testdata.DeterministicIterations(suite.ctx, suite.Require(), req, suite.queryClient.SpendableBalances, 0, true)
	})

	coins := sdk.NewCoins(
		sdk.NewCoin("stake", sdk.NewInt(10)),
		sdk.NewCoin("denom", sdk.NewInt(100)),
	)

	err := banktestutil.FundAccount(suite.bankKeeper, suite.ctx, addr1, coins)
	suite.Require().NoError(err)

	req := banktypes.NewQuerySpendableBalancesRequest(addr1, nil)
	testdata.DeterministicIterations(suite.ctx, suite.Require(), req, suite.queryClient.SpendableBalances, 2032, false)
}

func (suite *DeterministicTestSuite) TestGRPCQueryTotalSupply() {
	res, err := suite.queryClient.TotalSupply(suite.ctx, &banktypes.QueryTotalSupplyRequest{})
	suite.Require().NoError(err)
	initialSupply := res.GetSupply()

	rapid.Check(suite.T(), func(t *rapid.T) {
		numCoins := rapid.IntRange(1, 3).Draw(t, "num-count")
		coins := make(sdk.Coins, 0, numCoins)

		for i := 0; i < numCoins; i++ {
			coin := sdk.NewCoin(
				rapid.StringMatching(denomRegex).Draw(t, "denom"),
				sdk.NewInt(rapid.Int64Min(1).Draw(t, "amount")),
			)

			coins = coins.Add(coin)
		}

		suite.Require().NoError(suite.bankKeeper.MintCoins(suite.ctx, minttypes.ModuleName, coins))

		initialSupply = initialSupply.Add(coins...)

		req := &banktypes.QueryTotalSupplyRequest{
			Pagination: testdata.PaginationGenerator(t, uint64(len(initialSupply))).Draw(t, "pagination"),
		}

		testdata.DeterministicIterations(suite.ctx, suite.Require(), req, suite.queryClient.TotalSupply, 0, true)
	})

	suite.SetupTest() // reset

	coins := sdk.NewCoins(
		sdk.NewCoin("foo", sdk.NewInt(10)),
		sdk.NewCoin("bar", sdk.NewInt(100)),
	)

	suite.Require().NoError(suite.bankKeeper.MintCoins(suite.ctx, minttypes.ModuleName, coins))

	req := &banktypes.QueryTotalSupplyRequest{}
	testdata.DeterministicIterations(suite.ctx, suite.Require(), req, suite.queryClient.TotalSupply, 243, false)
}

func (suite *DeterministicTestSuite) TestGRPCQueryTotalSupplyOf() {
	rapid.Check(suite.T(), func(t *rapid.T) {
		coin := sdk.NewCoin(
			rapid.StringMatching(denomRegex).Draw(t, "denom"),
			sdk.NewInt(rapid.Int64Min(1).Draw(t, "amount")),
		)

		suite.Require().NoError(suite.bankKeeper.MintCoins(suite.ctx, minttypes.ModuleName, sdk.NewCoins(coin)))

		req := &banktypes.QuerySupplyOfRequest{Denom: coin.GetDenom()}
		testdata.DeterministicIterations(suite.ctx, suite.Require(), req, suite.queryClient.SupplyOf, 0, true)
	})

	coin := sdk.NewCoin("bar", sdk.NewInt(100))

	suite.Require().NoError(suite.bankKeeper.MintCoins(suite.ctx, minttypes.ModuleName, sdk.NewCoins(coin)))
	req := &banktypes.QuerySupplyOfRequest{Denom: coin.GetDenom()}
	testdata.DeterministicIterations(suite.ctx, suite.Require(), req, suite.queryClient.SupplyOf, 1021, false)
}

func (suite *DeterministicTestSuite) TestGRPCQueryParams() {
	rapid.Check(suite.T(), func(t *rapid.T) {
		enabledStatus := banktypes.SendEnabled{
			Denom:   rapid.StringMatching(denomRegex).Draw(t, "denom"),
			Enabled: rapid.Bool().Draw(t, "status"),
		}

		params := banktypes.Params{
			SendEnabled:        []*banktypes.SendEnabled{&enabledStatus},
			DefaultSendEnabled: rapid.Bool().Draw(t, "send"),
		}

		suite.bankKeeper.SetParams(suite.ctx, params)

		req := &banktypes.QueryParamsRequest{}
		testdata.DeterministicIterations(suite.ctx, suite.Require(), req, suite.queryClient.Params, 0, true)
	})

	enabledStatus := banktypes.SendEnabled{
		Denom:   "denom",
		Enabled: true,
	}

	params := banktypes.Params{
		SendEnabled:        []*banktypes.SendEnabled{&enabledStatus},
		DefaultSendEnabled: false,
	}

	suite.bankKeeper.SetParams(suite.ctx, params)

	req := &banktypes.QueryParamsRequest{}
	testdata.DeterministicIterations(suite.ctx, suite.Require(), req, suite.queryClient.Params, 1003, false)
}

func (suite *DeterministicTestSuite) createAndReturnMetadatas(t *rapid.T, count int) []banktypes.Metadata {
	denomsMetadata := make([]banktypes.Metadata, 0, count)
	for i := 0; i < count; i++ {

		denom := rapid.StringMatching(denomRegex).Draw(t, "denom")

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

func (suite *DeterministicTestSuite) TestGRPCDenomsMetadata() {
	rapid.Check(suite.T(), func(t *rapid.T) {
		count := rapid.IntRange(1, 3).Draw(t, "count")
		denomsMetadata := suite.createAndReturnMetadatas(t, count)
		suite.Require().Len(denomsMetadata, count)

		for i := 0; i < count; i++ {
			suite.bankKeeper.SetDenomMetaData(suite.ctx, denomsMetadata[i])
		}

		req := &banktypes.QueryDenomsMetadataRequest{
			Pagination: testdata.PaginationGenerator(t, uint64(count)).Draw(t, "pagination"),
		}

		testdata.DeterministicIterations(suite.ctx, suite.Require(), req, suite.queryClient.DenomsMetadata, 0, true)
	})

	suite.SetupTest() // reset

	suite.bankKeeper.SetDenomMetaData(suite.ctx, metadataAtom)

	req := &banktypes.QueryDenomsMetadataRequest{}
	testdata.DeterministicIterations(suite.ctx, suite.Require(), req, suite.queryClient.DenomsMetadata, 660, false)
}

func (suite *DeterministicTestSuite) TestGRPCDenomMetadata() {
	rapid.Check(suite.T(), func(t *rapid.T) {
		denomMetadata := suite.createAndReturnMetadatas(t, 1)
		suite.Require().Len(denomMetadata, 1)
		suite.bankKeeper.SetDenomMetaData(suite.ctx, denomMetadata[0])

		req := &banktypes.QueryDenomMetadataRequest{
			Denom: denomMetadata[0].Base,
		}

		testdata.DeterministicIterations(suite.ctx, suite.Require(), req, suite.queryClient.DenomMetadata, 0, true)
	})

	suite.bankKeeper.SetDenomMetaData(suite.ctx, metadataAtom)

	req := &banktypes.QueryDenomMetadataRequest{
		Denom: metadataAtom.Base,
	}

	testdata.DeterministicIterations(suite.ctx, suite.Require(), req, suite.queryClient.DenomMetadata, 1300, false)
}

func (suite *DeterministicTestSuite) TestGRPCSendEnabled() {
	allDenoms := []string{}

	rapid.Check(suite.T(), func(t *rapid.T) {
		count := rapid.IntRange(0, 10).Draw(t, "count")
		denoms := make([]string, 0, count)

		for i := 0; i < count; i++ {
			coin := banktypes.SendEnabled{
				Denom:   rapid.StringMatching(denomRegex).Draw(t, "denom"),
				Enabled: rapid.Bool().Draw(t, "enabled-status"),
			}

			suite.bankKeeper.SetSendEnabled(suite.ctx, coin.Denom, coin.Enabled)
			denoms = append(denoms, coin.Denom)
		}

		allDenoms = append(allDenoms, denoms...)

		req := &banktypes.QuerySendEnabledRequest{
			Denoms: denoms,
			// Pagination is only taken into account when `denoms` is an empty array
			Pagination: testdata.PaginationGenerator(t, uint64(len(allDenoms))).Draw(t, "pagination"),
		}
		testdata.DeterministicIterations(suite.ctx, suite.Require(), req, suite.queryClient.SendEnabled, 0, true)
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

	req := &banktypes.QuerySendEnabledRequest{
		Denoms: []string{coin1.GetDenom(), coin2.GetDenom()},
	}

	testdata.DeterministicIterations(suite.ctx, suite.Require(), req, suite.queryClient.SendEnabled, 4063, false)
}

func (suite *DeterministicTestSuite) TestGRPCDenomOwners() {
	rapid.Check(suite.T(), func(t *rapid.T) {
		denom := rapid.StringMatching(denomRegex).Draw(t, "denom")
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
		testdata.DeterministicIterations(suite.ctx, suite.Require(), req, suite.queryClient.DenomOwners, 0, true)
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

	req := &banktypes.QueryDenomOwnersRequest{
		Denom: coin1.GetDenom(),
	}
	testdata.DeterministicIterations(suite.ctx, suite.Require(), req, suite.queryClient.DenomOwners, 2525, false)
}
