package keeper_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"cosmossdk.io/core/header"
	"cosmossdk.io/depinject"
	"cosmossdk.io/log"
	"cosmossdk.io/math"

	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	sdk "github.com/cosmos/cosmos-sdk/types"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	banktestutil "github.com/cosmos/cosmos-sdk/x/bank/testutil"
	distributionkeeper "github.com/cosmos/cosmos-sdk/x/distribution/keeper"
	slashingkeeper "github.com/cosmos/cosmos-sdk/x/slashing/keeper"
	"github.com/cosmos/cosmos-sdk/x/slashing/testutil"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

func TestSlashRedelegation(t *testing.T) {
	// setting up
	var stakingKeeper *stakingkeeper.Keeper
	var bankKeeper bankkeeper.Keeper
	var slashKeeper slashingkeeper.Keeper
	var distrKeeper distributionkeeper.Keeper

	app, err := simtestutil.Setup(depinject.Configs(
		depinject.Supply(log.NewNopLogger()),
		testutil.AppConfig,
	), &stakingKeeper, &bankKeeper, &slashKeeper, &distrKeeper)
	require.NoError(t, err)

	// get sdk context, staking msg server and bond denom
	ctx := app.BaseApp.NewContext(false)
	stakingMsgServer := stakingkeeper.NewMsgServerImpl(stakingKeeper)
	bondDenom, err := stakingKeeper.BondDenom(ctx)
	require.NoError(t, err)

	// evilVal will be slashed, goodVal won't be slashed
	evilValPubKey := secp256k1.GenPrivKey().PubKey()
	goodValPubKey := secp256k1.GenPrivKey().PubKey()

	// test acc 1, 2 and 3 delegated to evil val, all acc should be slashed when evil val is slashed
	// test acc 1 use the "undelegation after redelegation" trick (redelegate to good val and then undelegate) to avoid slashing
	// test acc 2 only undelegate from evil val
	// test acc 3 redelegate to good val and stays bonded
	// test acc 4 redelegates liquid staked tokens
	testAcc1 := sdk.AccAddress([]byte("addr1_______________"))
	testAcc2 := sdk.AccAddress([]byte("addr2_______________"))
	testAcc3 := sdk.AccAddress([]byte("addr3_______________"))
	testAcc4 := sdk.AccAddress([]byte("addr4_______________"))

	// fund accounts
	testCoins := sdk.NewCoins(sdk.NewCoin(bondDenom, stakingKeeper.TokensFromConsensusPower(ctx, 10)))
	validatorFundCoins := sdk.NewCoins(sdk.NewCoin(bondDenom, stakingKeeper.TokensFromConsensusPower(ctx, 20)))
	banktestutil.FundAccount(ctx, bankKeeper, testAcc1, testCoins)
	banktestutil.FundAccount(ctx, bankKeeper, testAcc2, testCoins)
	banktestutil.FundAccount(ctx, bankKeeper, testAcc3, testCoins)
	banktestutil.FundAccount(ctx, bankKeeper, testAcc4, testCoins)

	balance1Before := bankKeeper.GetBalance(ctx, testAcc1, bondDenom)
	balance2Before := bankKeeper.GetBalance(ctx, testAcc2, bondDenom)
	balance3Before := bankKeeper.GetBalance(ctx, testAcc3, bondDenom)
	balance4Before := bankKeeper.GetBalance(ctx, testAcc3, bondDenom)

	// assert acc balances
	require.Equal(t, balance1Before.Amount.String(), testCoins[0].Amount.String())
	require.Equal(t, balance2Before.Amount.String(), testCoins[0].Amount.String())
	require.Equal(t, balance3Before.Amount.String(), testCoins[0].Amount.String())
	require.Equal(t, balance4Before.Amount.String(), testCoins[0].Amount.String())

	// creating evil val
	evilValAddr := sdk.ValAddress(evilValPubKey.Address())
	banktestutil.FundAccount(ctx, bankKeeper, sdk.AccAddress(evilValAddr), validatorFundCoins)
	createValMsg1, _ := stakingtypes.NewMsgCreateValidator(
		evilValAddr.String(), evilValPubKey, testCoins[0], stakingtypes.Description{Details: "test"}, stakingtypes.NewCommissionRates(math.LegacyNewDecWithPrec(5, 1), math.LegacyNewDecWithPrec(5, 1), math.LegacyNewDec(0)))
	_, err = stakingMsgServer.CreateValidator(ctx, createValMsg1)
	require.NoError(t, err)

	// creating good val
	goodValAddr := sdk.ValAddress(goodValPubKey.Address())
	goodValAccAddr := sdk.AccAddress(goodValPubKey.Address())
	banktestutil.FundAccount(ctx, bankKeeper, sdk.AccAddress(goodValAddr), validatorFundCoins)
	createValMsg2, _ := stakingtypes.NewMsgCreateValidator(
		goodValAddr.String(), goodValPubKey, testCoins[0], stakingtypes.Description{Details: "test"}, stakingtypes.NewCommissionRates(math.LegacyNewDecWithPrec(5, 1), math.LegacyNewDecWithPrec(5, 1), math.LegacyNewDec(0)))
	_, err = stakingMsgServer.CreateValidator(ctx, createValMsg2)
	require.NoError(t, err)

	// self bond tokens
	bondValMsg := stakingtypes.NewMsgValidatorBond(goodValAccAddr.String(), goodValAddr.String())
	_, err = stakingMsgServer.ValidatorBond(ctx, bondValMsg)
	require.NoError(t, err)

	goodVal, err := stakingKeeper.GetValidator(ctx, goodValAddr)
	require.NoError(t, err)
	goodValSelfBondInitial := goodVal.ValidatorBondShares

	// next block, commit height 2, move to height 3
	// acc 1, 2, 3, 4 and good validator delegate to evil val
	ctx = ctx.WithBlockHeight(app.LastBlockHeight() + 1).WithHeaderInfo(header.Info{Height: app.LastBlockHeight() + 1})
	fmt.Println()
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

	// Acc 3 delegate
	delMsg = stakingtypes.NewMsgDelegate(testAcc3.String(), evilValAddr.String(), testCoins[0])
	_, err = stakingMsgServer.Delegate(ctx, delMsg)
	require.NoError(t, err)

	// Acc 4 delegate
	delMsg = stakingtypes.NewMsgDelegate(testAcc4.String(), evilValAddr.String(), testCoins[0])
	_, err = stakingMsgServer.Delegate(ctx, delMsg)
	require.NoError(t, err)

	// Good validator delegates to evil validator
	delMsg = stakingtypes.NewMsgDelegate(goodValAccAddr.String(), evilValAddr.String(), testCoins[0])
	_, err = stakingMsgServer.Delegate(ctx, delMsg)
	require.NoError(t, err)

	// Acc 4 liquid stake stokens
	tokenizeMsg := stakingtypes.NewMsgTokenizeShares(testAcc4.String(), evilValAddr.String(), testCoins[0], testAcc4.String())
	res, err := stakingMsgServer.TokenizeShares(ctx, tokenizeMsg)
	require.NoError(t, err)

	tokenizeSharedRcd, err := stakingKeeper.GetTokenizeShareRecordByDenom(ctx, res.Amount.Denom)
	require.NoError(t, err)
	moduleAddr := tokenizeSharedRcd.GetModuleAddress()

	// next block, commit height 3, move to height 4
	// with the new delegations, evil val increases in voting power and commit byzantine behavior at height 4 consensus
	// at the same time, acc 1 and acc 2 withdraw delegation from evil val, acc 3 redelegate but stays bonded
	// good validator redelegates to itself
	ctx, err = simtestutil.NextBlock(app, ctx, time.Duration(1))
	require.NoError(t, err)

	evilVal, err := stakingKeeper.GetValidator(ctx, evilValAddr)
	require.NoError(t, err)

	evilPower := stakingKeeper.TokensToConsensusPower(ctx, evilVal.Tokens)

	// Acc 1 redelegate from evil val to good val
	redelMsg := stakingtypes.NewMsgBeginRedelegate(testAcc1.String(), evilValAddr.String(), goodValAddr.String(), testCoins[0])
	_, err = stakingMsgServer.BeginRedelegate(ctx, redelMsg)
	require.NoError(t, err)

	// Acc 3 redelegate from evil val to good val
	redelMsg = stakingtypes.NewMsgBeginRedelegate(testAcc3.String(), evilValAddr.String(), goodValAddr.String(), testCoins[0])
	_, err = stakingMsgServer.BeginRedelegate(ctx, redelMsg)
	require.NoError(t, err)

	// Acc 4 redelegates tokenized shares from evil val to good val
	// we set moduleAddr as a delegator because delegator of a tokenize share is lsm module. This is hack for testing,
	// in normal scenario this will be ica address that liquid staked some tokens
	redelMsg = stakingtypes.NewMsgBeginRedelegate(moduleAddr.String(), evilValAddr.String(), goodValAddr.String(), testCoins[0])
	_, err = stakingMsgServer.BeginRedelegate(ctx, redelMsg)
	require.NoError(t, err)

	// Good validator redelegates to itself
	redelMsg = stakingtypes.NewMsgBeginRedelegate(goodValAccAddr.String(), evilValAddr.String(), goodValAddr.String(), testCoins[0])
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

	// get states before slashing
	delegation3BeforeSlashing, err := stakingKeeper.GetDelegatorBonded(ctx, testAcc3)
	require.NoError(t, err)

	goodVal, err = stakingKeeper.GetValidator(ctx, goodValAddr)
	require.NoError(t, err)
	goodValSelfBondAfterRedelegation := goodVal.ValidatorBondShares
	goodValLiquidSharesAfterRedelegation := goodVal.LiquidShares

	// next block, commit height 4, move to height 5
	// Slash evil val for byzantine behavior at height 4 consensus,
	// at which acc 1 and acc 2 still contributed to evil val voting power
	// even tho they undelegate at block 4, the valset update is applied after committed block 4 when height 4 consensus already passes
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
	delegation3AfterSlashing, err := stakingKeeper.GetDelegatorBonded(ctx, testAcc3)
	require.NoError(t, err)

	// check unbonded amounts
	require.Equal(t, balance1AfterSlashing.Amount.Mul(math.NewIntFromUint64(10)).String(), balance1Before.Amount.String())
	require.Equal(t, balance2AfterSlashing.Amount.Mul(math.NewIntFromUint64(10)).String(), balance2Before.Amount.String())

	// check bonded amount
	require.Equal(t, delegation3AfterSlashing.Mul(math.NewIntFromUint64(10)).String(), delegation3BeforeSlashing.String())

	// check validator bonded amount
	goodVal, err = stakingKeeper.GetValidator(ctx, goodValAddr)
	require.NoError(t, err)
	redelegatedAmountBeforeSlashing := goodValSelfBondAfterRedelegation.Sub(goodValSelfBondInitial) // unslashed redelegated amount
	redelegatedAmountAfterSlashing := goodVal.ValidatorBondShares.Sub(goodValSelfBondInitial)       // slashed redelegated amount
	require.Equal(t, redelegatedAmountAfterSlashing.Mul(math.LegacyNewDec(10)), redelegatedAmountBeforeSlashing)

	// check liquid shares
	require.Equal(t, goodVal.LiquidShares.Mul(math.LegacyNewDec(10)), goodValLiquidSharesAfterRedelegation)
}
