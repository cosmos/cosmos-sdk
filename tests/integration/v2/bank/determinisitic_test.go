package bank

import (
	"context"
	"fmt"
	"testing"

	"github.com/cosmos/gogoproto/proto"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"pgregory.net/rapid"

	"cosmossdk.io/core/gas"
	"cosmossdk.io/depinject"
	"cosmossdk.io/log"
	"cosmossdk.io/math"
	bankkeeper "cosmossdk.io/x/bank/keeper"
	banktestutil "cosmossdk.io/x/bank/testutil"
	banktypes "cosmossdk.io/x/bank/types"
	minttypes "cosmossdk.io/x/mint/types"

	codectestutil "github.com/cosmos/cosmos-sdk/codec/testutil"
	"github.com/cosmos/cosmos-sdk/tests/integration/v2"
	"github.com/cosmos/cosmos-sdk/testutil/configurator"
	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtestutil "github.com/cosmos/cosmos-sdk/x/auth/testutil"
)

var (
	denomRegex   = `[a-zA-Z][a-zA-Z0-9/:._-]{2,127}`
	coin1        = sdk.NewCoin("denom", math.NewInt(10))
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
	*testing.T
	ctx        context.Context
	app        *integration.App
	bankKeeper bankkeeper.Keeper
}

func queryFnFactory[RequestT, ResponseT proto.Message](
	f *deterministicFixture,
) func(RequestT) (ResponseT, error) {
	return func(req RequestT) (ResponseT, error) {
		var emptyResponse ResponseT
		res, err := f.app.Query(f.ctx, 0, req)
		if err != nil {
			return emptyResponse, err
		}
		castedRes, ok := res.(ResponseT)
		if !ok {
			return emptyResponse, fmt.Errorf("unexpected response type: %T", res)
		}
		return castedRes, nil
	}
}

func fundAccount(f *deterministicFixture, addr sdk.AccAddress, coin ...sdk.Coin) {
	err := banktestutil.FundAccount(f.ctx, f.bankKeeper, addr, sdk.NewCoins(coin...))
	require.NoError(f.T, err)
}

func getCoin(rt *rapid.T) sdk.Coin {
	return sdk.NewCoin(
		rapid.StringMatching(denomRegex).Draw(rt, "denom"),
		math.NewInt(rapid.Int64Min(1).Draw(rt, "amount")),
	)
}

func initDeterministicFixture(t *testing.T) *deterministicFixture {
	t.Helper()

	ctrl := gomock.NewController(t)
	acctsModKeeper := authtestutil.NewMockAccountsModKeeper(ctrl)
	accNum := uint64(0)
	acctsModKeeper.EXPECT().NextAccountNumber(gomock.Any()).AnyTimes().DoAndReturn(func(ctx context.Context) (
		uint64, error,
	) {
		currentNum := accNum
		accNum++
		return currentNum, nil
	})

	startupConfig := integration.DefaultStartUpConfig(t)
	startupConfig.GenesisBehavior = integration.Genesis_SKIP
	diConfig := configurator.NewAppV2Config(
		configurator.TxModule(),
		configurator.AuthModule(),
		configurator.BankModule(),
	)

	var bankKeeper bankkeeper.Keeper
	diConfig = depinject.Configs(diConfig, depinject.Supply(acctsModKeeper, log.NewNopLogger()))
	app, err := integration.NewApp(diConfig, startupConfig, &bankKeeper)
	require.NoError(t, err)
	require.NotNil(t, app)
	return &deterministicFixture{app: app, bankKeeper: bankKeeper, T: t}
}

func assertNonZeroGas(t *testing.T, gasUsed gas.Gas) {
	t.Helper()
	require.NotZero(t, gasUsed)
}

func TestQueryBalance(t *testing.T) {
	t.Parallel()
	f := initDeterministicFixture(t)
	f.ctx = f.app.StateLatestContext(t)
	gasMeterFactory := integration.GasMeterFactory(f.ctx)
	queryFn := queryFnFactory[*banktypes.QueryBalanceRequest, *banktypes.QueryBalanceResponse](f)
	assertBalance := func(coin sdk.Coin) func(t *testing.T, res *banktypes.QueryBalanceResponse) {
		return func(t *testing.T, res *banktypes.QueryBalanceResponse) {
			t.Helper()
			require.Equal(t, coin.Denom, res.Balance.Denom)
			require.Truef(t, coin.Amount.Equal(res.Balance.Amount),
				"expected %s, got %s", coin.Amount, res.Balance.Amount)
		}
	}

	rapid.Check(t, func(rt *rapid.T) {
		addr := testdata.AddressGenerator(rt).Draw(rt, "address")
		coin := getCoin(rt)
		fundAccount(f, addr, coin)

		addrStr, err := codectestutil.CodecOptions{}.GetAddressCodec().BytesToString(addr)
		require.NoError(t, err)

		req := banktypes.NewQueryBalanceRequest(addrStr, coin.GetDenom())

		testdata.DeterministicIterationsV2(t, req, gasMeterFactory, queryFn, assertNonZeroGas, assertBalance(coin))
	})

	fundAccount(f, addr1, coin1)
	addr1Str, err := codectestutil.CodecOptions{}.GetAddressCodec().BytesToString(addr1)
	require.NoError(t, err)
	req := banktypes.NewQueryBalanceRequest(addr1Str, coin1.GetDenom())
	testdata.DeterministicIterationsV2(t, req, gasMeterFactory, queryFn, assertNonZeroGas, assertBalance(coin1))
}

func TestQueryAllBalances(t *testing.T) {
	t.Parallel()
	f := initDeterministicFixture(t)
	f.ctx = f.app.StateLatestContext(t)
	gasMeterFactory := integration.GasMeterFactory(f.ctx)
	addressCodec := codectestutil.CodecOptions{}.GetAddressCodec()
	queryFn := queryFnFactory[*banktypes.QueryAllBalancesRequest, *banktypes.QueryAllBalancesResponse](f)

	rapid.Check(t, func(rt *rapid.T) {
		addr := testdata.AddressGenerator(rt).Draw(rt, "address")
		numCoins := rapid.IntRange(1, 10).Draw(rt, "num-count")
		coins := make(sdk.Coins, 0, numCoins)

		addrStr, err := addressCodec.BytesToString(addr)
		require.NoError(t, err)

		for i := 0; i < numCoins; i++ {
			coin := getCoin(rt)
			if exists, _ := coins.Find(coin.Denom); exists {
				t.Skip("duplicate denom")
			}
			// NewCoins sorts the denoms
			coins = sdk.NewCoins(append(coins, coin)...)
		}

		fundAccount(f, addr, coins...)

		req := banktypes.NewQueryAllBalancesRequest(
			addrStr, testdata.PaginationGenerator(rt, uint64(numCoins)).Draw(rt, "pagination"), false)
		testdata.DeterministicIterationsV2(t, req, gasMeterFactory, queryFn, assertNonZeroGas, nil)
	})

	coins := sdk.NewCoins(
		sdk.NewCoin("stake", math.NewInt(10)),
		sdk.NewCoin("denom", math.NewInt(100)),
	)

	fundAccount(f, addr1, coins...)
	addr1Str, err := addressCodec.BytesToString(addr1)
	require.NoError(t, err)

	req := banktypes.NewQueryAllBalancesRequest(addr1Str, nil, false)
	testdata.DeterministicIterationsV2(t, req, gasMeterFactory, queryFn, assertNonZeroGas, nil)
}

func TestQuerySpendableBalances(t *testing.T) {
	t.Parallel()
	f := initDeterministicFixture(t)
	f.ctx = f.app.StateLatestContext(t)
	gasMeterFactory := integration.GasMeterFactory(f.ctx)
	queryFn := queryFnFactory[*banktypes.QuerySpendableBalancesRequest, *banktypes.QuerySpendableBalancesResponse](f)

	rapid.Check(t, func(rt *rapid.T) {
		addr := testdata.AddressGenerator(rt).Draw(rt, "address")
		addrStr, err := codectestutil.CodecOptions{}.GetAddressCodec().BytesToString(addr)
		require.NoError(t, err)

		// Denoms must be unique, otherwise sdk.NewCoins will panic.
		denoms := rapid.SliceOfNDistinct(rapid.StringMatching(denomRegex), 1, 10, rapid.ID[string]).Draw(rt, "denoms")
		coins := make(sdk.Coins, 0, len(denoms))
		for _, denom := range denoms {
			coin := sdk.NewCoin(
				denom,
				math.NewInt(rapid.Int64Min(1).Draw(rt, "amount")),
			)

			// NewCoins sorts the denoms
			coins = sdk.NewCoins(append(coins, coin)...)
		}

		err = banktestutil.FundAccount(f.ctx, f.bankKeeper, addr, coins)
		require.NoError(t, err)

		req := banktypes.NewQuerySpendableBalancesRequest(addrStr, testdata.PaginationGenerator(rt, uint64(len(denoms))).Draw(rt, "pagination"))
		testdata.DeterministicIterationsV2(t, req, gasMeterFactory, queryFn, assertNonZeroGas, nil)
	})

	coins := sdk.NewCoins(
		sdk.NewCoin("stake", math.NewInt(10)),
		sdk.NewCoin("denom", math.NewInt(100)),
	)

	err := banktestutil.FundAccount(f.ctx, f.bankKeeper, addr1, coins)
	require.NoError(t, err)

	addr1Str, err := codectestutil.CodecOptions{}.GetAddressCodec().BytesToString(addr1)
	require.NoError(t, err)

	req := banktypes.NewQuerySpendableBalancesRequest(addr1Str, nil)
	testdata.DeterministicIterationsV2(t, req, gasMeterFactory, queryFn, assertNonZeroGas, nil)
}

func TestQueryTotalSupply(t *testing.T) {
	t.Parallel()
	f := initDeterministicFixture(t)
	f.ctx = f.app.StateLatestContext(t)
	gasMeterFactory := integration.GasMeterFactory(f.ctx)
	queryFn := queryFnFactory[*banktypes.QueryTotalSupplyRequest, *banktypes.QueryTotalSupplyResponse](f)

	res, err := queryFn(&banktypes.QueryTotalSupplyRequest{})
	require.NoError(t, err)
	initialSupply := res.GetSupply()

	rapid.Check(t, func(rt *rapid.T) {
		numCoins := rapid.IntRange(1, 3).Draw(rt, "num-count")
		coins := make(sdk.Coins, 0, numCoins)

		for i := 0; i < numCoins; i++ {
			coin := sdk.NewCoin(
				rapid.StringMatching(denomRegex).Draw(rt, "denom"),
				math.NewInt(rapid.Int64Min(1).Draw(rt, "amount")),
			)

			coins = coins.Add(coin)
		}

		require.NoError(t, f.bankKeeper.MintCoins(f.ctx, minttypes.ModuleName, coins))

		initialSupply = initialSupply.Add(coins...)

		req := &banktypes.QueryTotalSupplyRequest{
			Pagination: testdata.PaginationGenerator(rt, uint64(len(initialSupply))).Draw(rt, "pagination"),
		}

		testdata.DeterministicIterationsV2(t, req, gasMeterFactory, queryFn, assertNonZeroGas, nil)
	})

	f = initDeterministicFixture(t) // reset
	f.ctx = f.app.StateLatestContext(t)
	gasMeterFactory = integration.GasMeterFactory(f.ctx)
	queryFn = queryFnFactory[*banktypes.QueryTotalSupplyRequest, *banktypes.QueryTotalSupplyResponse](f)

	coins := sdk.NewCoins(
		sdk.NewCoin("foo", math.NewInt(10)),
		sdk.NewCoin("bar", math.NewInt(100)),
	)

	require.NoError(t, f.bankKeeper.MintCoins(f.ctx, minttypes.ModuleName, coins))

	req := &banktypes.QueryTotalSupplyRequest{}
	testdata.DeterministicIterationsV2(t, req, gasMeterFactory, queryFn, assertNonZeroGas, nil)
}

func TestQueryTotalSupplyOf(t *testing.T) {
	t.Parallel()
	f := initDeterministicFixture(t)
	f.ctx = f.app.StateLatestContext(t)
	gasMeterFactory := integration.GasMeterFactory(f.ctx)
	queryFn := queryFnFactory[*banktypes.QuerySupplyOfRequest, *banktypes.QuerySupplyOfResponse](f)

	rapid.Check(t, func(rt *rapid.T) {
		coin := sdk.NewCoin(
			rapid.StringMatching(denomRegex).Draw(rt, "denom"),
			math.NewInt(rapid.Int64Min(1).Draw(rt, "amount")),
		)

		require.NoError(t, f.bankKeeper.MintCoins(f.ctx, minttypes.ModuleName, sdk.NewCoins(coin)))

		req := &banktypes.QuerySupplyOfRequest{Denom: coin.GetDenom()}
		testdata.DeterministicIterationsV2(t, req, gasMeterFactory, queryFn, assertNonZeroGas, nil)
	})

	coin := sdk.NewCoin("bar", math.NewInt(100))

	require.NoError(t, f.bankKeeper.MintCoins(f.ctx, minttypes.ModuleName, sdk.NewCoins(coin)))
	req := &banktypes.QuerySupplyOfRequest{Denom: coin.GetDenom()}
	testdata.DeterministicIterationsV2(t, req, gasMeterFactory, queryFn, assertNonZeroGas, nil)
}

func TestQueryParams(t *testing.T) {
	t.Parallel()
	f := initDeterministicFixture(t)
	f.ctx = f.app.StateLatestContext(t)
	gasMeterFactory := integration.GasMeterFactory(f.ctx)
	queryFn := queryFnFactory[*banktypes.QueryParamsRequest, *banktypes.QueryParamsResponse](f)

	rapid.Check(t, func(rt *rapid.T) {
		enabledStatus := banktypes.SendEnabled{
			Denom:   rapid.StringMatching(denomRegex).Draw(rt, "denom"),
			Enabled: rapid.Bool().Draw(rt, "status"),
		}

		params := banktypes.Params{
			SendEnabled:        []*banktypes.SendEnabled{&enabledStatus},
			DefaultSendEnabled: rapid.Bool().Draw(rt, "send"),
		}

		err := f.bankKeeper.SetParams(f.ctx, params)
		require.NoError(t, err)

		req := &banktypes.QueryParamsRequest{}
		testdata.DeterministicIterationsV2(t, req, gasMeterFactory, queryFn, assertNonZeroGas, nil)
	})

	enabledStatus := banktypes.SendEnabled{
		Denom:   "denom",
		Enabled: true,
	}

	params := banktypes.Params{
		SendEnabled:        []*banktypes.SendEnabled{&enabledStatus},
		DefaultSendEnabled: false,
	}

	err := f.bankKeeper.SetParams(f.ctx, params)
	require.NoError(t, err)
	req := &banktypes.QueryParamsRequest{}
	testdata.DeterministicIterationsV2(t, req, gasMeterFactory, queryFn, assertNonZeroGas, nil)
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

func TestDenomsMetadata(t *testing.T) {
	t.Parallel()
	f := initDeterministicFixture(t)
	f.ctx = f.app.StateLatestContext(t)
	gasMeterFactory := integration.GasMeterFactory(f.ctx)
	queryFn := queryFnFactory[*banktypes.QueryDenomsMetadataRequest, *banktypes.QueryDenomsMetadataResponse](f)

	rapid.Check(t, func(rt *rapid.T) {
		count := rapid.IntRange(1, 3).Draw(rt, "count")
		denomsMetadata := createAndReturnMetadatas(rt, count)
		require.True(t, len(denomsMetadata) == count)

		for i := 0; i < count; i++ {
			f.bankKeeper.SetDenomMetaData(f.ctx, denomsMetadata[i])
		}

		req := &banktypes.QueryDenomsMetadataRequest{
			Pagination: testdata.PaginationGenerator(rt, uint64(count)).Draw(rt, "pagination"),
		}

		testdata.DeterministicIterationsV2(t, req, gasMeterFactory, queryFn, assertNonZeroGas, nil)
	})

	require.NoError(t, f.app.Close())

	f = initDeterministicFixture(t) // reset
	f.ctx = f.app.StateLatestContext(t)
	gasMeterFactory = integration.GasMeterFactory(f.ctx)
	queryFn = queryFnFactory[*banktypes.QueryDenomsMetadataRequest, *banktypes.QueryDenomsMetadataResponse](f)

	f.bankKeeper.SetDenomMetaData(f.ctx, metadataAtom)

	req := &banktypes.QueryDenomsMetadataRequest{}
	testdata.DeterministicIterationsV2(t, req, gasMeterFactory, queryFn, assertNonZeroGas, nil)
}

func TestDenomMetadata(t *testing.T) {
	t.Parallel()
	f := initDeterministicFixture(t)
	f.ctx = f.app.StateLatestContext(t)
	gasMeterFactory := integration.GasMeterFactory(f.ctx)
	queryFn := queryFnFactory[*banktypes.QueryDenomMetadataRequest, *banktypes.QueryDenomMetadataResponse](f)

	rapid.Check(t, func(rt *rapid.T) {
		denomMetadata := createAndReturnMetadatas(rt, 1)
		require.True(t, len(denomMetadata) == 1)
		f.bankKeeper.SetDenomMetaData(f.ctx, denomMetadata[0])

		req := &banktypes.QueryDenomMetadataRequest{
			Denom: denomMetadata[0].Base,
		}

		testdata.DeterministicIterationsV2(t, req, gasMeterFactory, queryFn, assertNonZeroGas, nil)
	})

	f.bankKeeper.SetDenomMetaData(f.ctx, metadataAtom)

	req := &banktypes.QueryDenomMetadataRequest{
		Denom: metadataAtom.Base,
	}

	testdata.DeterministicIterationsV2(t, req, gasMeterFactory, queryFn, assertNonZeroGas, nil)
}

func TestSendEnabled(t *testing.T) {
	t.Parallel()
	f := initDeterministicFixture(t)
	f.ctx = f.app.StateLatestContext(t)
	gasMeterFactory := integration.GasMeterFactory(f.ctx)
	queryFn := queryFnFactory[*banktypes.QuerySendEnabledRequest, *banktypes.QuerySendEnabledResponse](f)
	allDenoms := []string{}

	rapid.Check(t, func(rt *rapid.T) {
		count := rapid.IntRange(1, 10).Draw(rt, "count")
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
		testdata.DeterministicIterationsV2(t, req, gasMeterFactory, queryFn, assertNonZeroGas, nil)
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

	testdata.DeterministicIterationsV2(t, req, gasMeterFactory, queryFn, assertNonZeroGas, nil)
}

func TestDenomOwners(t *testing.T) {
	t.Parallel()
	f := initDeterministicFixture(t)
	f.ctx = f.app.StateLatestContext(t)
	gasMeterFactory := integration.GasMeterFactory(f.ctx)
	queryFn := queryFnFactory[*banktypes.QueryDenomOwnersRequest, *banktypes.QueryDenomOwnersResponse](f)

	rapid.Check(t, func(rt *rapid.T) {
		denom := rapid.StringMatching(denomRegex).Draw(rt, "denom")
		numAddr := rapid.IntRange(1, 10).Draw(rt, "number-address")
		for i := 0; i < numAddr; i++ {
			addr := testdata.AddressGenerator(rt).Draw(rt, "address")

			coin := sdk.NewCoin(
				denom,
				math.NewInt(rapid.Int64Min(1).Draw(rt, "amount")),
			)

			err := banktestutil.FundAccount(f.ctx, f.bankKeeper, addr, sdk.NewCoins(coin))
			require.NoError(t, err)
		}

		req := &banktypes.QueryDenomOwnersRequest{
			Denom:      denom,
			Pagination: testdata.PaginationGenerator(rt, uint64(numAddr)).Draw(rt, "pagination"),
		}
		testdata.DeterministicIterationsV2(t, req, gasMeterFactory, queryFn, assertNonZeroGas, nil)
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
		require.NoError(t, err)

		err = banktestutil.FundAccount(f.ctx, f.bankKeeper, addr, sdk.NewCoins(coin1))
		require.NoError(t, err)
	}

	req := &banktypes.QueryDenomOwnersRequest{
		Denom: coin1.GetDenom(),
	}
	testdata.DeterministicIterationsV2(t, req, gasMeterFactory, queryFn, assertNonZeroGas, nil)
}
