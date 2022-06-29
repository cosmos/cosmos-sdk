package keeper_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	"github.com/cosmos/cosmos-sdk/codec"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	banktestutil "github.com/cosmos/cosmos-sdk/x/bank/testutil"
	paramskeeper "github.com/cosmos/cosmos-sdk/x/params/keeper"
	"github.com/cosmos/cosmos-sdk/x/staking/keeper"
	"github.com/cosmos/cosmos-sdk/x/staking/teststaking"
	"github.com/cosmos/cosmos-sdk/x/staking/testutil"
	"github.com/cosmos/cosmos-sdk/x/staking/types"
)

type SlashTestSuite struct {
	suite.Suite

	ctx sdk.Context

	addrDels []sdk.AccAddress
	addrVals []sdk.ValAddress

	accountKeeper authkeeper.AccountKeeper
	bankKeeper    bankkeeper.Keeper
	stakingKeeper *keeper.Keeper
}

func (suite *SlashTestSuite) SetupTest() {
	var (
		codec        codec.Codec
		paramsKeeper paramskeeper.Keeper
	)

	app, err := simtestutil.Setup(testutil.AppConfig,
		&suite.bankKeeper,
		&suite.stakingKeeper,
		&suite.accountKeeper,
		&codec,
		&paramsKeeper,
	)
	suite.Require().NoError(err)
	suite.ctx = app.BaseApp.NewContext(false, tmproto.Header{})

	addrDels := simtestutil.AddTestAddrsIncremental(suite.bankKeeper, suite.stakingKeeper, suite.ctx, 100, sdk.NewInt(10000))
	addrVals := simtestutil.ConvertAddrsToValAddrs(addrDels)

	amt := suite.stakingKeeper.TokensFromConsensusPower(suite.ctx, 10)
	totalSupply := sdk.NewCoins(sdk.NewCoin(suite.stakingKeeper.BondDenom(suite.ctx), amt.MulRaw(int64(len(addrDels)))))

	notBondedPool := suite.stakingKeeper.GetNotBondedPool(suite.ctx)
	suite.Require().NoError(banktestutil.FundModuleAccount(suite.bankKeeper, suite.ctx, notBondedPool.GetName(), totalSupply))

	suite.accountKeeper.SetModuleAccount(suite.ctx, notBondedPool)

	numVals := int64(3)
	bondedCoins := sdk.NewCoins(sdk.NewCoin(suite.stakingKeeper.BondDenom(suite.ctx), amt.MulRaw(numVals)))
	bondedPool := suite.stakingKeeper.GetBondedPool(suite.ctx)

	// set bonded pool balance
	suite.accountKeeper.SetModuleAccount(suite.ctx, bondedPool)
	suite.Require().NoError(banktestutil.FundModuleAccount(suite.bankKeeper, suite.ctx, bondedPool.GetName(), bondedCoins))

	for i := int64(0); i < numVals; i++ {
		validator := teststaking.NewValidator(suite.T(), addrVals[i], PKs[i])
		validator, _ = validator.AddTokensFromDel(amt)
		validator = keeper.TestingUpdateValidator(suite.stakingKeeper, suite.ctx, validator, true)
		suite.stakingKeeper.SetValidatorByConsAddr(suite.ctx, validator)
	}

	suite.addrVals = addrVals
	suite.addrDels = addrDels

	// overwrite with custom StakingKeeper to avoid messing with the hooks.
	stakingSubspace, ok := paramsKeeper.GetSubspace(types.ModuleName)
	suite.Require().True(ok)
	suite.stakingKeeper = keeper.NewKeeper(codec, app.UnsafeFindStoreKey(types.StoreKey), suite.accountKeeper, suite.bankKeeper, stakingSubspace)
}

// tests Jail, Unjail
func (suite *SlashTestSuite) TestRevocation() {
	stakingKeeper, ctx, addrVals := suite.stakingKeeper, suite.ctx, suite.addrVals

	consAddr := sdk.ConsAddress(PKs[0].Address())

	// initial state
	val, found := stakingKeeper.GetValidator(ctx, addrVals[0])
	suite.Require().True(found)
	suite.Require().False(val.IsJailed())

	// test jail
	stakingKeeper.Jail(ctx, consAddr)
	val, found = stakingKeeper.GetValidator(ctx, addrVals[0])
	suite.Require().True(found)
	suite.Require().True(val.IsJailed())

	// test unjail
	stakingKeeper.Unjail(ctx, consAddr)
	val, found = stakingKeeper.GetValidator(ctx, addrVals[0])
	suite.Require().True(found)
	suite.Require().False(val.IsJailed())
}

// tests slashUnbondingDelegation
func (suite *SlashTestSuite) TestSlashUnbondingDelegation() {
	bankKeeper, stakingKeeper, ctx, addrVals, addrDels := suite.bankKeeper, suite.stakingKeeper, suite.ctx, suite.addrVals, suite.addrDels

	fraction := sdk.NewDecWithPrec(5, 1)

	// set an unbonding delegation with expiration timestamp (beyond which the
	// unbonding delegation shouldn't be slashed)
	ubd := types.NewUnbondingDelegation(addrDels[0], addrVals[0], 0,
		time.Unix(5, 0), sdk.NewInt(10))

	stakingKeeper.SetUnbondingDelegation(ctx, ubd)

	// unbonding started prior to the infraction height, stakw didn't contribute
	slashAmount := stakingKeeper.SlashUnbondingDelegation(ctx, ubd, 1, fraction)
	suite.Require().True(slashAmount.Equal(sdk.NewInt(0)))

	// after the expiration time, no longer eligible for slashing
	ctx = ctx.WithBlockHeader(tmproto.Header{Time: time.Unix(10, 0)})
	stakingKeeper.SetUnbondingDelegation(ctx, ubd)
	slashAmount = stakingKeeper.SlashUnbondingDelegation(ctx, ubd, 0, fraction)
	suite.Require().True(slashAmount.Equal(sdk.NewInt(0)))

	// test valid slash, before expiration timestamp and to which stake contributed
	notBondedPool := stakingKeeper.GetNotBondedPool(ctx)
	oldUnbondedPoolBalances := bankKeeper.GetAllBalances(ctx, notBondedPool.GetAddress())
	ctx = ctx.WithBlockHeader(tmproto.Header{Time: time.Unix(0, 0)})
	stakingKeeper.SetUnbondingDelegation(ctx, ubd)
	slashAmount = stakingKeeper.SlashUnbondingDelegation(ctx, ubd, 0, fraction)
	suite.Require().True(slashAmount.Equal(sdk.NewInt(5)))
	ubd, found := stakingKeeper.GetUnbondingDelegation(ctx, addrDels[0], addrVals[0])
	suite.Require().True(found)
	suite.Require().Len(ubd.Entries, 1)

	// initial balance unchanged
	suite.Require().Equal(sdk.NewInt(10), ubd.Entries[0].InitialBalance)

	// balance decreased
	suite.Require().Equal(sdk.NewInt(5), ubd.Entries[0].Balance)
	newUnbondedPoolBalances := bankKeeper.GetAllBalances(ctx, notBondedPool.GetAddress())
	diffTokens := oldUnbondedPoolBalances.Sub(newUnbondedPoolBalances...)
	suite.Require().True(diffTokens.AmountOf(stakingKeeper.BondDenom(ctx)).Equal(sdk.NewInt(5)))
}

// tests slashRedelegation
func (suite *SlashTestSuite) TestSlashRedelegation() {
	bankKeeper, stakingKeeper, accountKeeper, ctx, addrVals, addrDels := suite.bankKeeper, suite.stakingKeeper, suite.accountKeeper, suite.ctx, suite.addrVals, suite.addrDels
	fraction := sdk.NewDecWithPrec(5, 1)

	// add bonded tokens to pool for (re)delegations
	startCoins := sdk.NewCoins(sdk.NewInt64Coin(stakingKeeper.BondDenom(ctx), 15))
	bondedPool := stakingKeeper.GetBondedPool(ctx)
	balances := bankKeeper.GetAllBalances(ctx, bondedPool.GetAddress())

	suite.Require().NoError(banktestutil.FundModuleAccount(bankKeeper, ctx, bondedPool.GetName(), startCoins))
	accountKeeper.SetModuleAccount(ctx, bondedPool)

	// set a redelegation with an expiration timestamp beyond which the
	// redelegation shouldn't be slashed
	rd := types.NewRedelegation(addrDels[0], addrVals[0], addrVals[1], 0,
		time.Unix(5, 0), sdk.NewInt(10), sdk.NewDec(10))

	stakingKeeper.SetRedelegation(ctx, rd)

	// set the associated delegation
	del := types.NewDelegation(addrDels[0], addrVals[1], sdk.NewDec(10))
	stakingKeeper.SetDelegation(ctx, del)

	// started redelegating prior to the current height, stake didn't contribute to infraction
	validator, found := stakingKeeper.GetValidator(ctx, addrVals[1])
	suite.Require().True(found)
	slashAmount := stakingKeeper.SlashRedelegation(ctx, validator, rd, 1, fraction)
	suite.Require().True(slashAmount.Equal(sdk.NewInt(0)))

	// after the expiration time, no longer eligible for slashing
	ctx = ctx.WithBlockHeader(tmproto.Header{Time: time.Unix(10, 0)})
	stakingKeeper.SetRedelegation(ctx, rd)
	validator, found = stakingKeeper.GetValidator(ctx, addrVals[1])
	suite.Require().True(found)
	slashAmount = stakingKeeper.SlashRedelegation(ctx, validator, rd, 0, fraction)
	suite.Require().True(slashAmount.Equal(sdk.NewInt(0)))

	balances = bankKeeper.GetAllBalances(ctx, bondedPool.GetAddress())

	// test valid slash, before expiration timestamp and to which stake contributed
	ctx = ctx.WithBlockHeader(tmproto.Header{Time: time.Unix(0, 0)})
	stakingKeeper.SetRedelegation(ctx, rd)
	validator, found = stakingKeeper.GetValidator(ctx, addrVals[1])
	suite.Require().True(found)
	slashAmount = stakingKeeper.SlashRedelegation(ctx, validator, rd, 0, fraction)
	suite.Require().True(slashAmount.Equal(sdk.NewInt(5)))
	rd, found = stakingKeeper.GetRedelegation(ctx, addrDels[0], addrVals[0], addrVals[1])
	suite.Require().True(found)
	suite.Require().Len(rd.Entries, 1)

	// end block
	applyValidatorSetUpdates(suite.T(), ctx, stakingKeeper, 1)

	// initialbalance unchanged
	suite.Require().Equal(sdk.NewInt(10), rd.Entries[0].InitialBalance)

	// shares decreased
	del, found = stakingKeeper.GetDelegation(ctx, addrDels[0], addrVals[1])
	suite.Require().True(found)
	suite.Require().Equal(int64(5), del.Shares.RoundInt64())

	// pool bonded tokens should decrease
	burnedCoins := sdk.NewCoins(sdk.NewCoin(stakingKeeper.BondDenom(ctx), slashAmount))
	suite.Require().Equal(balances.Sub(burnedCoins...), bankKeeper.GetAllBalances(ctx, bondedPool.GetAddress()))
}

// tests Slash at a future height (must panic)
func (suite *SlashTestSuite) TestSlashAtFutureHeight() {
	stakingKeeper, ctx := suite.stakingKeeper, suite.ctx

	consAddr := sdk.ConsAddress(PKs[0].Address())
	fraction := sdk.NewDecWithPrec(5, 1)
	suite.Require().Panics(func() { stakingKeeper.Slash(ctx, consAddr, 1, 10, fraction) })
}

// test slash at a negative height
// this just represents pre-genesis and should have the same effect as slashing at height 0
func (suite *SlashTestSuite) TestSlashAtNegativeHeight() {
	bankKeeper, stakingKeeper, ctx := suite.bankKeeper, suite.stakingKeeper, suite.ctx

	consAddr := sdk.ConsAddress(PKs[0].Address())
	fraction := sdk.NewDecWithPrec(5, 1)

	bondedPool := stakingKeeper.GetBondedPool(ctx)
	oldBondedPoolBalances := bankKeeper.GetAllBalances(ctx, bondedPool.GetAddress())

	validator, found := stakingKeeper.GetValidatorByConsAddr(ctx, consAddr)
	suite.Require().True(found)
	stakingKeeper.Slash(ctx, consAddr, -2, 10, fraction)

	// read updated state
	validator, found = stakingKeeper.GetValidatorByConsAddr(ctx, consAddr)
	suite.Require().True(found)

	// end block
	applyValidatorSetUpdates(suite.T(), ctx, stakingKeeper, 1)

	validator, found = stakingKeeper.GetValidator(ctx, validator.GetOperator())
	suite.Require().True(found)
	// power decreased
	suite.Require().Equal(int64(5), validator.GetConsensusPower(stakingKeeper.PowerReduction(ctx)))

	// pool bonded shares decreased
	newBondedPoolBalances := bankKeeper.GetAllBalances(ctx, bondedPool.GetAddress())
	diffTokens := oldBondedPoolBalances.Sub(newBondedPoolBalances...).AmountOf(stakingKeeper.BondDenom(ctx))
	suite.Require().Equal(stakingKeeper.TokensFromConsensusPower(ctx, 5).String(), diffTokens.String())
}

// tests Slash at the current height
func (suite *SlashTestSuite) TestSlashValidatorAtCurrentHeight() {
	bankKeeper, stakingKeeper, ctx := suite.bankKeeper, suite.stakingKeeper, suite.ctx
	consAddr := sdk.ConsAddress(PKs[0].Address())
	fraction := sdk.NewDecWithPrec(5, 1)

	bondedPool := stakingKeeper.GetBondedPool(ctx)
	oldBondedPoolBalances := bankKeeper.GetAllBalances(ctx, bondedPool.GetAddress())

	validator, found := stakingKeeper.GetValidatorByConsAddr(ctx, consAddr)
	suite.Require().True(found)
	stakingKeeper.Slash(ctx, consAddr, ctx.BlockHeight(), 10, fraction)

	// read updated state
	validator, found = stakingKeeper.GetValidatorByConsAddr(ctx, consAddr)
	suite.Require().True(found)

	// end block
	applyValidatorSetUpdates(suite.T(), ctx, stakingKeeper, 1)

	validator, found = stakingKeeper.GetValidator(ctx, validator.GetOperator())
	suite.Require().True(found)
	// power decreased
	suite.Require().Equal(int64(5), validator.GetConsensusPower(stakingKeeper.PowerReduction(ctx)))

	// pool bonded shares decreased
	newBondedPoolBalances := bankKeeper.GetAllBalances(ctx, bondedPool.GetAddress())
	diffTokens := oldBondedPoolBalances.Sub(newBondedPoolBalances...).AmountOf(stakingKeeper.BondDenom(ctx))
	suite.Require().Equal(stakingKeeper.TokensFromConsensusPower(ctx, 5).String(), diffTokens.String())
}

// tests Slash at a previous height with an unbonding delegation
func (suite *SlashTestSuite) TestSlashWithUnbondingDelegation() {
	bankKeeper, stakingKeeper, ctx, addrVals, addrDels := suite.bankKeeper, suite.stakingKeeper, suite.ctx, suite.addrVals, suite.addrDels

	consAddr := sdk.ConsAddress(PKs[0].Address())
	fraction := sdk.NewDecWithPrec(5, 1)

	// set an unbonding delegation with expiration timestamp beyond which the
	// unbonding delegation shouldn't be slashed
	ubdTokens := stakingKeeper.TokensFromConsensusPower(ctx, 4)
	ubd := types.NewUnbondingDelegation(addrDels[0], addrVals[0], 11, time.Unix(0, 0), ubdTokens)
	stakingKeeper.SetUnbondingDelegation(ctx, ubd)

	// slash validator for the first time
	ctx = ctx.WithBlockHeight(12)
	bondedPool := stakingKeeper.GetBondedPool(ctx)
	oldBondedPoolBalances := bankKeeper.GetAllBalances(ctx, bondedPool.GetAddress())

	validator, found := stakingKeeper.GetValidatorByConsAddr(ctx, consAddr)
	suite.Require().True(found)
	stakingKeeper.Slash(ctx, consAddr, 10, 10, fraction)

	// end block
	applyValidatorSetUpdates(suite.T(), ctx, stakingKeeper, 1)

	// read updating unbonding delegation
	ubd, found = stakingKeeper.GetUnbondingDelegation(ctx, addrDels[0], addrVals[0])
	suite.Require().True(found)
	suite.Require().Len(ubd.Entries, 1)

	// balance decreased
	suite.Require().Equal(stakingKeeper.TokensFromConsensusPower(ctx, 2), ubd.Entries[0].Balance)

	// bonded tokens burned
	newBondedPoolBalances := bankKeeper.GetAllBalances(ctx, bondedPool.GetAddress())
	diffTokens := oldBondedPoolBalances.Sub(newBondedPoolBalances...).AmountOf(stakingKeeper.BondDenom(ctx))
	suite.Require().Equal(stakingKeeper.TokensFromConsensusPower(ctx, 3), diffTokens)

	// read updated validator
	validator, found = stakingKeeper.GetValidatorByConsAddr(ctx, consAddr)
	suite.Require().True(found)

	// power decreased by 3 - 6 stake originally bonded at the time of infraction
	// was still bonded at the time of discovery and was slashed by half, 4 stake
	// bonded at the time of discovery hadn't been bonded at the time of infraction
	// and wasn't slashed
	suite.Require().Equal(int64(7), validator.GetConsensusPower(stakingKeeper.PowerReduction(ctx)))

	// slash validator again
	ctx = ctx.WithBlockHeight(13)
	stakingKeeper.Slash(ctx, consAddr, 9, 10, fraction)

	ubd, found = stakingKeeper.GetUnbondingDelegation(ctx, addrDels[0], addrVals[0])
	suite.Require().True(found)
	suite.Require().Len(ubd.Entries, 1)

	// balance decreased again
	suite.Require().Equal(sdk.NewInt(0), ubd.Entries[0].Balance)

	// bonded tokens burned again
	newBondedPoolBalances = bankKeeper.GetAllBalances(ctx, bondedPool.GetAddress())
	diffTokens = oldBondedPoolBalances.Sub(newBondedPoolBalances...).AmountOf(stakingKeeper.BondDenom(ctx))
	suite.Require().Equal(stakingKeeper.TokensFromConsensusPower(ctx, 6), diffTokens)

	// read updated validator
	validator, found = stakingKeeper.GetValidatorByConsAddr(ctx, consAddr)
	suite.Require().True(found)

	// power decreased by 3 again
	suite.Require().Equal(int64(4), validator.GetConsensusPower(stakingKeeper.PowerReduction(ctx)))

	// slash validator again
	// all originally bonded stake has been slashed, so this will have no effect
	// on the unbonding delegation, but it will slash stake bonded since the infraction
	// this may not be the desirable behaviour, ref https://github.com/cosmos/cosmos-sdk/issues/1440
	ctx = ctx.WithBlockHeight(13)
	stakingKeeper.Slash(ctx, consAddr, 9, 10, fraction)

	ubd, found = stakingKeeper.GetUnbondingDelegation(ctx, addrDels[0], addrVals[0])
	suite.Require().True(found)
	suite.Require().Len(ubd.Entries, 1)

	// balance unchanged
	suite.Require().Equal(sdk.NewInt(0), ubd.Entries[0].Balance)

	// bonded tokens burned again
	newBondedPoolBalances = bankKeeper.GetAllBalances(ctx, bondedPool.GetAddress())
	diffTokens = oldBondedPoolBalances.Sub(newBondedPoolBalances...).AmountOf(stakingKeeper.BondDenom(ctx))
	suite.Require().Equal(stakingKeeper.TokensFromConsensusPower(ctx, 9), diffTokens)

	// read updated validator
	validator, found = stakingKeeper.GetValidatorByConsAddr(ctx, consAddr)
	suite.Require().True(found)

	// power decreased by 3 again
	suite.Require().Equal(int64(1), validator.GetConsensusPower(stakingKeeper.PowerReduction(ctx)))

	// slash validator again
	// all originally bonded stake has been slashed, so this will have no effect
	// on the unbonding delegation, but it will slash stake bonded since the infraction
	// this may not be the desirable behaviour, ref https://github.com/cosmos/cosmos-sdk/issues/1440
	ctx = ctx.WithBlockHeight(13)
	stakingKeeper.Slash(ctx, consAddr, 9, 10, fraction)

	ubd, found = stakingKeeper.GetUnbondingDelegation(ctx, addrDels[0], addrVals[0])
	suite.Require().True(found)
	suite.Require().Len(ubd.Entries, 1)

	// balance unchanged
	suite.Require().Equal(sdk.NewInt(0), ubd.Entries[0].Balance)

	// just 1 bonded token burned again since that's all the validator now has
	newBondedPoolBalances = bankKeeper.GetAllBalances(ctx, bondedPool.GetAddress())
	diffTokens = oldBondedPoolBalances.Sub(newBondedPoolBalances...).AmountOf(stakingKeeper.BondDenom(ctx))
	suite.Require().Equal(stakingKeeper.TokensFromConsensusPower(ctx, 10), diffTokens)

	// apply TM updates
	applyValidatorSetUpdates(suite.T(), ctx, stakingKeeper, -1)

	// read updated validator
	// power decreased by 1 again, validator is out of stake
	// validator should be in unbonding period
	validator, _ = stakingKeeper.GetValidatorByConsAddr(ctx, consAddr)
	suite.Require().Equal(validator.GetStatus(), types.Unbonding)
}

// tests Slash at a previous height with a redelegation
func (suite *SlashTestSuite) TestSlashWithRedelegation() {
	bankKeeper, stakingKeeper, accountKeeper, ctx, addrVals, addrDels := suite.bankKeeper, suite.stakingKeeper, suite.accountKeeper, suite.ctx, suite.addrVals, suite.addrDels

	consAddr := sdk.ConsAddress(PKs[0].Address())
	fraction := sdk.NewDecWithPrec(5, 1)
	bondDenom := stakingKeeper.BondDenom(ctx)

	// set a redelegation
	rdTokens := stakingKeeper.TokensFromConsensusPower(ctx, 6)
	rd := types.NewRedelegation(addrDels[0], addrVals[0], addrVals[1], 11, time.Unix(0, 0), rdTokens, sdk.NewDecFromInt(rdTokens))
	stakingKeeper.SetRedelegation(ctx, rd)

	// set the associated delegation
	del := types.NewDelegation(addrDels[0], addrVals[1], sdk.NewDecFromInt(rdTokens))
	stakingKeeper.SetDelegation(ctx, del)

	// update bonded tokens
	bondedPool := stakingKeeper.GetBondedPool(ctx)
	notBondedPool := stakingKeeper.GetNotBondedPool(ctx)
	rdCoins := sdk.NewCoins(sdk.NewCoin(bondDenom, rdTokens.MulRaw(2)))

	suite.Require().NoError(banktestutil.FundModuleAccount(bankKeeper, ctx, bondedPool.GetName(), rdCoins))

	accountKeeper.SetModuleAccount(ctx, bondedPool)

	oldBonded := bankKeeper.GetBalance(ctx, bondedPool.GetAddress(), bondDenom).Amount
	oldNotBonded := bankKeeper.GetBalance(ctx, notBondedPool.GetAddress(), bondDenom).Amount

	// slash validator
	ctx = ctx.WithBlockHeight(12)
	validator, found := stakingKeeper.GetValidatorByConsAddr(ctx, consAddr)
	suite.Require().True(found)

	suite.Require().NotPanics(func() { stakingKeeper.Slash(ctx, consAddr, 10, 10, fraction) })
	burnAmount := sdk.NewDecFromInt(stakingKeeper.TokensFromConsensusPower(ctx, 10)).Mul(fraction).TruncateInt()

	bondedPool = stakingKeeper.GetBondedPool(ctx)
	notBondedPool = stakingKeeper.GetNotBondedPool(ctx)

	// burn bonded tokens from only from delegations
	bondedPoolBalance := bankKeeper.GetBalance(ctx, bondedPool.GetAddress(), bondDenom).Amount
	require.True(sdk.IntEq(suite.T(), oldBonded.Sub(burnAmount), bondedPoolBalance))

	notBondedPoolBalance := bankKeeper.GetBalance(ctx, notBondedPool.GetAddress(), bondDenom).Amount
	require.True(sdk.IntEq(suite.T(), oldNotBonded, notBondedPoolBalance))
	oldBonded = bankKeeper.GetBalance(ctx, bondedPool.GetAddress(), bondDenom).Amount

	// read updating redelegation
	rd, found = stakingKeeper.GetRedelegation(ctx, addrDels[0], addrVals[0], addrVals[1])
	suite.Require().True(found)
	suite.Require().Len(rd.Entries, 1)
	// read updated validator
	validator, found = stakingKeeper.GetValidatorByConsAddr(ctx, consAddr)
	suite.Require().True(found)
	// power decreased by 2 - 4 stake originally bonded at the time of infraction
	// was still bonded at the time of discovery and was slashed by half, 4 stake
	// bonded at the time of discovery hadn't been bonded at the time of infraction
	// and wasn't slashed
	suite.Require().Equal(int64(8), validator.GetConsensusPower(stakingKeeper.PowerReduction(ctx)))

	// slash the validator again
	validator, found = stakingKeeper.GetValidatorByConsAddr(ctx, consAddr)
	suite.Require().True(found)

	suite.Require().NotPanics(func() { stakingKeeper.Slash(ctx, consAddr, 10, 10, sdk.OneDec()) })
	burnAmount = stakingKeeper.TokensFromConsensusPower(ctx, 7)

	// read updated pool
	bondedPool = stakingKeeper.GetBondedPool(ctx)
	notBondedPool = stakingKeeper.GetNotBondedPool(ctx)

	// seven bonded tokens burned
	bondedPoolBalance = bankKeeper.GetBalance(ctx, bondedPool.GetAddress(), bondDenom).Amount
	require.True(sdk.IntEq(suite.T(), oldBonded.Sub(burnAmount), bondedPoolBalance))
	require.True(sdk.IntEq(suite.T(), oldNotBonded, notBondedPoolBalance))

	bondedPoolBalance = bankKeeper.GetBalance(ctx, bondedPool.GetAddress(), bondDenom).Amount
	require.True(sdk.IntEq(suite.T(), oldBonded.Sub(burnAmount), bondedPoolBalance))

	notBondedPoolBalance = bankKeeper.GetBalance(ctx, notBondedPool.GetAddress(), bondDenom).Amount
	require.True(sdk.IntEq(suite.T(), oldNotBonded, notBondedPoolBalance))
	oldBonded = bankKeeper.GetBalance(ctx, bondedPool.GetAddress(), bondDenom).Amount

	// read updating redelegation
	rd, found = stakingKeeper.GetRedelegation(ctx, addrDels[0], addrVals[0], addrVals[1])
	suite.Require().True(found)
	suite.Require().Len(rd.Entries, 1)
	// read updated validator
	validator, found = stakingKeeper.GetValidatorByConsAddr(ctx, consAddr)
	suite.Require().True(found)
	// power decreased by 4
	suite.Require().Equal(int64(4), validator.GetConsensusPower(stakingKeeper.PowerReduction(ctx)))

	// slash the validator again, by 100%
	ctx = ctx.WithBlockHeight(12)
	validator, found = stakingKeeper.GetValidatorByConsAddr(ctx, consAddr)
	suite.Require().True(found)

	suite.Require().NotPanics(func() { stakingKeeper.Slash(ctx, consAddr, 10, 10, sdk.OneDec()) })

	burnAmount = sdk.NewDecFromInt(stakingKeeper.TokensFromConsensusPower(ctx, 10)).Mul(sdk.OneDec()).TruncateInt()
	burnAmount = burnAmount.Sub(sdk.OneDec().MulInt(rdTokens).TruncateInt())

	// read updated pool
	bondedPool = stakingKeeper.GetBondedPool(ctx)
	notBondedPool = stakingKeeper.GetNotBondedPool(ctx)

	bondedPoolBalance = bankKeeper.GetBalance(ctx, bondedPool.GetAddress(), bondDenom).Amount
	require.True(sdk.IntEq(suite.T(), oldBonded.Sub(burnAmount), bondedPoolBalance))
	notBondedPoolBalance = bankKeeper.GetBalance(ctx, notBondedPool.GetAddress(), bondDenom).Amount
	require.True(sdk.IntEq(suite.T(), oldNotBonded, notBondedPoolBalance))
	oldBonded = bankKeeper.GetBalance(ctx, bondedPool.GetAddress(), bondDenom).Amount

	// read updating redelegation
	rd, found = stakingKeeper.GetRedelegation(ctx, addrDels[0], addrVals[0], addrVals[1])
	suite.Require().True(found)
	suite.Require().Len(rd.Entries, 1)
	// apply TM updates
	applyValidatorSetUpdates(suite.T(), ctx, stakingKeeper, -1)
	// read updated validator
	// validator decreased to zero power, should be in unbonding period
	validator, _ = stakingKeeper.GetValidatorByConsAddr(ctx, consAddr)
	suite.Require().Equal(validator.GetStatus(), types.Unbonding)

	// slash the validator again, by 100%
	// no stake remains to be slashed
	ctx = ctx.WithBlockHeight(12)
	// validator still in unbonding period
	validator, _ = stakingKeeper.GetValidatorByConsAddr(ctx, consAddr)
	suite.Require().Equal(validator.GetStatus(), types.Unbonding)

	suite.Require().NotPanics(func() { stakingKeeper.Slash(ctx, consAddr, 10, 10, sdk.OneDec()) })

	// read updated pool
	bondedPool = stakingKeeper.GetBondedPool(ctx)
	notBondedPool = stakingKeeper.GetNotBondedPool(ctx)

	bondedPoolBalance = bankKeeper.GetBalance(ctx, bondedPool.GetAddress(), bondDenom).Amount
	require.True(sdk.IntEq(suite.T(), oldBonded, bondedPoolBalance))
	notBondedPoolBalance = bankKeeper.GetBalance(ctx, notBondedPool.GetAddress(), bondDenom).Amount
	require.True(sdk.IntEq(suite.T(), oldNotBonded, notBondedPoolBalance))

	// read updating redelegation
	rd, found = stakingKeeper.GetRedelegation(ctx, addrDels[0], addrVals[0], addrVals[1])
	suite.Require().True(found)
	suite.Require().Len(rd.Entries, 1)
	// read updated validator
	// power still zero, still in unbonding period
	validator, _ = stakingKeeper.GetValidatorByConsAddr(ctx, consAddr)
	suite.Require().Equal(validator.GetStatus(), types.Unbonding)
}

// tests Slash at a previous height with both an unbonding delegation and a redelegation
func (suite *SlashTestSuite) TestSlashBoth() {
	bankKeeper, stakingKeeper, accountKeeper, ctx, addrVals, addrDels := suite.bankKeeper, suite.stakingKeeper, suite.accountKeeper, suite.ctx, suite.addrVals, suite.addrDels

	fraction := sdk.NewDecWithPrec(5, 1)
	bondDenom := stakingKeeper.BondDenom(ctx)

	// set a redelegation with expiration timestamp beyond which the
	// redelegation shouldn't be slashed
	rdATokens := stakingKeeper.TokensFromConsensusPower(ctx, 6)
	rdA := types.NewRedelegation(addrDels[0], addrVals[0], addrVals[1], 11, time.Unix(0, 0), rdATokens, sdk.NewDecFromInt(rdATokens))
	stakingKeeper.SetRedelegation(ctx, rdA)

	// set the associated delegation
	delA := types.NewDelegation(addrDels[0], addrVals[1], sdk.NewDecFromInt(rdATokens))
	stakingKeeper.SetDelegation(ctx, delA)

	// set an unbonding delegation with expiration timestamp (beyond which the
	// unbonding delegation shouldn't be slashed)
	ubdATokens := stakingKeeper.TokensFromConsensusPower(ctx, 4)
	ubdA := types.NewUnbondingDelegation(addrDels[0], addrVals[0], 11,
		time.Unix(0, 0), ubdATokens)
	stakingKeeper.SetUnbondingDelegation(ctx, ubdA)

	bondedCoins := sdk.NewCoins(sdk.NewCoin(bondDenom, rdATokens.MulRaw(2)))
	notBondedCoins := sdk.NewCoins(sdk.NewCoin(bondDenom, ubdATokens))

	// update bonded tokens
	bondedPool := stakingKeeper.GetBondedPool(ctx)
	notBondedPool := stakingKeeper.GetNotBondedPool(ctx)

	suite.Require().NoError(banktestutil.FundModuleAccount(bankKeeper, ctx, bondedPool.GetName(), bondedCoins))
	suite.Require().NoError(banktestutil.FundModuleAccount(bankKeeper, ctx, notBondedPool.GetName(), notBondedCoins))

	accountKeeper.SetModuleAccount(ctx, bondedPool)
	accountKeeper.SetModuleAccount(ctx, notBondedPool)

	oldBonded := bankKeeper.GetBalance(ctx, bondedPool.GetAddress(), bondDenom).Amount
	oldNotBonded := bankKeeper.GetBalance(ctx, notBondedPool.GetAddress(), bondDenom).Amount
	// slash validator
	ctx = ctx.WithBlockHeight(12)
	validator, found := stakingKeeper.GetValidatorByConsAddr(ctx, sdk.GetConsAddress(PKs[0]))
	suite.Require().True(found)
	consAddr0 := sdk.ConsAddress(PKs[0].Address())
	stakingKeeper.Slash(ctx, consAddr0, 10, 10, fraction)

	burnedNotBondedAmount := fraction.MulInt(ubdATokens).TruncateInt()
	burnedBondAmount := sdk.NewDecFromInt(stakingKeeper.TokensFromConsensusPower(ctx, 10)).Mul(fraction).TruncateInt()
	burnedBondAmount = burnedBondAmount.Sub(burnedNotBondedAmount)

	// read updated pool
	bondedPool = stakingKeeper.GetBondedPool(ctx)
	notBondedPool = stakingKeeper.GetNotBondedPool(ctx)

	bondedPoolBalance := bankKeeper.GetBalance(ctx, bondedPool.GetAddress(), bondDenom).Amount
	require.True(sdk.IntEq(suite.T(), oldBonded.Sub(burnedBondAmount), bondedPoolBalance))

	notBondedPoolBalance := bankKeeper.GetBalance(ctx, notBondedPool.GetAddress(), bondDenom).Amount
	require.True(sdk.IntEq(suite.T(), oldNotBonded.Sub(burnedNotBondedAmount), notBondedPoolBalance))

	// read updating redelegation
	rdA, found = stakingKeeper.GetRedelegation(ctx, addrDels[0], addrVals[0], addrVals[1])
	suite.Require().True(found)
	suite.Require().Len(rdA.Entries, 1)
	// read updated validator
	validator, found = stakingKeeper.GetValidatorByConsAddr(ctx, sdk.GetConsAddress(PKs[0]))
	suite.Require().True(found)
	// power not decreased, all stake was bonded since
	suite.Require().Equal(int64(10), validator.GetConsensusPower(stakingKeeper.PowerReduction(ctx)))
}

func (suite *SlashTestSuite) TestSlashAmount() {
	bankKeeper, stakingKeeper, ctx := suite.bankKeeper, suite.stakingKeeper, suite.ctx

	consAddr := sdk.ConsAddress(PKs[0].Address())
	fraction := sdk.NewDecWithPrec(5, 1)
	burnedCoins := stakingKeeper.Slash(ctx, consAddr, ctx.BlockHeight(), 10, fraction)
	suite.Require().True(burnedCoins.GT(sdk.ZeroInt()))

	// test the case where the validator was not found, which should return no coins
	addrDels := simtestutil.AddTestAddrsIncremental(bankKeeper, stakingKeeper, ctx, 10, sdk.NewInt(10000))
	addrVals := simtestutil.ConvertAddrsToValAddrs(addrDels)

	noBurned := stakingKeeper.Slash(ctx, sdk.ConsAddress(addrVals[0]), ctx.BlockHeight(), 10, fraction)
	suite.Require().True(sdk.NewInt(0).Equal(noBurned))
}

func TestSlashTestSuite(t *testing.T) {
	suite.Run(t, new(SlashTestSuite))
}
