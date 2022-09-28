package keeper_test

import (
	"testing"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/testutil"
	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/bank"
	"github.com/cosmos/cosmos-sdk/x/bank/keeper"
	banktestutil "github.com/cosmos/cosmos-sdk/x/bank/testutil"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/suite"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	tmtime "github.com/tendermint/tendermint/types/time"
	"pgregory.net/rapid"
)

type DeterministicTestSuite struct {
	suite.Suite

	ctx        sdk.Context
	bankKeeper keeper.BaseKeeper
	authKeeper *banktestutil.MockAccountKeeper

	queryClient banktypes.QueryClient
	encCfg      moduletestutil.TestEncodingConfig
	mintAcc     *authtypes.ModuleAccount
}

const (
	denomRegex = `[a-zA-Z][a-zA-Z0-9/:._-]{2,127}`
	name       = `[a-zA-Z][a-zA-Z0-9]{2,127}`
)

func TestDeterministicTestSuite(t *testing.T) {
	suite.Run(t, new(DeterministicTestSuite))
}

func (suite *DeterministicTestSuite) SetupTest() {
	key := sdk.NewKVStoreKey(banktypes.StoreKey)
	testCtx := testutil.DefaultContextWithDB(suite.T(), key, sdk.NewTransientStoreKey("transient_test"))
	ctx := testCtx.Ctx.WithBlockHeader(tmproto.Header{Time: tmtime.Now()})
	encCfg := moduletestutil.MakeTestEncodingConfig(bank.AppModuleBasic{})

	// gomock initializations
	ctrl := gomock.NewController(suite.T())
	authKeeper := banktestutil.NewMockAccountKeeper(ctrl)

	suite.ctx = ctx
	suite.authKeeper = authKeeper
	suite.bankKeeper = keeper.NewBaseKeeper(
		encCfg.Codec,
		key,
		suite.authKeeper,
		map[string]bool{accAddrs[4].String(): true},
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)

	queryHelper := baseapp.NewQueryServerTestHelper(ctx, encCfg.InterfaceRegistry)
	banktypes.RegisterQueryServer(queryHelper, suite.bankKeeper)
	queryClient := banktypes.NewQueryClient(queryHelper)

	suite.queryClient = queryClient
	suite.encCfg = encCfg

	suite.mintAcc = authtypes.NewEmptyModuleAccount(minttypes.ModuleName, authtypes.Minter)
}

func (suite *DeterministicTestSuite) mockMintCoins(moduleAcc *authtypes.ModuleAccount) {
	suite.authKeeper.EXPECT().GetModuleAccount(suite.ctx, moduleAcc.Name).Return(moduleAcc)
}

func (suite *DeterministicTestSuite) mockSendCoinsFromModuleToAccount(moduleAcc *authtypes.ModuleAccount, accAddr sdk.AccAddress) {
	suite.authKeeper.EXPECT().GetModuleAddress(moduleAcc.Name).Return(moduleAcc.GetAddress())
	suite.authKeeper.EXPECT().GetAccount(suite.ctx, moduleAcc.GetAddress()).Return(moduleAcc)
	suite.authKeeper.EXPECT().HasAccount(suite.ctx, accAddr).Return(true)
}

func (suite *DeterministicTestSuite) mockFundAccount(receiver sdk.AccAddress) {
	suite.mockMintCoins(suite.mintAcc)
	suite.mockSendCoinsFromModuleToAccount(mintAcc, receiver)
}

func (suite *DeterministicTestSuite) fundAccount(addr sdk.AccAddress, coin ...sdk.Coin) {
	suite.mockFundAccount(addr)
	err := banktestutil.FundAccount(suite.bankKeeper, suite.ctx, addr, sdk.NewCoins(coin...))
	suite.Require().NoError(err)
}

func (suite *DeterministicTestSuite) getCoin(t *rapid.T) sdk.Coin {
	return sdk.NewCoin(
		rapid.StringMatching(denomRegex).Draw(t, "denom"),
		sdk.NewInt(rapid.Int64Min(1).Draw(t, "amount")),
	)
}

func (suite *DeterministicTestSuite) runQueryBalanceIterations(addr sdk.AccAddress, prevRes *sdk.Coin) {
	for i := 0; i < 1000; i++ {
		res, err := suite.queryClient.Balance(suite.ctx, banktypes.NewQueryBalanceRequest(addr, prevRes.GetDenom()))
		suite.Require().NoError(err)
		suite.Require().NotNil(res)

		suite.Require().Equal(res.GetBalance(), prevRes)
		prevRes = res.Balance
	}
}

func (suite *DeterministicTestSuite) TestGRPCQueryBalance() {
	rapid.Check(suite.T(), func(t *rapid.T) {
		addr := testdata.AddressGenerator(t).Draw(t, "address")
		coin := suite.getCoin(t)
		suite.fundAccount(addr, coin)

		suite.runQueryBalanceIterations(addr, &coin)
	})

	addr, err := sdk.AccAddressFromBech32("cosmos139f7kncmglres2nf3h4hc4tade85ekfr8sulz5")
	suite.Require().NoError(err)

	coin := sdk.NewCoin(
		"denom",
		sdk.NewInt(10),
	)

	suite.fundAccount(addr, coin)
	suite.runQueryBalanceIterations(addr, &coin)
}

func (suite *DeterministicTestSuite) runAllBalancesIterations(addr sdk.AccAddress, prevRes sdk.Coins) {
	for i := 0; i < 1000; i++ {
		res, err := suite.queryClient.AllBalances(suite.ctx, &banktypes.QueryAllBalancesRequest{
			Address: addr.String(),
		})

		suite.Require().NoError(err)
		suite.Require().NotNil(res)
		suite.Require().NotNil(res.Balances)

		suite.Require().Equal(res.GetBalances(), prevRes)
		prevRes = res.GetBalances()
	}
}

func (suite *DeterministicTestSuite) TestGRPCQueryAllBalances() {
	rapid.Check(suite.T(), func(t *rapid.T) {
		suite.SetupTest() // reset
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

	suite.SetupTest() // reset
	addr, err := sdk.AccAddressFromBech32("cosmos139f7kncmglres2nf3h4hc4tade85ekfr8sulz5")
	suite.Require().NoError(err)

	coins := sdk.NewCoins(
		sdk.NewCoin("stake", sdk.NewInt(10)),
		sdk.NewCoin("denom", sdk.NewInt(100)),
	)

	suite.fundAccount(addr, coins...)
	suite.runAllBalancesIterations(addr, coins)
}

func (suite *DeterministicTestSuite) runSpendableBalancesIterations(addr sdk.AccAddress, prevRes sdk.Coins) {
	suite.authKeeper.EXPECT().GetAccount(suite.ctx, addr).Return(authtypes.NewBaseAccount(addr, nil, 10087, 0)).Times(1000)
	for i := 0; i < 1000; i++ {
		res, err := suite.queryClient.SpendableBalances(suite.ctx, &banktypes.QuerySpendableBalancesRequest{
			Address: addr.String(),
		})

		suite.Require().NoError(err)
		suite.Require().NotNil(res)
		suite.Require().NotNil(res.Balances)

		suite.Require().Equal(res.GetBalances(), prevRes)
		prevRes = res.GetBalances()
	}
}

func (suite *DeterministicTestSuite) TestGRPCQuerySpendableBalances() {
	rapid.Check(suite.T(), func(t *rapid.T) {
		suite.SetupTest() // reset
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

		suite.mockFundAccount(addr)
		err := banktestutil.FundAccount(suite.bankKeeper, suite.ctx, addr, coins)
		suite.Require().NoError(err)

		suite.runSpendableBalancesIterations(addr, coins)
	})

	suite.SetupTest() // reset
	addr, err := sdk.AccAddressFromBech32("cosmos139f7kncmglres2nf3h4hc4tade85ekfr8sulz5")
	suite.Require().NoError(err)

	coins := sdk.NewCoins(
		sdk.NewCoin("stake", sdk.NewInt(10)),
		sdk.NewCoin("denom", sdk.NewInt(100)),
	)

	suite.mockFundAccount(addr)
	err = banktestutil.FundAccount(suite.bankKeeper, suite.ctx, addr, coins)
	suite.Require().NoError(err)

	suite.runSpendableBalancesIterations(addr, coins)
}

func (suite *DeterministicTestSuite) runTotalSupplyIterations(prevRes sdk.Coins) {
	for i := 0; i < 1000; i++ {
		res, err := suite.queryClient.TotalSupply(suite.ctx, &banktypes.QueryTotalSupplyRequest{})

		suite.Require().NoError(err)
		suite.Require().NotNil(res)

		suite.Require().Equal(res.GetSupply(), prevRes)
		prevRes = res.GetSupply()
	}
}

func (suite *DeterministicTestSuite) TestGRPCQueryTotalSupply() {
	rapid.Check(suite.T(), func(t *rapid.T) {
		suite.SetupTest() // reset

		res, err := suite.queryClient.TotalSupply(suite.ctx, &banktypes.QueryTotalSupplyRequest{})
		suite.Require().NoError(err)
		suite.Require().NotNil(res)
		genesisSupply := res.GetSupply()

		numCoins := rapid.IntRange(1, 2).Draw(t, "num-count")
		coins := make(sdk.Coins, 0, numCoins)

		for i := 0; i < numCoins; i++ {
			coin := sdk.NewCoin(
				rapid.StringMatching(denomRegex).Draw(t, "denom"),
				sdk.NewInt(rapid.Int64Min(1).Draw(t, "amount")),
			)

			coins = coins.Add(coin)
		}

		suite.mockMintCoins(suite.mintAcc)
		suite.Require().NoError(suite.bankKeeper.MintCoins(suite.ctx, minttypes.ModuleName, coins))

		coins = genesisSupply.Add(coins...)
		suite.runTotalSupplyIterations(coins)
	})

	suite.SetupTest() // reset
	res, err := suite.queryClient.TotalSupply(suite.ctx, &banktypes.QueryTotalSupplyRequest{})
	suite.Require().NoError(err)
	suite.Require().NotNil(res)
	genesisSupply := res.GetSupply()

	coins := sdk.NewCoins(
		sdk.NewCoin("foo", sdk.NewInt(10)),
		sdk.NewCoin("bar", sdk.NewInt(100)),
	)

	suite.mockMintCoins(suite.mintAcc)
	suite.Require().NoError(suite.bankKeeper.MintCoins(suite.ctx, minttypes.ModuleName, coins))

	coins = genesisSupply.Add(coins...)
	suite.runTotalSupplyIterations(coins)
}

func (suite *DeterministicTestSuite) runTotalSupplyOfIterations(denom string, prevRes sdk.Coin) {
	for i := 0; i < 1000; i++ {
		res, err := suite.queryClient.SupplyOf(suite.ctx, &banktypes.QuerySupplyOfRequest{
			Denom: denom,
		})

		suite.Require().NoError(err)
		suite.Require().NotNil(res)

		suite.Require().Equal(res.GetAmount(), prevRes)
		prevRes = res.GetAmount()
	}
}

func (suite *DeterministicTestSuite) TestGRPCQueryTotalSupplyOf() {
	rapid.Check(suite.T(), func(t *rapid.T) {
		coin := sdk.NewCoin(
			rapid.StringMatching(denomRegex).Draw(t, "denom"),
			sdk.NewInt(rapid.Int64Min(1).Draw(t, "amount")),
		)

		suite.mockMintCoins(suite.mintAcc)
		suite.Require().NoError(suite.bankKeeper.MintCoins(suite.ctx, minttypes.ModuleName, sdk.NewCoins(coin)))

		suite.runTotalSupplyOfIterations(coin.Denom, coin)
	})

	coin := sdk.NewCoin("bar", sdk.NewInt(100))

	suite.mockMintCoins(suite.mintAcc)
	suite.Require().NoError(suite.bankKeeper.MintCoins(suite.ctx, minttypes.ModuleName, sdk.NewCoins(coin)))

	suite.runTotalSupplyOfIterations(coin.Denom, coin)
}

func (suite *DeterministicTestSuite) runParamsIterations(prevRes banktypes.Params) {
	for i := 0; i < 1000; i++ {
		res, err := suite.queryClient.Params(suite.ctx, &banktypes.QueryParamsRequest{})

		suite.Require().NoError(err)
		suite.Require().NotNil(res)

		suite.Require().Equal(res.GetParams(), prevRes)
		prevRes = res.GetParams()
	}
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

		denom := rapid.StringMatching(denomRegex).Draw(t, "denom")

		metadata := banktypes.Metadata{
			Description: rapid.StringN(1, 100, 100).Draw(t, "desc"),
			DenomUnits: []*banktypes.DenomUnit{
				{
					Denom:    denom,
					Exponent: rapid.Uint32().Draw(t, "exponent"),
					Aliases:  []string{denom},
				},
			},
			Base:    denom,
			Display: denom,
		}

		denomsMetadata = append(denomsMetadata, metadata)
	}

	return denomsMetadata
}

func (suite *DeterministicTestSuite) runDenomsMetadataIterations(prevRes []banktypes.Metadata) {
	for i := 0; i < 1000; i++ {
		res, err := suite.queryClient.DenomsMetadata(suite.ctx, &banktypes.QueryDenomsMetadataRequest{})
		suite.Require().NoError(err)
		suite.Require().NotNil(res)

		for i := 0; i < len(res.GetMetadatas()); i++ {
			suite.Require().Equal(res.GetMetadatas()[i], prevRes[i])
		}

		prevRes = res.GetMetadatas()
	}
}

func (suite *DeterministicTestSuite) TestGRPCDenomsMetadata() {
	rapid.Check(suite.T(), func(t *rapid.T) {
		suite.SetupTest() // reset
		count := rapid.IntRange(1, 5).Draw(t, "count")
		denomsMetadata := suite.createAndReturnMetadatas(t, count)
		suite.Require().Len(denomsMetadata, count)

		for i := 0; i < count; i++ {
			suite.bankKeeper.SetDenomMetaData(suite.ctx, denomsMetadata[i])
		}

		res, err := suite.queryClient.DenomsMetadata(suite.ctx, &banktypes.QueryDenomsMetadataRequest{})
		suite.Require().NoError(err)
		suite.Require().NotNil(res)

		suite.runDenomsMetadataIterations(res.Metadatas)
	})

	suite.SetupTest() // reset

	metadataAtom := banktypes.Metadata{
		Description: "The native staking token of the Cosmos Hub.",
		DenomUnits: []*banktypes.DenomUnit{
			{
				Denom:    "utest",
				Exponent: 0,
				Aliases:  []string{"microtest"},
			},
			{
				Denom:    "test",
				Exponent: 6,
				Aliases:  []string{"TEST"},
			},
		},
		Base:    "utest",
		Display: "test",
	}

	suite.bankKeeper.SetDenomMetaData(suite.ctx, metadataAtom)
	suite.runDenomsMetadataIterations([]banktypes.Metadata{metadataAtom})
}

func (suite *DeterministicTestSuite) runDenomMetadataIterations(denom string, prevRes banktypes.Metadata) {
	for i := 0; i < 1000; i++ {
		res, err := suite.queryClient.DenomMetadata(suite.ctx, &banktypes.QueryDenomMetadataRequest{
			Denom: denom,
		})

		suite.Require().NoError(err)
		suite.Require().NotNil(res)

		suite.Require().Equal(res.GetMetadata(), prevRes)
		prevRes = res.GetMetadata()
	}
}

func (suite *DeterministicTestSuite) TestGRPCDenomMetadata() {
	rapid.Check(suite.T(), func(t *rapid.T) {
		denomMetadata := suite.createAndReturnMetadatas(t, 1)
		suite.Require().Len(denomMetadata, 1)
		suite.bankKeeper.SetDenomMetaData(suite.ctx, denomMetadata[0])
		suite.runDenomMetadataIterations(denomMetadata[0].Base, denomMetadata[0])
	})

	metadataAtom := banktypes.Metadata{
		Description: "The native staking token of the Cosmos Hub.",
		DenomUnits: []*banktypes.DenomUnit{
			{
				Denom:    "utest",
				Exponent: 0,
				Aliases:  []string{"microtest"},
			},
			{
				Denom:    "test",
				Exponent: 6,
				Aliases:  []string{"TEST"},
			},
		},
		Base:    "utest",
		Display: "test",
	}

	suite.bankKeeper.SetDenomMetaData(suite.ctx, metadataAtom)
	suite.runDenomMetadataIterations(metadataAtom.Base, metadataAtom)
}

func (suite *DeterministicTestSuite) runSendEnabledIterations(denoms []string, prevRes []*banktypes.SendEnabled) {
	for i := 0; i < 1000; i++ {
		res, err := suite.queryClient.SendEnabled(suite.ctx, &banktypes.QuerySendEnabledRequest{
			Denoms: denoms,
		})
		suite.Require().NoError(err)
		suite.Require().NotNil(res)

		suite.Require().Equal(res.SendEnabled, prevRes)
		prevRes = res.SendEnabled
	}
}

func (suite *DeterministicTestSuite) TestGRPCSendEnabled() {
	rapid.Check(suite.T(), func(t *rapid.T) {
		count := rapid.IntRange(1, 10).Draw(t, "count")
		sendEnabled := make([]*banktypes.SendEnabled, 0, count)
		denoms := make([]string, 0, count)

		for i := 0; i < count; i++ {
			coin := banktypes.SendEnabled{
				Denom:   rapid.StringMatching(denomRegex).Draw(t, "denom"),
				Enabled: rapid.Bool().Draw(t, "enabled-status"),
			}

			suite.bankKeeper.SetSendEnabled(suite.ctx, coin.Denom, coin.Enabled)
			sendEnabled = append(sendEnabled, &coin)
			denoms = append(denoms, coin.Denom)
		}

		suite.runSendEnabledIterations(denoms, sendEnabled)
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
		[]string{coin1.Denom, coin2.Denom},
		[]*banktypes.SendEnabled{
			&coin1,
			&coin2,
		},
	)
}

func (suite *DeterministicTestSuite) runDenomOwnerIterations(denom string, prevRes []*banktypes.DenomOwner) {
	for i := 0; i < 1000; i++ {
		res, err := suite.queryClient.DenomOwners(suite.ctx, &banktypes.QueryDenomOwnersRequest{
			Denom: denom,
		})

		suite.Require().NoError(err)
		suite.Require().Equal(res.DenomOwners, prevRes)
		prevRes = res.DenomOwners
	}
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

			suite.mockFundAccount(addr)

			err := banktestutil.FundAccount(suite.bankKeeper, suite.ctx, addr, sdk.NewCoins(coin))
			suite.Require().NoError(err)
		}

		res, err := suite.queryClient.DenomOwners(suite.ctx, &banktypes.QueryDenomOwnersRequest{
			Denom: denom,
		})

		suite.Require().NoError(err)
		suite.runDenomOwnerIterations(denom, res.DenomOwners)
	})

	coin := sdk.NewCoin(
		"denom",
		sdk.NewInt(10),
	)

	denomOwners := []*banktypes.DenomOwner{
		{
			Address: "cosmos1qg65a9q6k2sqq7l3ycp428sqqpmqcucgzze299",
			Balance: coin,
		},
		{
			Address: "cosmos1qglnsqgpq48l7qqzgs8qdshr6fh3gqq9ej3qut",
			Balance: coin,
		},
	}

	for i := 0; i < len(denomOwners); i++ {
		addr, err := sdk.AccAddressFromBech32(denomOwners[i].Address)
		suite.Require().NoError(err)

		suite.mockFundAccount(addr)
		err = banktestutil.FundAccount(suite.bankKeeper, suite.ctx, addr, sdk.NewCoins(coin))
		suite.Require().NoError(err)
	}

	suite.runDenomOwnerIterations(coin.Denom, denomOwners)
}
