package slashing_test

import (
	"testing"
	"time"

	"cosmossdk.io/depinject"
	"cosmossdk.io/log"
	abci "github.com/cometbft/cometbft/abci/types"
	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	"github.com/stretchr/testify/require"

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
	var interfaceRegistry codectypes.InterfaceRegistry
	var bankKeeper bankkeeper.Keeper
	var stakingKeeper *stakingkeeper.Keeper
	var slashingKeeper slashingkeeper.Keeper

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

	ctx := app.BaseApp.NewContext(false, cmtproto.Header{})

	pks := simtestutil.CreateTestPubKeys(1)
	simtestutil.AddTestAddrsFromPubKeys(bankKeeper, stakingKeeper, ctx, pks, stakingKeeper.TokensFromConsensusPower(ctx, 200))
	addr, pk := sdk.ValAddress(pks[0].Address()), pks[0]
	tstaking := stakingtestutil.NewHelper(t, ctx, stakingKeeper)

	// bond the validator
	power := int64(100)
	amt := tstaking.CreateValidatorWithValPower(addr, pk, power, true)
	stakingKeeper.EndBlocker(ctx)
	require.Equal(
		t, bankKeeper.GetAllBalances(ctx, sdk.AccAddress(addr)),
		sdk.NewCoins(sdk.NewCoin(stakingKeeper.GetParams(ctx).BondDenom, testutil.InitTokens.Sub(amt))),
	)
	require.Equal(t, amt, stakingKeeper.Validator(ctx, addr).GetBondedTokens())

	val := abci.Validator{
		Address: pk.Address(),
		Power:   power,
	}

	ctx = ctx.WithVoteInfos([]abci.VoteInfo{{
		Validator:   val,
		BlockIdFlag: cmtproto.BlockIDFlagCommit,
	}})

	slashing.BeginBlocker(ctx, slashingKeeper)

	info, found := slashingKeeper.GetValidatorSigningInfo(ctx, sdk.ConsAddress(pk.Address()))
	require.True(t, found)
	require.Equal(t, ctx.BlockHeight(), info.StartHeight)
	require.Equal(t, int64(1), info.IndexOffset)
	require.Equal(t, time.Unix(0, 0).UTC(), info.JailedUntil)
	require.Equal(t, int64(0), info.MissedBlocksCounter)

	height := int64(0)

	// for 100 blocks, mark the validator as having signed
	for ; height < slashingKeeper.SignedBlocksWindow(ctx); height++ {
		ctx = ctx.WithBlockHeight(height).
			WithVoteInfos([]abci.VoteInfo{{
				Validator:   val,
				BlockIdFlag: cmtproto.BlockIDFlagCommit,
			}})

		slashing.BeginBlocker(ctx, slashingKeeper)
	}

	// for 50 blocks, mark the validator as having not signed
	for ; height < ((slashingKeeper.SignedBlocksWindow(ctx) * 2) - slashingKeeper.MinSignedPerWindow(ctx) + 1); height++ {
		ctx = ctx.WithBlockHeight(height).
			WithVoteInfos([]abci.VoteInfo{{
				Validator:   val,
				BlockIdFlag: cmtproto.BlockIDFlagAbsent,
			}})

		slashing.BeginBlocker(ctx, slashingKeeper)
	}

	// end block
	stakingKeeper.EndBlocker(ctx)

	// validator should be jailed
	validator, found := stakingKeeper.GetValidatorByConsAddr(ctx, sdk.GetConsAddress(pk))
	require.True(t, found)
	require.Equal(t, stakingtypes.Unbonding, validator.GetStatus())
}
