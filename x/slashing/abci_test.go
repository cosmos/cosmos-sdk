package slashing_test

import (
	"testing"
	"time"

	cmtproto "github.com/cometbft/cometbft/api/cometbft/types/v2"
	abci "github.com/cometbft/cometbft/v2/abci/types"
	"github.com/stretchr/testify/require"

	"cosmossdk.io/depinject"
	"cosmossdk.io/log"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	sdk "github.com/cosmos/cosmos-sdk/types"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	"github.com/cosmos/cosmos-sdk/x/slashing"
	slashingkeeper "github.com/cosmos/cosmos-sdk/x/slashing/keeper"
	"github.com/cosmos/cosmos-sdk/x/slashing/testutil"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	stakingtestutil "github.com/cosmos/cosmos-sdk/x/staking/testutil"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

func TestBeginBlocker(t *testing.T) {
	var (
		interfaceRegistry codectypes.InterfaceRegistry
		bankKeeper        bankkeeper.Keeper
		stakingKeeper     *stakingkeeper.Keeper
		slashingKeeper    slashingkeeper.Keeper
	)

	app, err := simtestutil.Setup(
		depinject.Configs(
			testutil.AppConfig,
			depinject.Supply(log.NewNopLogger()),
		),
		&interfaceRegistry,
		&bankKeeper,
		&stakingKeeper,
		&slashingKeeper,
	)
	require.NoError(t, err)

	ctx := app.NewContext(false)

	pks := simtestutil.CreateTestPubKeys(1)
	simtestutil.AddTestAddrsFromPubKeys(bankKeeper, stakingKeeper, ctx, pks, stakingKeeper.TokensFromConsensusPower(ctx, 200))
	addr, pk := sdk.ValAddress(pks[0].Address()), pks[0]
	tstaking := stakingtestutil.NewHelper(t, ctx, stakingKeeper)

	// bond the validator
	power := int64(100)
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

	abciVal := abci.Validator{
		Address: pk.Address(),
		Power:   power,
	}

	ctx = ctx.WithVoteInfos([]abci.VoteInfo{{
		Validator:   abciVal,
		BlockIdFlag: cmtproto.BlockIDFlagCommit,
	}})

	err = slashing.BeginBlocker(ctx, slashingKeeper)
	require.NoError(t, err)

	info, err := slashingKeeper.GetValidatorSigningInfo(ctx, sdk.ConsAddress(pk.Address()))
	require.NoError(t, err)
	require.Equal(t, ctx.BlockHeight(), info.StartHeight)
	require.Equal(t, int64(1), info.IndexOffset)
	require.Equal(t, time.Unix(0, 0).UTC(), info.JailedUntil)
	require.Equal(t, int64(0), info.MissedBlocksCounter)

	height := int64(0)

	signedBlocksWindow, err := slashingKeeper.SignedBlocksWindow(ctx)
	require.NoError(t, err)
	// for 100 blocks, mark the validator as having signed
	for ; height < signedBlocksWindow; height++ {
		ctx = ctx.WithBlockHeight(height).
			WithVoteInfos([]abci.VoteInfo{{
				Validator:   abciVal,
				BlockIdFlag: cmtproto.BlockIDFlagCommit,
			}})

		err = slashing.BeginBlocker(ctx, slashingKeeper)
		require.NoError(t, err)
	}

	minSignedPerWindow, err := slashingKeeper.MinSignedPerWindow(ctx)
	require.NoError(t, err)
	// for 50 blocks, mark the validator as having not signed
	for ; height < ((signedBlocksWindow * 2) - minSignedPerWindow + 1); height++ {
		ctx = ctx.WithBlockHeight(height).
			WithVoteInfos([]abci.VoteInfo{{
				Validator:   abciVal,
				BlockIdFlag: cmtproto.BlockIDFlagAbsent,
			}})

		err = slashing.BeginBlocker(ctx, slashingKeeper)
		require.NoError(t, err)
	}

	// end block
	_, err = stakingKeeper.EndBlocker(ctx)
	require.NoError(t, err)

	// validator should be jailed
	validator, err := stakingKeeper.GetValidatorByConsAddr(ctx, sdk.GetConsAddress(pk))
	require.NoError(t, err)
	require.Equal(t, stakingtypes.Unbonding, validator.GetStatus())
}
