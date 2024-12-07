package slashing

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"cosmossdk.io/core/comet"
	coreheader "cosmossdk.io/core/header"
	"cosmossdk.io/runtime/v2/services"
	"cosmossdk.io/x/slashing"
	"cosmossdk.io/x/slashing/testutil"
	stakingtestutil "cosmossdk.io/x/staking/testutil"

	"github.com/cosmos/cosmos-sdk/tests/integration/v2"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// TestBeginBlocker is a unit test function that tests the behavior of the BeginBlocker function.
// It sets up the necessary dependencies and context, creates a validator, and performs various operations
// to test the slashing logic. It checks if the validator is correctly jailed after a certain number of blocks.
func TestBeginBlocker(t *testing.T) {
	f := initFixture(t)

	ctx := f.ctx

	pks := simtestutil.CreateTestPubKeys(1)
	simtestutil.AddTestAddrsFromPubKeys(f.bankKeeper, f.stakingKeeper, ctx, pks, f.stakingKeeper.TokensFromConsensusPower(ctx, 200))
	addr, pk := sdk.ValAddress(pks[0].Address()), pks[0]
	tstaking := stakingtestutil.NewHelper(t, ctx, f.stakingKeeper)

	// bond the validator
	power := int64(100)
	acc := f.accountKeeper.NewAccountWithAddress(ctx, sdk.AccAddress(addr))
	f.accountKeeper.SetAccount(ctx, acc)
	amt := tstaking.CreateValidatorWithValPower(addr, pk, power, true)
	_, err := f.stakingKeeper.EndBlocker(ctx)
	require.NoError(t, err)
	bondDenom, err := f.stakingKeeper.BondDenom(ctx)
	require.NoError(t, err)
	require.Equal(
		t, f.bankKeeper.GetAllBalances(ctx, sdk.AccAddress(addr)),
		sdk.NewCoins(sdk.NewCoin(bondDenom, testutil.InitTokens.Sub(amt))),
	)
	val, err := f.stakingKeeper.Validator(ctx, addr)
	require.NoError(t, err)
	require.Equal(t, amt, val.GetBondedTokens())

	abciVal := comet.Validator{
		Address: pk.Address(),
		Power:   power,
	}

	ctx = integration.SetCometInfo(ctx, comet.Info{
		LastCommit: comet.CommitInfo{Votes: []comet.VoteInfo{{
			Validator:   abciVal,
			BlockIDFlag: comet.BlockIDFlagCommit,
		}}},
	})
	cometInfoService := &services.ContextAwareCometInfoService{}

	err = slashing.BeginBlocker(ctx, f.slashingKeeper, cometInfoService)
	require.NoError(t, err)

	info, err := f.slashingKeeper.ValidatorSigningInfo.Get(ctx, sdk.ConsAddress(pk.Address()))
	require.NoError(t, err)
	require.Equal(t, integration.HeaderInfoFromContext(ctx).Height, info.StartHeight)
	require.Equal(t, int64(0), info.IndexOffset)
	require.Equal(t, time.Unix(0, 0).UTC(), info.JailedUntil)
	require.Equal(t, int64(0), info.MissedBlocksCounter)

	height := int64(0)

	signedBlocksWindow, err := f.slashingKeeper.SignedBlocksWindow(ctx)
	require.NoError(t, err)
	// for 100 blocks, mark the validator as having signed
	for ; height < signedBlocksWindow; height++ {
		ctx = integration.SetHeaderInfo(ctx, coreheader.Info{Height: height})

		err = slashing.BeginBlocker(ctx, f.slashingKeeper, cometInfoService)
		require.NoError(t, err)
	}

	minSignedPerWindow, err := f.slashingKeeper.MinSignedPerWindow(ctx)
	require.NoError(t, err)
	// for 50 blocks, mark the validator as having not signed
	for ; height < ((signedBlocksWindow * 2) - minSignedPerWindow + 1); height++ {
		ctx = integration.SetHeaderInfo(ctx, coreheader.Info{Height: height})
		ctx = integration.SetCometInfo(ctx, comet.Info{
			LastCommit: comet.CommitInfo{Votes: []comet.VoteInfo{{
				Validator:   abciVal,
				BlockIDFlag: comet.BlockIDFlagAbsent,
			}}},
		})

		err = slashing.BeginBlocker(ctx, f.slashingKeeper, cometInfoService)
		require.NoError(t, err)
	}

	// end block
	_, err = f.stakingKeeper.EndBlocker(ctx)
	require.NoError(t, err)

	// validator should be jailed
	validator, err := f.stakingKeeper.GetValidatorByConsAddr(ctx, sdk.GetConsAddress(pk))
	require.NoError(t, err)
	require.Equal(t, sdk.Unbonding, validator.GetStatus())
}
