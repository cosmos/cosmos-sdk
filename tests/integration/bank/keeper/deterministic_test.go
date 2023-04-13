package keeper_test

import (
	"testing"

	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	"gotest.tools/v3/assert"
	"pgregory.net/rapid"

	"github.com/cosmos/cosmos-sdk/baseapp"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/testutil/configurator"
	simstestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
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

type deterministicFixture struct {
	ctx        sdk.Context
	bankKeeper keeper.BaseKeeper

	queryClient banktypes.QueryClient
}

func initDeterministicFixture(t *testing.T) *deterministicFixture {
	f := &deterministicFixture{}

	var interfaceRegistry codectypes.InterfaceRegistry

	app, err := simstestutil.Setup(
		configurator.NewAppConfig(
			configurator.AuthModule(),
			configurator.TxModule(),
			configurator.ParamsModule(),
			configurator.ConsensusModule(),
			configurator.BankModule(),
			configurator.StakingModule(),
		),
		&f.bankKeeper,
		&interfaceRegistry,
	)
	assert.NilError(t, err)

	ctx := app.BaseApp.NewContext(false, cmtproto.Header{})
	f.ctx = ctx

	queryHelper := baseapp.NewQueryServerTestHelper(ctx, interfaceRegistry)
	banktypes.RegisterQueryServer(queryHelper, f.bankKeeper)
	f.queryClient = banktypes.NewQueryClient(queryHelper)

	return f
}

func fundAccount(f *deterministicFixture, addr sdk.AccAddress, coin ...sdk.Coin) {
	err := banktestutil.FundAccount(f.bankKeeper, f.ctx, addr, sdk.NewCoins(coin...))
	assert.NilError(&testing.T{}, err)
}

func getCoin(rt *rapid.T) sdk.Coin {
	return sdk.NewCoin(
		rapid.StringMatching(denomRegex).Draw(rt, "denom"),
		sdk.NewInt(rapid.Int64Min(1).Draw(rt, "amount")),
	)
}

func TestGRPCQueryBalance(t *testing.T) {
	t.Parallel()
	f := initDeterministicFixture(t)

	rapid.Check(t, func(rt *rapid.T) {
		addr := testdata.AddressGenerator(rt).Draw(rt, "address")
		coin := getCoin(rt)
		fundAccount(f, addr, coin)

		req := banktypes.NewQueryBalanceRequest(addr, coin.GetDenom())

		testdata.DeterministicIterations(f.ctx, t, req, f.queryClient.Balance, 0, true)
	})

	fundAccount(f, addr1, coin1)
	req := banktypes.NewQueryBalanceRequest(addr1, coin1.GetDenom())
	testdata.DeterministicIterations(f.ctx, t, req, f.queryClient.Balance, 1087, false)
}

func TestGRPCQueryAllBalances(t *testing.T) {
	t.Parallel()
	f := initDeterministicFixture(t)

	rapid.Check(t, func(rt *rapid.T) {
		addr := testdata.AddressGenerator(rt).Draw(rt, "address")
		numCoins := rapid.IntRange(1, 10).Draw(rt, "num-count")
		coins := make(sdk.Coins, 0, numCoins)

		for i := 0; i < numCoins; i++ {
			coin := getCoin(rt)

			// NewCoins sorts the denoms
			coins = sdk.NewCoins(append(coins, coin)...)
		}

		fundAccount(f, addr, coins...)

		req := banktypes.NewQueryAllBalancesRequest(addr, testdata.PaginationGenerator(rt, uint64(numCoins)).Draw(rt, "pagination"), false)
		testdata.DeterministicIterations(f.ctx, t, req, f.queryClient.AllBalances, 0, true)
	})

	coins := sdk.NewCoins(
		sdk.NewCoin("stake", sdk.NewInt(10)),
		sdk.NewCoin("denom", sdk.NewInt(100)),
	)

	fundAccount(f, addr1, coins...)
	req := banktypes.NewQueryAllBalancesRequest(addr1, nil, false)

	testdata.DeterministicIterations(f.ctx, t, req, f.queryClient.AllBalances, 357, false)
}

func TestGRPCQuerySpendableBalances(t *testing.T) {
	t.Parallel()
	f := initDeterministicFixture(t)

	rapid.Check(t, func(rt *rapid.T) {
		addr := testdata.AddressGenerator(rt).Draw(rt, "address")

		// Denoms must be unique, otherwise sdk.NewCoins will panic.
		denoms := rapid.SliceOfNDistinct(rapid.StringMatching(denomRegex), 1, 10, rapid.ID[string]).Draw(rt, "denoms")
		coins := make(sdk.Coins, 0, len(denoms))
		for _, denom := range denoms {
			coin := sdk.NewCoin(
				denom,
				sdk.NewInt(rapid.Int64Min(1).Draw(rt, "amount")),
			)

			// NewCoins sorts the denoms
			coins = sdk.NewCoins(append(coins, coin)...)
		}

		err := banktestutil.FundAccount(f.bankKeeper, f.ctx, addr, coins)
		assert.NilError(t, err)

		req := banktypes.NewQuerySpendableBalancesRequest(addr, testdata.PaginationGenerator(rt, uint64(len(denoms))).Draw(rt, "pagination"))
		testdata.DeterministicIterations(f.ctx, t, req, f.queryClient.SpendableBalances, 0, true)
	})

	coins := sdk.NewCoins(
		sdk.NewCoin("stake", sdk.NewInt(10)),
		sdk.NewCoin("denom", sdk.NewInt(100)),
	)

	err := banktestutil.FundAccount(f.bankKeeper, f.ctx, addr1, coins)
	assert.NilError(t, err)

	req := banktypes.NewQuerySpendableBalancesRequest(addr1, nil)
	testdata.DeterministicIterations(f.ctx, t, req, f.queryClient.SpendableBalances, 2032, false)
}

func TestGRPCQueryTotalSupply(t *testing.T) {
	t.Parallel()
	f := initDeterministicFixture(t)

	res, err := f.queryClient.TotalSupply(f.ctx, &banktypes.QueryTotalSupplyRequest{})
	assert.NilError(t, err)
	initialSupply := res.GetSupply()

	rapid.Check(t, func(rt *rapid.T) {
		numCoins := rapid.IntRange(1, 3).Draw(rt, "num-count")
		coins := make(sdk.Coins, 0, numCoins)

		for i := 0; i < numCoins; i++ {
			coin := sdk.NewCoin(
				rapid.StringMatching(denomRegex).Draw(rt, "denom"),
				sdk.NewInt(rapid.Int64Min(1).Draw(rt, "amount")),
			)

			coins = coins.Add(coin)
		}

		assert.NilError(t, f.bankKeeper.MintCoins(f.ctx, minttypes.ModuleName, coins))

		initialSupply = initialSupply.Add(coins...)

		req := &banktypes.QueryTotalSupplyRequest{
			Pagination: testdata.PaginationGenerator(rt, uint64(len(initialSupply))).Draw(rt, "pagination"),
		}

		testdata.DeterministicIterations(f.ctx, t, req, f.queryClient.TotalSupply, 0, true)
	})

	f = initDeterministicFixture(t) // reset

	coins := sdk.NewCoins(
		sdk.NewCoin("foo", sdk.NewInt(10)),
		sdk.NewCoin("bar", sdk.NewInt(100)),
	)

	assert.NilError(t, f.bankKeeper.MintCoins(f.ctx, minttypes.ModuleName, coins))

	req := &banktypes.QueryTotalSupplyRequest{}
	testdata.DeterministicIterations(f.ctx, t, req, f.queryClient.TotalSupply, 243, false)
}

func TestGRPCQueryTotalSupplyOf(t *testing.T) {
	t.Parallel()
	f := initDeterministicFixture(t)

	rapid.Check(t, func(rt *rapid.T) {
		coin := sdk.NewCoin(
			rapid.StringMatching(denomRegex).Draw(rt, "denom"),
			sdk.NewInt(rapid.Int64Min(1).Draw(rt, "amount")),
		)

		assert.NilError(t, f.bankKeeper.MintCoins(f.ctx, minttypes.ModuleName, sdk.NewCoins(coin)))

		req := &banktypes.QuerySupplyOfRequest{Denom: coin.GetDenom()}
		testdata.DeterministicIterations(f.ctx, t, req, f.queryClient.SupplyOf, 0, true)
	})

	coin := sdk.NewCoin("bar", sdk.NewInt(100))

	assert.NilError(t, f.bankKeeper.MintCoins(f.ctx, minttypes.ModuleName, sdk.NewCoins(coin)))
	req := &banktypes.QuerySupplyOfRequest{Denom: coin.GetDenom()}
	testdata.DeterministicIterations(f.ctx, t, req, f.queryClient.SupplyOf, 1021, false)
}

func TestGRPCQueryParams(t *testing.T) {
	t.Parallel()
	f := initDeterministicFixture(t)

	rapid.Check(t, func(rt *rapid.T) {
		enabledStatus := banktypes.SendEnabled{
			Denom:   rapid.StringMatching(denomRegex).Draw(rt, "denom"),
			Enabled: rapid.Bool().Draw(rt, "status"),
		}

		params := banktypes.Params{
			SendEnabled:        []*banktypes.SendEnabled{&enabledStatus},
			DefaultSendEnabled: rapid.Bool().Draw(rt, "send"),
		}

		f.bankKeeper.SetParams(f.ctx, params)

		req := &banktypes.QueryParamsRequest{}
		testdata.DeterministicIterations(f.ctx, t, req, f.queryClient.Params, 0, true)
	})

	enabledStatus := banktypes.SendEnabled{
		Denom:   "denom",
		Enabled: true,
	}

	params := banktypes.Params{
		SendEnabled:        []*banktypes.SendEnabled{&enabledStatus},
		DefaultSendEnabled: false,
	}

	f.bankKeeper.SetParams(f.ctx, params)

	req := &banktypes.QueryParamsRequest{}
	testdata.DeterministicIterations(f.ctx, t, req, f.queryClient.Params, 1003, false)
}

func createAndReturnMetadatas(t *rapid.T, count int) []banktypes.Metadata {
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

func TestGRPCDenomsMetadata(t *testing.T) {
	t.Parallel()
	f := initDeterministicFixture(t)

	rapid.Check(t, func(rt *rapid.T) {
		count := rapid.IntRange(1, 3).Draw(rt, "count")
		denomsMetadata := createAndReturnMetadatas(rt, count)
		assert.Assert(t, len(denomsMetadata) == count)

		for i := 0; i < count; i++ {
			f.bankKeeper.SetDenomMetaData(f.ctx, denomsMetadata[i])
		}

		req := &banktypes.QueryDenomsMetadataRequest{
			Pagination: testdata.PaginationGenerator(rt, uint64(count)).Draw(rt, "pagination"),
		}

		testdata.DeterministicIterations(f.ctx, t, req, f.queryClient.DenomsMetadata, 0, true)
	})

	f = initDeterministicFixture(t) // reset

	f.bankKeeper.SetDenomMetaData(f.ctx, metadataAtom)

	req := &banktypes.QueryDenomsMetadataRequest{}
	testdata.DeterministicIterations(f.ctx, t, req, f.queryClient.DenomsMetadata, 660, false)
}

func TestGRPCDenomMetadata(t *testing.T) {
	t.Parallel()
	f := initDeterministicFixture(t)

	rapid.Check(t, func(rt *rapid.T) {
		denomMetadata := createAndReturnMetadatas(rt, 1)
		assert.Assert(t, len(denomMetadata) == 1)
		f.bankKeeper.SetDenomMetaData(f.ctx, denomMetadata[0])

		req := &banktypes.QueryDenomMetadataRequest{
			Denom: denomMetadata[0].Base,
		}

		testdata.DeterministicIterations(f.ctx, t, req, f.queryClient.DenomMetadata, 0, true)
	})

	f.bankKeeper.SetDenomMetaData(f.ctx, metadataAtom)

	req := &banktypes.QueryDenomMetadataRequest{
		Denom: metadataAtom.Base,
	}

	testdata.DeterministicIterations(f.ctx, t, req, f.queryClient.DenomMetadata, 1300, false)
}

func TestGRPCSendEnabled(t *testing.T) {
	t.Parallel()
	f := initDeterministicFixture(t)

	allDenoms := []string{}

	rapid.Check(t, func(rt *rapid.T) {
		count := rapid.IntRange(0, 10).Draw(rt, "count")
		denoms := make([]string, 0, count)

		for i := 0; i < count; i++ {
			coin := banktypes.SendEnabled{
				Denom:   rapid.StringMatching(denomRegex).Draw(rt, "denom"),
				Enabled: rapid.Bool().Draw(rt, "enabled-status"),
			}

			f.bankKeeper.SetSendEnabled(f.ctx, coin.Denom, coin.Enabled)
			denoms = append(denoms, coin.Denom)
		}

		allDenoms = append(allDenoms, denoms...)

		req := &banktypes.QuerySendEnabledRequest{
			Denoms: denoms,
			// Pagination is only taken into account when `denoms` is an empty array
			Pagination: testdata.PaginationGenerator(rt, uint64(len(allDenoms))).Draw(rt, "pagination"),
		}
		testdata.DeterministicIterations(f.ctx, t, req, f.queryClient.SendEnabled, 0, true)
	})

	coin1 := banktypes.SendEnabled{
		Denom:   "falsecoin",
		Enabled: false,
	}
	coin2 := banktypes.SendEnabled{
		Denom:   "truecoin",
		Enabled: true,
	}

	f.bankKeeper.SetSendEnabled(f.ctx, coin1.Denom, false)
	f.bankKeeper.SetSendEnabled(f.ctx, coin2.Denom, true)

	req := &banktypes.QuerySendEnabledRequest{
		Denoms: []string{coin1.GetDenom(), coin2.GetDenom()},
	}

	testdata.DeterministicIterations(f.ctx, t, req, f.queryClient.SendEnabled, 4063, false)
}

func TestGRPCDenomOwners(t *testing.T) {
	t.Parallel()
	f := initDeterministicFixture(t)

	rapid.Check(t, func(rt *rapid.T) {
		denom := rapid.StringMatching(denomRegex).Draw(rt, "denom")
		numAddr := rapid.IntRange(1, 10).Draw(rt, "number-address")
		for i := 0; i < numAddr; i++ {
			addr := testdata.AddressGenerator(rt).Draw(rt, "address")

			coin := sdk.NewCoin(
				denom,
				sdk.NewInt(rapid.Int64Min(1).Draw(rt, "amount")),
			)

			err := banktestutil.FundAccount(f.bankKeeper, f.ctx, addr, sdk.NewCoins(coin))
			assert.NilError(t, err)
		}

		req := &banktypes.QueryDenomOwnersRequest{
			Denom:      denom,
			Pagination: testdata.PaginationGenerator(rt, uint64(numAddr)).Draw(rt, "pagination"),
		}
		testdata.DeterministicIterations(f.ctx, t, req, f.queryClient.DenomOwners, 0, true)
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
		assert.NilError(t, err)

		err = banktestutil.FundAccount(f.bankKeeper, f.ctx, addr, sdk.NewCoins(coin1))
		assert.NilError(t, err)
	}

	req := &banktypes.QueryDenomOwnersRequest{
		Denom: coin1.GetDenom(),
	}
	testdata.DeterministicIterations(f.ctx, t, req, f.queryClient.DenomOwners, 2516, false)
}
