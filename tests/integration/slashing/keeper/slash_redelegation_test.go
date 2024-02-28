package keeper_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"cosmossdk.io/core/header"
	"cosmossdk.io/depinject"
	"cosmossdk.io/log"
	"cosmossdk.io/math"
	authkeeper "cosmossdk.io/x/auth/keeper"
	bankkeeper "cosmossdk.io/x/bank/keeper"
	banktestutil "cosmossdk.io/x/bank/testutil"
	distributionkeeper "cosmossdk.io/x/distribution/keeper"
	slashingkeeper "cosmossdk.io/x/slashing/keeper"
	stakingkeeper "cosmossdk.io/x/staking/keeper"
	stakingtypes "cosmossdk.io/x/staking/types"

	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	"github.com/cosmos/cosmos-sdk/tests/integration/slashing"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func TestSlashRedelegation(t *testing.T) {
	// setting up
	var (
		authKeeper    authkeeper.AccountKeeper
		stakingKeeper *stakingkeeper.Keeper
		bankKeeper    bankkeeper.Keeper
		slashKeeper   slashingkeeper.Keeper
		distrKeeper   distributionkeeper.Keeper
	)

	app, err := simtestutil.Setup(
		depinject.Configs(
			depinject.Supply(log.NewNopLogger()),
			slashing.AppConfig,
		),
		&stakingKeeper,
		&bankKeeper,
		&slashKeeper,
		&distrKeeper,
		&authKeeper,
	)
	require.NoError(t, err)

	// get sdk context, staking msg server and bond denom
	ctx := app.BaseApp.NewContext(false)
	stakingMsgServer := stakingkeeper.NewMsgServerImpl(stakingKeeper)
	bondDenom, err := stakingKeeper.BondDenom(ctx)
	require.NoError(t, err)

	// evilVal will be slashed, goodVal won't be slashed
	evilValPubKey := secp256k1.GenPrivKey().PubKey()
	goodValPubKey := secp256k1.GenPrivKey().PubKey()

	// both test acc 1 and 2 delegated to evil val, both acc should be slashed when evil val is slashed
	// test acc 1 use the "undelegation after redelegation" trick (redelegate to good val and then undelegate) to avoid slashing
	// test acc 2 only undelegate from evil val
	testAcc1 := sdk.AccAddress([]byte("addr1_______________"))
	testAcc2 := sdk.AccAddress([]byte("addr2_______________"))

	// fund acc 1 and acc 2
	testCoins := sdk.NewCoins(sdk.NewCoin(bondDenom, stakingKeeper.TokensFromConsensusPower(ctx, 10)))
	fundAccount(t, ctx, bankKeeper, authKeeper, testAcc1, testCoins)
	fundAccount(t, ctx, bankKeeper, authKeeper, testAcc2, testCoins)

	balance1Before := bankKeeper.GetBalance(ctx, testAcc1, bondDenom)
	balance2Before := bankKeeper.GetBalance(ctx, testAcc2, bondDenom)

	// assert acc 1 and acc 2 balance
	require.Equal(t, balance1Before.Amount.String(), testCoins[0].Amount.String())
	require.Equal(t, balance2Before.Amount.String(), testCoins[0].Amount.String())

	// creating evil val
	evilValAddr := sdk.ValAddress(evilValPubKey.Address())
	fundAccount(t, ctx, bankKeeper, authKeeper, sdk.AccAddress(evilValAddr), testCoins)
	createValMsg1, _ := stakingtypes.NewMsgCreateValidator(
		evilValAddr.String(), evilValPubKey, testCoins[0], stakingtypes.Description{Details: "test"}, stakingtypes.NewCommissionRates(math.LegacyNewDecWithPrec(5, 1), math.LegacyNewDecWithPrec(5, 1), math.LegacyNewDec(0)), math.OneInt())
	_, err = stakingMsgServer.CreateValidator(ctx, createValMsg1)
	require.NoError(t, err)

	// creating good val
	goodValAddr := sdk.ValAddress(goodValPubKey.Address())
	fundAccount(t, ctx, bankKeeper, authKeeper, sdk.AccAddress(goodValAddr), testCoins)
	createValMsg2, _ := stakingtypes.NewMsgCreateValidator(
		goodValAddr.String(), goodValPubKey, testCoins[0], stakingtypes.Description{Details: "test"}, stakingtypes.NewCommissionRates(math.LegacyNewDecWithPrec(5, 1), math.LegacyNewDecWithPrec(5, 1), math.LegacyNewDec(0)), math.OneInt())
	_, err = stakingMsgServer.CreateValidator(ctx, createValMsg2)
	require.NoError(t, err)

	ctx = ctx.WithBlockHeight(1).WithHeaderInfo(header.Info{Height: 1})
	// next block, commit height 1, move to height 2
	// acc 1 and acc 2 delegate to evil val
	ctx, err = simtestutil.NextBlock(app, ctx, time.Duration(1))
	require.NoError(t, err)

	// Acc 2 delegate
	delMsg := stakingtypes.NewMsgDelegate(testAcc2.String(), evilValAddr.String(), testCoins[0])
	_, err = stakingMsgServer.Delegate(ctx, delMsg)
	require.NoError(t, err)

	// Acc 1 delegate
	delMsg = stakingtypes.NewMsgDelegate(testAcc1.String(), evilValAddr.String(), testCoins[0])
	_, err = stakingMsgServer.Delegate(ctx, delMsg)
	require.NoError(t, err)

	// next block, commit height 2, move to height 3
	// with the new delegations, evil val increases in voting power and commit byzantine behavior at height 3 consensus
	// at the same time, acc 1 and acc 2 withdraw delegation from evil val
	ctx, err = simtestutil.NextBlock(app, ctx, time.Duration(1))
	require.NoError(t, err)

	evilVal, err := stakingKeeper.GetValidator(ctx, evilValAddr)
	require.NoError(t, err)

	evilPower := stakingKeeper.TokensToConsensusPower(ctx, evilVal.Tokens)
	fmt.Println(evilPower)

	// Acc 1 redelegate from evil val to good val
	redelMsg := stakingtypes.NewMsgBeginRedelegate(testAcc1.String(), evilValAddr.String(), goodValAddr.String(), testCoins[0])
	_, err = stakingMsgServer.BeginRedelegate(ctx, redelMsg)
	require.NoError(t, err)

	// Acc 1 undelegate from good val
	undelMsg := stakingtypes.NewMsgUndelegate(testAcc1.String(), goodValAddr.String(), testCoins[0])
	_, err = stakingMsgServer.Undelegate(ctx, undelMsg)
	require.NoError(t, err)

	// Acc 2 undelegate from evil val
	undelMsg = stakingtypes.NewMsgUndelegate(testAcc2.String(), evilValAddr.String(), testCoins[0])
	_, err = stakingMsgServer.Undelegate(ctx, undelMsg)
	require.NoError(t, err)

	// next block, commit height 3, move to height 4
	// Slash evil val for byzantine behavior at height 3 consensus,
	// at which acc 1 and acc 2 still contributed to evil val voting power
	// even tho they undelegate at block 3, the valset update is applied after committed block 3 when height 3 consensus already passes
	ctx, err = simtestutil.NextBlock(app, ctx, time.Duration(1))
	require.NoError(t, err)

	// slash evil val with slash factor = 0.9, leaving only 10% of stake after slashing
	evilVal, _ = stakingKeeper.GetValidator(ctx, evilValAddr)
	evilValConsAddr, err := evilVal.GetConsAddr()
	require.NoError(t, err)

	err = slashKeeper.Slash(ctx, evilValConsAddr, math.LegacyMustNewDecFromStr("0.9"), evilPower, 3)
	require.NoError(t, err)

	// assert invariant to make sure we conduct slashing correctly
	_, stop := stakingkeeper.AllInvariants(stakingKeeper)(ctx)
	require.False(t, stop)

	_, stop = bankkeeper.AllInvariants(bankKeeper)(ctx)
	require.False(t, stop)

	_, stop = distributionkeeper.AllInvariants(distrKeeper)(ctx)
	require.False(t, stop)

	// one eternity later
	ctx, err = simtestutil.NextBlock(app, ctx, time.Duration(1000000000000000000))
	require.NoError(t, err)
	ctx, err = simtestutil.NextBlock(app, ctx, time.Duration(1))
	require.NoError(t, err)

	// confirm that account 1 and account 2 has been slashed, and the slash amount is correct
	balance1AfterSlashing := bankKeeper.GetBalance(ctx, testAcc1, bondDenom)
	balance2AfterSlashing := bankKeeper.GetBalance(ctx, testAcc2, bondDenom)

	require.Equal(t, balance1AfterSlashing.Amount.Mul(math.NewIntFromUint64(10)).String(), balance1Before.Amount.String())
	require.Equal(t, balance2AfterSlashing.Amount.Mul(math.NewIntFromUint64(10)).String(), balance2Before.Amount.String())
}

func fundAccount(t *testing.T, ctx context.Context, bankKeeper bankkeeper.Keeper, authKeeper authkeeper.AccountKeeper, addr sdk.AccAddress, amount sdk.Coins) {
	t.Helper()

	if authKeeper.GetAccount(ctx, addr) == nil {
		addrAcc := authKeeper.NewAccountWithAddress(ctx, addr)
		authKeeper.SetAccount(ctx, addrAcc)
	}

	require.NoError(t, banktestutil.FundAccount(ctx, bankKeeper, addr, amount))
}
