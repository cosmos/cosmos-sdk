package keeper_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"cosmossdk.io/core/header"
	"cosmossdk.io/depinject"
	"cosmossdk.io/log"
	"cosmossdk.io/math"
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
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
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
	testCoin := sdk.NewCoin(bondDenom, stakingKeeper.TokensFromConsensusPower(ctx, 10))
	fundAccount(t, ctx, bankKeeper, authKeeper, testAcc1, testCoin)
	fundAccount(t, ctx, bankKeeper, authKeeper, testAcc2, testCoin)

	balance1Before := bankKeeper.GetBalance(ctx, testAcc1, bondDenom)
	balance2Before := bankKeeper.GetBalance(ctx, testAcc2, bondDenom)

	// assert acc 1 and acc 2 balance
	require.Equal(t, balance1Before.Amount.String(), testCoin.Amount.String())
	require.Equal(t, balance2Before.Amount.String(), testCoin.Amount.String())

	// creating evil val
	evilValAddr := sdk.ValAddress(evilValPubKey.Address())
	fundAccount(t, ctx, bankKeeper, authKeeper, sdk.AccAddress(evilValAddr), testCoin)
	createValMsg1, _ := stakingtypes.NewMsgCreateValidator(
		evilValAddr.String(), evilValPubKey, testCoin, stakingtypes.Description{Details: "test"}, stakingtypes.NewCommissionRates(math.LegacyNewDecWithPrec(5, 1), math.LegacyNewDecWithPrec(5, 1), math.LegacyNewDec(0)), math.OneInt())
	_, err = stakingMsgServer.CreateValidator(ctx, createValMsg1)
	require.NoError(t, err)

	// creating good val
	goodValAddr := sdk.ValAddress(goodValPubKey.Address())
	fundAccount(t, ctx, bankKeeper, authKeeper, sdk.AccAddress(goodValAddr), testCoin)
	createValMsg2, _ := stakingtypes.NewMsgCreateValidator(
		goodValAddr.String(), goodValPubKey, testCoin, stakingtypes.Description{Details: "test"}, stakingtypes.NewCommissionRates(math.LegacyNewDecWithPrec(5, 1), math.LegacyNewDecWithPrec(5, 1), math.LegacyNewDec(0)), math.OneInt())
	_, err = stakingMsgServer.CreateValidator(ctx, createValMsg2)
	require.NoError(t, err)

	ctx = ctx.WithBlockHeight(1).WithHeaderInfo(header.Info{Height: 1})
	// next block, commit height 1, move to height 2
	// acc 1 and acc 2 delegate to evil val
	ctx, err = simtestutil.NextBlock(app, ctx, time.Duration(1))
	require.NoError(t, err)

	// Acc 2 delegate
	delMsg := stakingtypes.NewMsgDelegate(testAcc2.String(), evilValAddr.String(), testCoin)
	_, err = stakingMsgServer.Delegate(ctx, delMsg)
	require.NoError(t, err)

	// Acc 1 delegate
	delMsg = stakingtypes.NewMsgDelegate(testAcc1.String(), evilValAddr.String(), testCoin)
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

	// Acc 1 redelegate from evil val to good val
	redelMsg := stakingtypes.NewMsgBeginRedelegate(testAcc1.String(), evilValAddr.String(), goodValAddr.String(), testCoin)
	_, err = stakingMsgServer.BeginRedelegate(ctx, redelMsg)
	require.NoError(t, err)

	// Acc 1 undelegate from good val
	undelMsg := stakingtypes.NewMsgUndelegate(testAcc1.String(), goodValAddr.String(), testCoin)
	_, err = stakingMsgServer.Undelegate(ctx, undelMsg)
	require.NoError(t, err)

	// Acc 2 undelegate from evil val
	undelMsg = stakingtypes.NewMsgUndelegate(testAcc2.String(), evilValAddr.String(), testCoin)
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

func fundAccount(t *testing.T, ctx context.Context, bankKeeper bankkeeper.Keeper, authKeeper authkeeper.AccountKeeper, addr sdk.AccAddress, amount ...sdk.Coin) {
	t.Helper()

	if authKeeper.GetAccount(ctx, addr) == nil {
		addrAcc := authKeeper.NewAccountWithAddress(ctx, addr)
		authKeeper.SetAccount(ctx, addrAcc)
	}

	require.NoError(t, banktestutil.FundAccount(ctx, bankKeeper, addr, amount))
}

func TestOverSlashing(t *testing.T) {
	// slash penalty percentage
	slashFraction := "0.45"

	// percentage of (undelegation/(undelegation + redelegation))
	undelegationPercentageStr := "0.30"

	// setting up
	var (
		stakingKeeper *stakingkeeper.Keeper
		bankKeeper    bankkeeper.Keeper
		slashKeeper   slashingkeeper.Keeper
		distrKeeper   distributionkeeper.Keeper
		authKeeper    authkeeper.AccountKeeper
	)

	app, err := simtestutil.Setup(depinject.Configs(
		depinject.Supply(log.NewNopLogger()),
		slashing.AppConfig,
	), &stakingKeeper, &bankKeeper, &slashKeeper, &distrKeeper, &authKeeper)
	require.NoError(t, err)

	// get sdk context, staking msg server and bond denom
	ctx := app.BaseApp.NewContext(false)
	stakingMsgServer := stakingkeeper.NewMsgServerImpl(stakingKeeper)
	bondDenom, err := stakingKeeper.BondDenom(ctx)
	require.NoError(t, err)

	// evilVal will be slashed, goodVal won't be slashed
	evilValPubKey := secp256k1.GenPrivKey().PubKey()
	goodValPubKey := secp256k1.GenPrivKey().PubKey()

	/*
	   all test accs will delegate to evil val, which evil validator will eventually be slashed

	   - test acc 1: redelegate -> undelegate full amount
	   - test acc 2: simple undelegation. intended scenario.
	   - test acc 3: redelegate -> undelegate some amount

	*/

	testAcc1 := sdk.AccAddress([]byte("addr1new____________"))
	testAcc2 := sdk.AccAddress([]byte("addr2new____________"))
	testAcc3 := sdk.AccAddress([]byte("addr3new____________"))

	// fund all accounts
	testCoin := sdk.NewCoin(bondDenom, math.NewInt(1_000_000))
	fundAccount(t, ctx, bankKeeper, authKeeper, testAcc1, testCoin)
	fundAccount(t, ctx, bankKeeper, authKeeper, testAcc2, testCoin)
	fundAccount(t, ctx, bankKeeper, authKeeper, testAcc3, testCoin)

	balance1Before := bankKeeper.GetBalance(ctx, testAcc1, bondDenom)
	balance2Before := bankKeeper.GetBalance(ctx, testAcc2, bondDenom)
	balance3Before := bankKeeper.GetBalance(ctx, testAcc3, bondDenom)

	// assert acc 1, 2 and 3 balance
	require.Equal(t, testCoin.Amount.String(), balance1Before.Amount.String())
	require.Equal(t, testCoin.Amount.String(), balance2Before.Amount.String())
	require.Equal(t, testCoin.Amount.String(), balance3Before.Amount.String())

	// create evil val
	evilValAddr := sdk.ValAddress(evilValPubKey.Address())
	fundAccount(t, ctx, bankKeeper, authKeeper, sdk.AccAddress(evilValAddr), testCoin)
	createValMsg1, _ := stakingtypes.NewMsgCreateValidator(
		evilValAddr.String(), evilValPubKey, testCoin, stakingtypes.Description{Details: "test"}, stakingtypes.NewCommissionRates(math.LegacyNewDecWithPrec(5, 1), math.LegacyNewDecWithPrec(5, 1), math.LegacyNewDec(0)), math.OneInt())
	_, err = stakingMsgServer.CreateValidator(ctx, createValMsg1)
	require.NoError(t, err)

	// create good val 1
	goodValAddr := sdk.ValAddress(goodValPubKey.Address())
	fundAccount(t, ctx, bankKeeper, authKeeper, sdk.AccAddress(goodValAddr), testCoin)
	createValMsg2, _ := stakingtypes.NewMsgCreateValidator(
		goodValAddr.String(), goodValPubKey, testCoin, stakingtypes.Description{Details: "test"}, stakingtypes.NewCommissionRates(math.LegacyNewDecWithPrec(5, 1), math.LegacyNewDecWithPrec(5, 1), math.LegacyNewDec(0)), math.OneInt())
	_, err = stakingMsgServer.CreateValidator(ctx, createValMsg2)
	require.NoError(t, err)

	// next block
	ctx = ctx.WithBlockHeight(app.LastBlockHeight() + 1).WithHeaderInfo(header.Info{Height: app.LastBlockHeight() + 1})
	ctx, err = simtestutil.NextBlock(app, ctx, time.Duration(1))
	require.NoError(t, err)

	// delegate all accs to evil val
	delMsg := stakingtypes.NewMsgDelegate(testAcc1.String(), evilValAddr.String(), testCoin)
	_, err = stakingMsgServer.Delegate(ctx, delMsg)
	require.NoError(t, err)

	delMsg = stakingtypes.NewMsgDelegate(testAcc2.String(), evilValAddr.String(), testCoin)
	_, err = stakingMsgServer.Delegate(ctx, delMsg)
	require.NoError(t, err)

	delMsg = stakingtypes.NewMsgDelegate(testAcc3.String(), evilValAddr.String(), testCoin)
	_, err = stakingMsgServer.Delegate(ctx, delMsg)
	require.NoError(t, err)

	// next block
	ctx, err = simtestutil.NextBlock(app, ctx, time.Duration(1))
	require.NoError(t, err)

	// evilValAddr done something bad
	misbehaveHeight := ctx.BlockHeader().Height
	evilVal, err := stakingKeeper.GetValidator(ctx, evilValAddr)
	require.NoError(t, err)

	evilValConsAddr, err := evilVal.GetConsAddr()
	require.NoError(t, err)

	evilPower := stakingKeeper.TokensToConsensusPower(ctx, evilVal.Tokens)

	// next block
	ctx, err = simtestutil.NextBlock(app, ctx, time.Duration(1))
	require.NoError(t, err)

	// acc 1: redelegate to goodval1 and undelegate FULL amount
	redelMsg := stakingtypes.NewMsgBeginRedelegate(testAcc1.String(), evilValAddr.String(), goodValAddr.String(), testCoin)
	_, err = stakingMsgServer.BeginRedelegate(ctx, redelMsg)
	require.NoError(t, err)
	undelMsg := stakingtypes.NewMsgUndelegate(testAcc1.String(), goodValAddr.String(), testCoin)
	_, err = stakingMsgServer.Undelegate(ctx, undelMsg)
	require.NoError(t, err)

	// acc 2: undelegate full amount
	undelMsg = stakingtypes.NewMsgUndelegate(testAcc2.String(), evilValAddr.String(), testCoin)
	_, err = stakingMsgServer.Undelegate(ctx, undelMsg)
	require.NoError(t, err)

	// acc 3: redelegate to goodval1 and undelegate some amount
	redelMsg = stakingtypes.NewMsgBeginRedelegate(testAcc3.String(), evilValAddr.String(), goodValAddr.String(), testCoin)
	_, err = stakingMsgServer.BeginRedelegate(ctx, redelMsg)
	require.NoError(t, err)

	undelegationPercentage := math.LegacyMustNewDecFromStr(undelegationPercentageStr)
	undelegationAmountDec := math.LegacyNewDecFromInt(testCoin.Amount).Mul(undelegationPercentage)
	amountToUndelegate := undelegationAmountDec.TruncateInt()

	// next block
	ctx, err = simtestutil.NextBlock(app, ctx, time.Duration(1))
	require.NoError(t, err)

	portionofTestCoins := sdk.NewCoin(bondDenom, amountToUndelegate)
	undelMsg = stakingtypes.NewMsgUndelegate(testAcc3.String(), goodValAddr.String(), portionofTestCoins)
	_, err = stakingMsgServer.Undelegate(ctx, undelMsg)
	require.NoError(t, err)

	// next block
	ctx, err = simtestutil.NextBlock(app, ctx, time.Duration(1))
	require.NoError(t, err)

	// slash the evil val
	err = slashKeeper.Slash(ctx, evilValConsAddr, math.LegacyMustNewDecFromStr(slashFraction), evilPower, misbehaveHeight)
	require.NoError(t, err)

	// assert invariants
	_, stop := stakingkeeper.AllInvariants(stakingKeeper)(ctx)
	require.False(t, stop)
	_, stop = bankkeeper.AllInvariants(bankKeeper)(ctx)
	require.False(t, stop)
	_, stop = distributionkeeper.AllInvariants(distrKeeper)(ctx)
	require.False(t, stop)

	// fastforward 2 blocks to complete redelegations and unbondings
	for i := 0; i < 2; i++ {
		ctx, err = simtestutil.NextBlock(app, ctx, time.Duration(1000000000000000000))
		require.NoError(t, err)
	}

	// we check all accounts should be slashed with the equal amount, and they should end up with same balance including staked amount
	stakedAcc1, err := stakingKeeper.GetDelegatorBonded(ctx, testAcc1)
	require.NoError(t, err)
	stakedAcc2, err := stakingKeeper.GetDelegatorBonded(ctx, testAcc2)
	require.NoError(t, err)
	stakedAcc3, err := stakingKeeper.GetDelegatorBonded(ctx, testAcc3)
	require.NoError(t, err)

	balance1AfterSlashing := bankKeeper.GetBalance(ctx, testAcc1, bondDenom).Add(sdk.NewCoin(bondDenom, stakedAcc1))
	balance2AfterSlashing := bankKeeper.GetBalance(ctx, testAcc2, bondDenom).Add(sdk.NewCoin(bondDenom, stakedAcc2))
	balance3AfterSlashing := bankKeeper.GetBalance(ctx, testAcc3, bondDenom).Add(sdk.NewCoin(bondDenom, stakedAcc3))

	require.Equal(t, "550000stake", balance1AfterSlashing.String())
	require.Equal(t, "550000stake", balance2AfterSlashing.String())
	require.Equal(t, "550000stake", balance3AfterSlashing.String())
}

func TestSlashRedelegation_ValidatorLeftWithNoTokens(t *testing.T) {
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
	ctx := app.BaseApp.NewContext(false).WithBlockHeight(1).WithHeaderInfo(header.Info{Height: 1})
	stakingMsgServer := stakingkeeper.NewMsgServerImpl(stakingKeeper)
	bondDenom, err := stakingKeeper.BondDenom(ctx)
	require.NoError(t, err)

	// create validators DST and SRC
	dstPubKey := secp256k1.GenPrivKey().PubKey()
	srcPubKey := secp256k1.GenPrivKey().PubKey()

	dstAddr := sdk.ValAddress(dstPubKey.Address())
	srcAddr := sdk.ValAddress(srcPubKey.Address())

	testCoin := sdk.NewCoin(bondDenom, stakingKeeper.TokensFromConsensusPower(ctx, 1000))
	fundAccount(t, ctx, bankKeeper, authKeeper, sdk.AccAddress(dstAddr), testCoin)
	fundAccount(t, ctx, bankKeeper, authKeeper, sdk.AccAddress(srcAddr), testCoin)

	createValMsgDST, _ := stakingtypes.NewMsgCreateValidator(
		dstAddr.String(), dstPubKey, testCoin, stakingtypes.Description{Details: "Validator DST"}, stakingtypes.NewCommissionRates(math.LegacyNewDecWithPrec(5, 1), math.LegacyNewDecWithPrec(5, 1), math.LegacyNewDec(0)), math.OneInt())
	_, err = stakingMsgServer.CreateValidator(ctx, createValMsgDST)
	require.NoError(t, err)

	createValMsgSRC, _ := stakingtypes.NewMsgCreateValidator(
		srcAddr.String(), srcPubKey, testCoin, stakingtypes.Description{Details: "Validator SRC"}, stakingtypes.NewCommissionRates(math.LegacyNewDecWithPrec(5, 1), math.LegacyNewDecWithPrec(5, 1), math.LegacyNewDec(0)), math.OneInt())
	_, err = stakingMsgServer.CreateValidator(ctx, createValMsgSRC)
	require.NoError(t, err)

	// create a user accounts and delegate to SRC and DST
	userAcc := sdk.AccAddress([]byte("user1_______________"))
	fundAccount(t, ctx, bankKeeper, authKeeper, userAcc, testCoin)

	userAcc2 := sdk.AccAddress([]byte("user2_______________"))
	fundAccount(t, ctx, bankKeeper, authKeeper, userAcc2, testCoin)

	delMsg := stakingtypes.NewMsgDelegate(userAcc.String(), srcAddr.String(), testCoin)
	_, err = stakingMsgServer.Delegate(ctx, delMsg)
	require.NoError(t, err)

	delMsg = stakingtypes.NewMsgDelegate(userAcc2.String(), dstAddr.String(), testCoin)
	_, err = stakingMsgServer.Delegate(ctx, delMsg)
	require.NoError(t, err)

	ctx, err = simtestutil.NextBlock(app, ctx, time.Duration(1))
	require.NoError(t, err)

	// commit an infraction with DST and store the power at this height
	dstVal, err := stakingKeeper.GetValidator(ctx, dstAddr)
	require.NoError(t, err)
	dstPower := stakingKeeper.TokensToConsensusPower(ctx, dstVal.Tokens)
	dstConsAddr, err := dstVal.GetConsAddr()
	require.NoError(t, err)
	dstInfractionHeight := ctx.BlockHeight()

	ctx, err = simtestutil.NextBlock(app, ctx, time.Duration(1))
	require.NoError(t, err)

	// undelegate all the user tokens from DST
	undelMsg := stakingtypes.NewMsgUndelegate(userAcc2.String(), dstAddr.String(), testCoin)
	_, err = stakingMsgServer.Undelegate(ctx, undelMsg)
	require.NoError(t, err)

	// commit an infraction with SRC and store the power at this height
	srcVal, err := stakingKeeper.GetValidator(ctx, srcAddr)
	require.NoError(t, err)
	srcPower := stakingKeeper.TokensToConsensusPower(ctx, srcVal.Tokens)
	srcConsAddr, err := srcVal.GetConsAddr()
	require.NoError(t, err)
	srcInfractionHeight := ctx.BlockHeight()

	ctx, err = simtestutil.NextBlock(app, ctx, time.Duration(1))
	require.NoError(t, err)

	// redelegate all the user tokens from SRC to DST
	redelMsg := stakingtypes.NewMsgBeginRedelegate(userAcc.String(), srcAddr.String(), dstAddr.String(), testCoin)
	_, err = stakingMsgServer.BeginRedelegate(ctx, redelMsg)
	require.NoError(t, err)

	// undelegate the self delegation from DST
	undelMsg = stakingtypes.NewMsgUndelegate(sdk.AccAddress(dstAddr).String(), dstAddr.String(), testCoin)
	_, err = stakingMsgServer.Undelegate(ctx, undelMsg)
	require.NoError(t, err)

	ctx, err = simtestutil.NextBlock(app, ctx, time.Duration(1))
	require.NoError(t, err)

	undelMsg = stakingtypes.NewMsgUndelegate(userAcc.String(), dstAddr.String(), testCoin)
	_, err = stakingMsgServer.Undelegate(ctx, undelMsg)
	require.NoError(t, err)

	// check that dst now has zero tokens
	valDst, err := stakingKeeper.GetValidator(ctx, dstAddr)
	require.NoError(t, err)
	require.Equal(t, math.ZeroInt().String(), valDst.Tokens.String())

	// slash the infractions
	err = slashKeeper.Slash(ctx, dstConsAddr, math.LegacyMustNewDecFromStr("0.8"), dstPower, dstInfractionHeight)
	require.NoError(t, err)

	err = slashKeeper.Slash(ctx, srcConsAddr, math.LegacyMustNewDecFromStr("0.5"), srcPower, srcInfractionHeight)
	require.NoError(t, err)

	// assert invariants to ensure correctness
	_, stop := stakingkeeper.AllInvariants(stakingKeeper)(ctx)
	require.False(t, stop)

	_, stop = bankkeeper.AllInvariants(bankKeeper)(ctx)
	require.False(t, stop)

	_, stop = distributionkeeper.AllInvariants(distrKeeper)(ctx)
	require.False(t, stop)
}
