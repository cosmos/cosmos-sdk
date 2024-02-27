package slashing

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"cosmossdk.io/core/comet"
	coreheader "cosmossdk.io/core/header"
	"cosmossdk.io/depinject"
	"cosmossdk.io/log"
	authkeeper "cosmossdk.io/x/auth/keeper"
	bankkeeper "cosmossdk.io/x/bank/keeper"
	"cosmossdk.io/x/slashing"
	slashingkeeper "cosmossdk.io/x/slashing/keeper"
	"cosmossdk.io/x/slashing/testutil"
	stakingkeeper "cosmossdk.io/x/staking/keeper"
	stakingtestutil "cosmossdk.io/x/staking/testutil"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// TestBeginBlocker is a unit test function that tests the behavior of the BeginBlocker function.
// It sets up the necessary dependencies and context, creates a validator, and performs various operations
// to test the slashing logic. It checks if the validator is correctly jailed after a certain number of blocks.
func TestBeginBlocker(t *testing.T) {
	var (
		interfaceRegistry codectypes.InterfaceRegistry
		accountKeeper     authkeeper.AccountKeeper
		bankKeeper        bankkeeper.Keeper
		stakingKeeper     *stakingkeeper.Keeper
		slashingKeeper    slashingkeeper.Keeper
	)

	app, err := simtestutil.Setup(
		depinject.Configs(
			AppConfig,
			depinject.Supply(log.NewNopLogger()),
		),
		&interfaceRegistry,
		&accountKeeper,
		&bankKeeper,
		&stakingKeeper,
		&slashingKeeper,
	)
	require.NoError(t, err)

	ctx := app.BaseApp.NewContext(false)

	pks := simtestutil.CreateTestPubKeys(1)
	simtestutil.AddTestAddrsFromPubKeys(bankKeeper, stakingKeeper, ctx, pks, stakingKeeper.TokensFromConsensusPower(ctx, 200))
	addr, pk := sdk.ValAddress(pks[0].Address()), pks[0]
	tstaking := stakingtestutil.NewHelper(t, ctx, stakingKeeper)

	// bond the validator
	power := int64(100)
	acc := accountKeeper.NewAccountWithAddress(ctx, sdk.AccAddress(addr))
	accountKeeper.SetAccount(ctx, acc)
	amt := tstaking.CreateValidatorWithValPower(addr, pk, power, true)
	_, err = stakingKeeper.EndBlocker(ctx)
	require.NoError(t, err)
	bondDenom, err := stakingKeeper.BondDenom(ctx)
	require.NoError(t, err)
	require.Equal(
		t, bankKeeper.GetAllBalances(ctx, sdk.AccAddress(addr)),
		sdk.NewCoins(sdk.NewCoin(bondDenom, testutil.InitTokens.Sub(amt))),
	)
	val, err := stakingKeeper.Validator(ctx, addr)
	require.NoError(t, err)
	require.Equal(t, amt, val.GetBondedTokens())

	abciVal := comet.Validator{
		Address: pk.Address(),
		Power:   power,
	}

	ctx = ctx.WithCometInfo(comet.Info{
		LastCommit: comet.CommitInfo{Votes: []comet.VoteInfo{{
			Validator:   abciVal,
			BlockIDFlag: comet.BlockIDFlagCommit,
		}}},
	})

	err = slashing.BeginBlocker(ctx, slashingKeeper)
	require.NoError(t, err)

	info, err := slashingKeeper.ValidatorSigningInfo.Get(ctx, sdk.ConsAddress(pk.Address()))
	require.NoError(t, err)
	require.Equal(t, ctx.HeaderInfo().Height, info.StartHeight)
	require.Equal(t, int64(0), info.IndexOffset)
	require.Equal(t, time.Unix(0, 0).UTC(), info.JailedUntil)
	require.Equal(t, int64(0), info.MissedBlocksCounter)

	height := int64(0)

	signedBlocksWindow, err := slashingKeeper.SignedBlocksWindow(ctx)
	require.NoError(t, err)
	// for 100 blocks, mark the validator as having signed
	for ; height < signedBlocksWindow; height++ {
		ctx = ctx.WithHeaderInfo(coreheader.Info{Height: height})

		err = slashing.BeginBlocker(ctx, slashingKeeper)
		require.NoError(t, err)
	}

	minSignedPerWindow, err := slashingKeeper.MinSignedPerWindow(ctx)
	require.NoError(t, err)
	// for 50 blocks, mark the validator as having not signed
	for ; height < ((signedBlocksWindow * 2) - minSignedPerWindow + 1); height++ {
		ctx = ctx.WithHeaderInfo(coreheader.Info{Height: height}).WithCometInfo(comet.Info{
			LastCommit: comet.CommitInfo{Votes: []comet.VoteInfo{{
				Validator:   abciVal,
				BlockIDFlag: comet.BlockIDFlagAbsent,
			}}},
		})

		err = slashing.BeginBlocker(ctx, slashingKeeper)
		require.NoError(t, err)
	}

	// end block
	_, err = stakingKeeper.EndBlocker(ctx)
	require.NoError(t, err)

	// validator should be jailed
	validator, err := stakingKeeper.GetValidatorByConsAddr(ctx, sdk.GetConsAddress(pk))
	require.NoError(t, err)
	require.Equal(t, sdk.Unbonding, validator.GetStatus())
}
