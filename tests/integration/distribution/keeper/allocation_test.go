package keeper_test

import (
	"testing"

	"cosmossdk.io/math"
	abci "github.com/cometbft/cometbft/abci/types"
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	"github.com/stretchr/testify/require"

	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	"github.com/cosmos/cosmos-sdk/x/auth/types"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	banktestutil "github.com/cosmos/cosmos-sdk/x/bank/testutil"
	dist "github.com/cosmos/cosmos-sdk/x/distribution"
	"github.com/cosmos/cosmos-sdk/x/distribution/keeper"
	"github.com/cosmos/cosmos-sdk/x/distribution/testutil"
	disttypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	stakingtestutil "github.com/cosmos/cosmos-sdk/x/staking/testutil"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

func TestAllocateTokensToValidatorWithCommission(t *testing.T) {
	var (
		bankKeeper    bankkeeper.Keeper
		distrKeeper   keeper.Keeper
		stakingKeeper *stakingkeeper.Keeper
	)

	app, err := simtestutil.Setup(testutil.AppConfig,
		&bankKeeper,
		&distrKeeper,
		&stakingKeeper,
	)
	require.NoError(t, err)

	ctx := app.BaseApp.NewContext(false, tmproto.Header{})

	addrs := simtestutil.AddTestAddrs(bankKeeper, stakingKeeper, ctx, 3, sdk.NewInt(1234))
	valAddrs := simtestutil.ConvertAddrsToValAddrs(addrs)
	tstaking := stakingtestutil.NewHelper(t, ctx, stakingKeeper)

	// create validator with 50% commission
	tstaking.Commission = stakingtypes.NewCommissionRates(sdk.NewDecWithPrec(5, 1), sdk.NewDecWithPrec(5, 1), math.LegacyNewDec(0))
	tstaking.CreateValidator(sdk.ValAddress(addrs[0]), valConsPk0, sdk.NewInt(100), true)
	val := stakingKeeper.Validator(ctx, valAddrs[0])

	// allocate tokens
	tokens := sdk.DecCoins{
		{Denom: sdk.DefaultBondDenom, Amount: math.LegacyNewDec(10)},
	}
	distrKeeper.AllocateTokensToValidator(ctx, val, tokens)

	// check commission
	expected := sdk.DecCoins{
		{Denom: sdk.DefaultBondDenom, Amount: math.LegacyNewDec(5)},
	}
	require.Equal(t, expected, distrKeeper.GetValidatorAccumulatedCommission(ctx, val.GetOperator()).Commission)

	// check current rewards
	require.Equal(t, expected, distrKeeper.GetValidatorCurrentRewards(ctx, val.GetOperator()).Rewards)
}

func TestAllocateTokensToManyValidators(t *testing.T) {
	var (
		accountKeeper authkeeper.AccountKeeper
		bankKeeper    bankkeeper.Keeper
		distrKeeper   keeper.Keeper
		stakingKeeper *stakingkeeper.Keeper
	)

	// set distribute to every block for first test.
	// we test the delayed distribution in the next test.
	dist.BlockMultipleToDistributeRewards = 1

	app, err := simtestutil.Setup(testutil.AppConfig,
		&accountKeeper,
		&bankKeeper,
		&distrKeeper,
		&stakingKeeper,
	)
	require.NoError(t, err)

	ctx := app.BaseApp.NewContext(false, tmproto.Header{})

	// reset fee pool
	distrKeeper.SetFeePool(ctx, disttypes.InitialFeePool())

	addrs := simtestutil.AddTestAddrs(bankKeeper, stakingKeeper, ctx, 2, sdk.NewInt(1234))
	valAddrs := simtestutil.ConvertAddrsToValAddrs(addrs)
	tstaking := stakingtestutil.NewHelper(t, ctx, stakingKeeper)

	// create validator with 50% commission
	tstaking.Commission = stakingtypes.NewCommissionRates(sdk.NewDecWithPrec(5, 1), sdk.NewDecWithPrec(5, 1), math.LegacyNewDec(0))
	tstaking.CreateValidator(valAddrs[0], valConsPk0, sdk.NewInt(100), true)

	// create second validator with 0% commission
	tstaking.Commission = stakingtypes.NewCommissionRates(math.LegacyNewDec(0), math.LegacyNewDec(0), math.LegacyNewDec(0))
	tstaking.CreateValidator(valAddrs[1], valConsPk1, sdk.NewInt(100), true)

	abciValA := abci.Validator{
		Address: valConsPk0.Address(),
		Power:   100,
	}
	abciValB := abci.Validator{
		Address: valConsPk1.Address(),
		Power:   100,
	}

	// assert initial state: zero outstanding rewards, zero community pool, zero commission, zero current rewards
	require.True(t, distrKeeper.GetValidatorOutstandingRewards(ctx, valAddrs[0]).Rewards.IsZero())
	require.True(t, distrKeeper.GetValidatorOutstandingRewards(ctx, valAddrs[1]).Rewards.IsZero())
	require.True(t, distrKeeper.GetFeePool(ctx).CommunityPool.IsZero())
	require.True(t, distrKeeper.GetValidatorAccumulatedCommission(ctx, valAddrs[0]).Commission.IsZero())
	require.True(t, distrKeeper.GetValidatorAccumulatedCommission(ctx, valAddrs[1]).Commission.IsZero())
	require.True(t, distrKeeper.GetValidatorCurrentRewards(ctx, valAddrs[0]).Rewards.IsZero())
	require.True(t, distrKeeper.GetValidatorCurrentRewards(ctx, valAddrs[1]).Rewards.IsZero())

	// allocate tokens as if both had voted and second was proposer
	fees := sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(100)))
	feeCollector := accountKeeper.GetModuleAccount(ctx, types.FeeCollectorName)
	require.NotNil(t, feeCollector)

	// fund fee collector
	require.NoError(t, banktestutil.FundModuleAccount(bankKeeper, ctx, feeCollector.GetName(), fees))

	accountKeeper.SetAccount(ctx, feeCollector)

	votes := []abci.VoteInfo{
		{
			Validator:       abciValA,
			SignedLastBlock: true,
		},
		{
			Validator:       abciValB,
			SignedLastBlock: true,
		},
	}
	distrKeeper.AllocateTokens(ctx, 200, votes)

	// 98 outstanding rewards (100 less 2 to community pool)
	firstValidator0OutstandingRewards := distrKeeper.GetValidatorOutstandingRewards(ctx, valAddrs[0]).Rewards
	firstValidator1OutstandingRewards := distrKeeper.GetValidatorOutstandingRewards(ctx, valAddrs[1]).Rewards
	require.Equal(t, sdk.DecCoins{{Denom: sdk.DefaultBondDenom, Amount: sdk.NewDecWithPrec(490, 1)}}, firstValidator0OutstandingRewards)
	require.Equal(t, sdk.DecCoins{{Denom: sdk.DefaultBondDenom, Amount: sdk.NewDecWithPrec(490, 1)}}, firstValidator1OutstandingRewards)

	// 2 community pool coins
	require.Equal(t, sdk.DecCoins{{Denom: sdk.DefaultBondDenom, Amount: math.LegacyNewDec(2)}}, distrKeeper.GetFeePool(ctx).CommunityPool)

	// 50% commission for first proposer, (0.5 * 98%) * 100 / 2 = 23.25
	firstValidator0Commission := distrKeeper.GetValidatorAccumulatedCommission(ctx, valAddrs[0]).Commission
	require.Equal(t, sdk.DecCoins{{Denom: sdk.DefaultBondDenom, Amount: sdk.NewDecWithPrec(2450, 2)}}, firstValidator0Commission)

	// zero commission for second proposer
	firstValidator1Commission := distrKeeper.GetValidatorAccumulatedCommission(ctx, valAddrs[1]).Commission
	require.True(t, firstValidator1Commission.IsZero())

	// just staking.proportional for first proposer less commission = (0.5 * 98%) * 100 / 2 = 24.50
	firstValidator0CurrentRewards := distrKeeper.GetValidatorCurrentRewards(ctx, valAddrs[0]).Rewards
	require.Equal(t, sdk.DecCoins{{Denom: sdk.DefaultBondDenom, Amount: sdk.NewDecWithPrec(2450, 2)}}, firstValidator0CurrentRewards)

	// proposer reward + staking.proportional for second proposer = (0.5 * (98%)) * 100 = 49
	firstValidator1CurrentRewards := distrKeeper.GetValidatorCurrentRewards(ctx, valAddrs[1]).Rewards
	require.Equal(t, sdk.DecCoins{{Denom: sdk.DefaultBondDenom, Amount: sdk.NewDecWithPrec(490, 1)}}, firstValidator1CurrentRewards)

	// test that the block height triggers the distribution
	dist.BlockMultipleToDistributeRewards = 50

	// block height is not a multiple, should not trigger allocation (no change in rewards)
	ctx = ctx.WithBlockHeight(dist.BlockMultipleToDistributeRewards - 1)
	app.BeginBlocker(ctx, abci.RequestBeginBlock{Header: tmproto.Header{ProposerAddress: valAddrs[0].Bytes()},
		LastCommitInfo: abci.CommitInfo{
			Votes: votes,
		},
	})
	require.Equal(t, firstValidator0OutstandingRewards, distrKeeper.GetValidatorOutstandingRewards(ctx, valAddrs[0]).Rewards)
	require.Equal(t, firstValidator1OutstandingRewards, distrKeeper.GetValidatorOutstandingRewards(ctx, valAddrs[1]).Rewards)
	require.Equal(t, firstValidator0Commission, distrKeeper.GetValidatorAccumulatedCommission(ctx, valAddrs[0]).Commission)
	require.Equal(t, firstValidator1Commission, distrKeeper.GetValidatorAccumulatedCommission(ctx, valAddrs[1]).Commission)
	require.Equal(t, firstValidator0CurrentRewards, distrKeeper.GetValidatorCurrentRewards(ctx, valAddrs[0]).Rewards)
	require.Equal(t, firstValidator1CurrentRewards, distrKeeper.GetValidatorCurrentRewards(ctx, valAddrs[1]).Rewards)

	// block height is a multiple, should trigger allocation
	ctx = ctx.WithBlockHeight(dist.BlockMultipleToDistributeRewards)

	feesCollectedInt := bankKeeper.GetAllBalances(ctx, feeCollector.GetAddress())

	// feesCollected was increased from last BeginBlocker call, then will occur again from the new BeginBlocker call,
	// so we need to double the feesCollected to simulate the new BeginBlocker call
	feesCollectedInt[0].Amount = feesCollectedInt[0].Amount.MulRaw(2)
	feesCollected := sdk.NewDecCoinsFromCoins(feesCollectedInt...)

	communityTax := distrKeeper.GetCommunityTax(ctx)
	voteMultiplier := math.LegacyOneDec().Sub(communityTax)
	feeMultiplier := feesCollected.MulDecTruncate(voteMultiplier)
	powerFraction := math.LegacyNewDec(100).QuoTruncate(math.LegacyNewDec(200))

	newRewards := feeMultiplier.MulDecTruncate(powerFraction)
	pendingRewards := firstValidator0OutstandingRewards.Add(newRewards...)

	pendingCommission := firstValidator0OutstandingRewards.Add(newRewards...)
	pendingCommission[0].Amount = pendingCommission[0].Amount.Quo(sdk.NewDec(2))

	app.BeginBlocker(ctx, abci.RequestBeginBlock{Header: tmproto.Header{ProposerAddress: valAddrs[0].Bytes()},
		LastCommitInfo: abci.CommitInfo{
			Votes: votes,
		},
	})

	require.Equal(t, pendingRewards, distrKeeper.GetValidatorOutstandingRewards(ctx, valAddrs[0]).Rewards)
	require.Equal(t, pendingRewards, distrKeeper.GetValidatorOutstandingRewards(ctx, valAddrs[1]).Rewards)
	require.Equal(t, pendingCommission, distrKeeper.GetValidatorAccumulatedCommission(ctx, valAddrs[0]).Commission)
	require.True(t, distrKeeper.GetValidatorAccumulatedCommission(ctx, valAddrs[1]).Commission.IsZero())
	require.Equal(t, pendingCommission, distrKeeper.GetValidatorCurrentRewards(ctx, valAddrs[0]).Rewards)
	require.Equal(t, pendingRewards, distrKeeper.GetValidatorCurrentRewards(ctx, valAddrs[1]).Rewards)
}

func TestAllocateTokensTruncation(t *testing.T) {
	var (
		accountKeeper authkeeper.AccountKeeper
		bankKeeper    bankkeeper.Keeper
		distrKeeper   keeper.Keeper
		stakingKeeper *stakingkeeper.Keeper
	)

	app, err := simtestutil.Setup(testutil.AppConfig,
		&accountKeeper,
		&bankKeeper,
		&distrKeeper,
		&stakingKeeper,
	)
	require.NoError(t, err)

	ctx := app.BaseApp.NewContext(false, tmproto.Header{})

	// reset fee pool
	distrKeeper.SetFeePool(ctx, disttypes.InitialFeePool())

	addrs := simtestutil.AddTestAddrs(bankKeeper, stakingKeeper, ctx, 3, sdk.NewInt(1234))
	valAddrs := simtestutil.ConvertAddrsToValAddrs(addrs)
	tstaking := stakingtestutil.NewHelper(t, ctx, stakingKeeper)

	// create validator with 10% commission
	tstaking.Commission = stakingtypes.NewCommissionRates(sdk.NewDecWithPrec(1, 1), sdk.NewDecWithPrec(1, 1), math.LegacyNewDec(0))
	tstaking.CreateValidator(valAddrs[0], valConsPk0, sdk.NewInt(110), true)

	// create second validator with 10% commission
	tstaking.Commission = stakingtypes.NewCommissionRates(sdk.NewDecWithPrec(1, 1), sdk.NewDecWithPrec(1, 1), math.LegacyNewDec(0))
	tstaking.CreateValidator(valAddrs[1], valConsPk1, sdk.NewInt(100), true)

	// create third validator with 10% commission
	tstaking.Commission = stakingtypes.NewCommissionRates(sdk.NewDecWithPrec(1, 1), sdk.NewDecWithPrec(1, 1), math.LegacyNewDec(0))
	tstaking.CreateValidator(valAddrs[2], valConsPk2, sdk.NewInt(100), true)

	abciValA := abci.Validator{
		Address: valConsPk0.Address(),
		Power:   11,
	}
	abciValB := abci.Validator{
		Address: valConsPk1.Address(),
		Power:   10,
	}
	abciValС := abci.Validator{
		Address: valConsPk2.Address(),
		Power:   10,
	}

	// assert initial state: zero outstanding rewards, zero community pool, zero commission, zero current rewards
	require.True(t, distrKeeper.GetValidatorOutstandingRewards(ctx, valAddrs[0]).Rewards.IsZero())
	require.True(t, distrKeeper.GetValidatorOutstandingRewards(ctx, valAddrs[1]).Rewards.IsZero())
	require.True(t, distrKeeper.GetValidatorOutstandingRewards(ctx, valAddrs[1]).Rewards.IsZero())
	require.True(t, distrKeeper.GetFeePool(ctx).CommunityPool.IsZero())
	require.True(t, distrKeeper.GetValidatorAccumulatedCommission(ctx, valAddrs[0]).Commission.IsZero())
	require.True(t, distrKeeper.GetValidatorAccumulatedCommission(ctx, valAddrs[1]).Commission.IsZero())
	require.True(t, distrKeeper.GetValidatorCurrentRewards(ctx, valAddrs[0]).Rewards.IsZero())
	require.True(t, distrKeeper.GetValidatorCurrentRewards(ctx, valAddrs[1]).Rewards.IsZero())

	// allocate tokens as if both had voted and second was proposer
	fees := sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(634195840)))

	feeCollector := accountKeeper.GetModuleAccount(ctx, types.FeeCollectorName)
	require.NotNil(t, feeCollector)

	require.NoError(t, banktestutil.FundModuleAccount(bankKeeper, ctx, feeCollector.GetName(), fees))

	accountKeeper.SetAccount(ctx, feeCollector)

	votes := []abci.VoteInfo{
		{
			Validator:       abciValA,
			SignedLastBlock: true,
		},
		{
			Validator:       abciValB,
			SignedLastBlock: true,
		},
		{
			Validator:       abciValС,
			SignedLastBlock: true,
		},
	}
	distrKeeper.AllocateTokens(ctx, 31, votes)

	require.True(t, distrKeeper.GetValidatorOutstandingRewards(ctx, valAddrs[0]).Rewards.IsValid())
	require.True(t, distrKeeper.GetValidatorOutstandingRewards(ctx, valAddrs[1]).Rewards.IsValid())
	require.True(t, distrKeeper.GetValidatorOutstandingRewards(ctx, valAddrs[2]).Rewards.IsValid())
}
