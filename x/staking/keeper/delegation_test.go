package keeper_test

import (
	"time"

	"github.com/stretchr/testify/require"

	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/bank/testutil"
	"github.com/cosmos/cosmos-sdk/x/staking/keeper"
	"github.com/cosmos/cosmos-sdk/x/staking/teststaking"
	"github.com/cosmos/cosmos-sdk/x/staking/types"
)

// tests GetDelegation, GetDelegatorDelegations, SetDelegation, RemoveDelegation, GetDelegatorDelegations
func (suite *KeeperTestSuite) TestDelegation() {
	// remove genesis validator delegations
	delegations := suite.stakingKeeper.GetAllDelegations(suite.ctx)
	suite.Require().Len(delegations, 1)

	suite.stakingKeeper.RemoveDelegation(suite.ctx, types.Delegation{
		ValidatorAddress: delegations[0].ValidatorAddress,
		DelegatorAddress: delegations[0].DelegatorAddress,
	})

	addrDels := simtestutil.AddTestAddrsIncremental(suite.bankKeeper, suite.stakingKeeper, suite.ctx, 3, sdk.NewInt(10000))
	valAddrs := simtestutil.ConvertAddrsToValAddrs(addrDels)

	// construct the validators
	amts := []sdk.Int{sdk.NewInt(9), sdk.NewInt(8), sdk.NewInt(7)}
	var validators [3]types.Validator
	for i, amt := range amts {
		validators[i] = teststaking.NewValidator(suite.T(), valAddrs[i], PKs[i])
		validators[i], _ = validators[i].AddTokensFromDel(amt)
	}

	validators[0] = keeper.TestingUpdateValidator(suite.stakingKeeper, suite.ctx, validators[0], true)
	validators[1] = keeper.TestingUpdateValidator(suite.stakingKeeper, suite.ctx, validators[1], true)
	validators[2] = keeper.TestingUpdateValidator(suite.stakingKeeper, suite.ctx, validators[2], true)

	// first add a validators[0] to delegate too
	bond1to1 := types.NewDelegation(addrDels[0], valAddrs[0], sdk.NewDec(9))

	// check the empty keeper first
	_, found := suite.stakingKeeper.GetDelegation(suite.ctx, addrDels[0], valAddrs[0])
	suite.Require().False(found)

	// set and retrieve a record
	suite.stakingKeeper.SetDelegation(suite.ctx, bond1to1)
	resBond, found := suite.stakingKeeper.GetDelegation(suite.ctx, addrDels[0], valAddrs[0])
	suite.Require().True(found)
	suite.Require().Equal(bond1to1, resBond)

	// modify a records, save, and retrieve
	bond1to1.Shares = sdk.NewDec(99)
	suite.stakingKeeper.SetDelegation(suite.ctx, bond1to1)
	resBond, found = suite.stakingKeeper.GetDelegation(suite.ctx, addrDels[0], valAddrs[0])
	suite.Require().True(found)
	suite.Require().Equal(bond1to1, resBond)

	// add some more records
	bond1to2 := types.NewDelegation(addrDels[0], valAddrs[1], sdk.NewDec(9))
	bond1to3 := types.NewDelegation(addrDels[0], valAddrs[2], sdk.NewDec(9))
	bond2to1 := types.NewDelegation(addrDels[1], valAddrs[0], sdk.NewDec(9))
	bond2to2 := types.NewDelegation(addrDels[1], valAddrs[1], sdk.NewDec(9))
	bond2to3 := types.NewDelegation(addrDels[1], valAddrs[2], sdk.NewDec(9))
	suite.stakingKeeper.SetDelegation(suite.ctx, bond1to2)
	suite.stakingKeeper.SetDelegation(suite.ctx, bond1to3)
	suite.stakingKeeper.SetDelegation(suite.ctx, bond2to1)
	suite.stakingKeeper.SetDelegation(suite.ctx, bond2to2)
	suite.stakingKeeper.SetDelegation(suite.ctx, bond2to3)

	// test all bond retrieve capabilities
	resBonds := suite.stakingKeeper.GetDelegatorDelegations(suite.ctx, addrDels[0], 5)
	suite.Require().Equal(3, len(resBonds))
	suite.Require().Equal(bond1to1, resBonds[0])
	suite.Require().Equal(bond1to2, resBonds[1])
	suite.Require().Equal(bond1to3, resBonds[2])
	resBonds = suite.stakingKeeper.GetAllDelegatorDelegations(suite.ctx, addrDels[0])
	suite.Require().Equal(3, len(resBonds))
	resBonds = suite.stakingKeeper.GetDelegatorDelegations(suite.ctx, addrDels[0], 2)
	suite.Require().Equal(2, len(resBonds))
	resBonds = suite.stakingKeeper.GetDelegatorDelegations(suite.ctx, addrDels[1], 5)
	suite.Require().Equal(3, len(resBonds))
	suite.Require().Equal(bond2to1, resBonds[0])
	suite.Require().Equal(bond2to2, resBonds[1])
	suite.Require().Equal(bond2to3, resBonds[2])
	allBonds := suite.stakingKeeper.GetAllDelegations(suite.ctx)
	suite.Require().Equal(6, len(allBonds))
	suite.Require().Equal(bond1to1, allBonds[0])
	suite.Require().Equal(bond1to2, allBonds[1])
	suite.Require().Equal(bond1to3, allBonds[2])
	suite.Require().Equal(bond2to1, allBonds[3])
	suite.Require().Equal(bond2to2, allBonds[4])
	suite.Require().Equal(bond2to3, allBonds[5])

	resVals := suite.stakingKeeper.GetDelegatorValidators(suite.ctx, addrDels[0], 3)
	suite.Require().Equal(3, len(resVals))
	resVals = suite.stakingKeeper.GetDelegatorValidators(suite.ctx, addrDels[1], 4)
	suite.Require().Equal(3, len(resVals))

	for i := 0; i < 3; i++ {
		resVal, err := suite.stakingKeeper.GetDelegatorValidator(suite.ctx, addrDels[0], valAddrs[i])
		suite.Require().Nil(err)
		suite.Require().Equal(valAddrs[i], resVal.GetOperator())

		resVal, err = suite.stakingKeeper.GetDelegatorValidator(suite.ctx, addrDels[1], valAddrs[i])
		suite.Require().Nil(err)
		suite.Require().Equal(valAddrs[i], resVal.GetOperator())

		resDels := suite.stakingKeeper.GetValidatorDelegations(suite.ctx, valAddrs[i])
		suite.Require().Len(resDels, 2)
	}

	// test total bonded for single delegator
	expBonded := bond1to1.Shares.Add(bond2to1.Shares).Add(bond1to3.Shares)
	resDelBond := suite.stakingKeeper.GetDelegatorBonded(suite.ctx, addrDels[0])
	suite.Require().Equal(expBonded, sdk.NewDecFromInt(resDelBond))

	// delete a record
	suite.stakingKeeper.RemoveDelegation(suite.ctx, bond2to3)
	_, found = suite.stakingKeeper.GetDelegation(suite.ctx, addrDels[1], valAddrs[2])
	suite.Require().False(found)
	resBonds = suite.stakingKeeper.GetDelegatorDelegations(suite.ctx, addrDels[1], 5)
	suite.Require().Equal(2, len(resBonds))
	suite.Require().Equal(bond2to1, resBonds[0])
	suite.Require().Equal(bond2to2, resBonds[1])

	resBonds = suite.stakingKeeper.GetAllDelegatorDelegations(suite.ctx, addrDels[1])
	suite.Require().Equal(2, len(resBonds))

	// delete all the records from delegator 2
	suite.stakingKeeper.RemoveDelegation(suite.ctx, bond2to1)
	suite.stakingKeeper.RemoveDelegation(suite.ctx, bond2to2)
	_, found = suite.stakingKeeper.GetDelegation(suite.ctx, addrDels[1], valAddrs[0])
	suite.Require().False(found)
	_, found = suite.stakingKeeper.GetDelegation(suite.ctx, addrDels[1], valAddrs[1])
	suite.Require().False(found)
	resBonds = suite.stakingKeeper.GetDelegatorDelegations(suite.ctx, addrDels[1], 5)
	suite.Require().Equal(0, len(resBonds))
}

// tests Get/Set/Remove UnbondingDelegation
func (suite *KeeperTestSuite) TestUnbondingDelegation() {
	delAddrs := simtestutil.AddTestAddrsIncremental(suite.bankKeeper, suite.stakingKeeper, suite.ctx, 2, sdk.NewInt(10000))
	valAddrs := simtestutil.ConvertAddrsToValAddrs(delAddrs)

	ubd := types.NewUnbondingDelegation(
		delAddrs[0],
		valAddrs[0],
		0,
		time.Unix(0, 0).UTC(),
		sdk.NewInt(5),
	)

	// set and retrieve a record
	suite.stakingKeeper.SetUnbondingDelegation(suite.ctx, ubd)
	resUnbond, found := suite.stakingKeeper.GetUnbondingDelegation(suite.ctx, delAddrs[0], valAddrs[0])
	suite.Require().True(found)
	suite.Require().Equal(ubd, resUnbond)

	// modify a records, save, and retrieve
	expUnbond := sdk.NewInt(21)
	ubd.Entries[0].Balance = expUnbond
	suite.stakingKeeper.SetUnbondingDelegation(suite.ctx, ubd)

	resUnbonds := suite.stakingKeeper.GetUnbondingDelegations(suite.ctx, delAddrs[0], 5)
	suite.Require().Equal(1, len(resUnbonds))

	resUnbonds = suite.stakingKeeper.GetAllUnbondingDelegations(suite.ctx, delAddrs[0])
	suite.Require().Equal(1, len(resUnbonds))

	resUnbond, found = suite.stakingKeeper.GetUnbondingDelegation(suite.ctx, delAddrs[0], valAddrs[0])
	suite.Require().True(found)
	suite.Require().Equal(ubd, resUnbond)

	resDelUnbond := suite.stakingKeeper.GetDelegatorUnbonding(suite.ctx, delAddrs[0])
	suite.Require().Equal(expUnbond, resDelUnbond)

	// delete a record
	suite.stakingKeeper.RemoveUnbondingDelegation(suite.ctx, ubd)
	_, found = suite.stakingKeeper.GetUnbondingDelegation(suite.ctx, delAddrs[0], valAddrs[0])
	suite.Require().False(found)

	resUnbonds = suite.stakingKeeper.GetUnbondingDelegations(suite.ctx, delAddrs[0], 5)
	suite.Require().Equal(0, len(resUnbonds))

	resUnbonds = suite.stakingKeeper.GetAllUnbondingDelegations(suite.ctx, delAddrs[0])
	suite.Require().Equal(0, len(resUnbonds))
}

func (suite *KeeperTestSuite) TestUnbondDelegation() {
	delAddrs := simtestutil.AddTestAddrsIncremental(suite.bankKeeper, suite.stakingKeeper, suite.ctx, 1, sdk.NewInt(10000))
	valAddrs := simtestutil.ConvertAddrsToValAddrs(delAddrs)

	startTokens := suite.stakingKeeper.TokensFromConsensusPower(suite.ctx, 10)
	notBondedPool := suite.stakingKeeper.GetNotBondedPool(suite.ctx)

	suite.Require().NoError(testutil.FundModuleAccount(suite.bankKeeper, suite.ctx, notBondedPool.GetName(), sdk.NewCoins(sdk.NewCoin(suite.stakingKeeper.BondDenom(suite.ctx), startTokens))))
	suite.accountKeeper.SetModuleAccount(suite.ctx, notBondedPool)

	// create a validator and a delegator to that validator
	// note this validator starts not-bonded
	validator := teststaking.NewValidator(suite.T(), valAddrs[0], PKs[0])

	validator, issuedShares := validator.AddTokensFromDel(startTokens)
	suite.Require().Equal(startTokens, issuedShares.RoundInt())

	validator = keeper.TestingUpdateValidator(suite.stakingKeeper, suite.ctx, validator, true)

	delegation := types.NewDelegation(delAddrs[0], valAddrs[0], issuedShares)
	suite.stakingKeeper.SetDelegation(suite.ctx, delegation)

	bondTokens := suite.stakingKeeper.TokensFromConsensusPower(suite.ctx, 6)
	amount, err := suite.stakingKeeper.Unbond(suite.ctx, delAddrs[0], valAddrs[0], sdk.NewDecFromInt(bondTokens))
	suite.Require().NoError(err)
	suite.Require().Equal(bondTokens, amount) // shares to be added to an unbonding delegation

	delegation, found := suite.stakingKeeper.GetDelegation(suite.ctx, delAddrs[0], valAddrs[0])
	suite.Require().True(found)
	validator, found = suite.stakingKeeper.GetValidator(suite.ctx, valAddrs[0])
	suite.Require().True(found)

	remainingTokens := startTokens.Sub(bondTokens)
	suite.Require().Equal(remainingTokens, delegation.Shares.RoundInt())
	suite.Require().Equal(remainingTokens, validator.BondedTokens())
}

func (suite *KeeperTestSuite) TestUnbondingDelegationsMaxEntries() {
	ctx := suite.ctx

	addrDels := simtestutil.AddTestAddrsIncremental(suite.bankKeeper, suite.stakingKeeper, ctx, 1, sdk.NewInt(10000))
	addrVals := simtestutil.ConvertAddrsToValAddrs(addrDels)

	startTokens := suite.stakingKeeper.TokensFromConsensusPower(ctx, 10)

	bondDenom := suite.stakingKeeper.BondDenom(ctx)
	notBondedPool := suite.stakingKeeper.GetNotBondedPool(ctx)

	suite.Require().NoError(testutil.FundModuleAccount(suite.bankKeeper, ctx, notBondedPool.GetName(), sdk.NewCoins(sdk.NewCoin(bondDenom, startTokens))))
	suite.accountKeeper.SetModuleAccount(ctx, notBondedPool)

	// create a validator and a delegator to that validator
	validator := teststaking.NewValidator(suite.T(), addrVals[0], PKs[0])

	validator, issuedShares := validator.AddTokensFromDel(startTokens)
	suite.Require().Equal(startTokens, issuedShares.RoundInt())

	validator = keeper.TestingUpdateValidator(suite.stakingKeeper, ctx, validator, true)
	require.True(sdk.IntEq(suite.T(), startTokens, validator.BondedTokens()))
	suite.Require().True(validator.IsBonded())

	delegation := types.NewDelegation(addrDels[0], addrVals[0], issuedShares)
	suite.stakingKeeper.SetDelegation(ctx, delegation)

	maxEntries := suite.stakingKeeper.MaxEntries(ctx)

	oldBonded := suite.bankKeeper.GetBalance(ctx, suite.stakingKeeper.GetBondedPool(ctx).GetAddress(), bondDenom).Amount
	oldNotBonded := suite.bankKeeper.GetBalance(ctx, suite.stakingKeeper.GetNotBondedPool(ctx).GetAddress(), bondDenom).Amount

	// should all pass
	var completionTime time.Time
	for i := uint32(0); i < maxEntries; i++ {
		var err error
		completionTime, err = suite.stakingKeeper.Undelegate(ctx, addrDels[0], addrVals[0], sdk.NewDec(1))
		suite.Require().NoError(err)
	}

	newBonded := suite.bankKeeper.GetBalance(ctx, suite.stakingKeeper.GetBondedPool(ctx).GetAddress(), bondDenom).Amount
	newNotBonded := suite.bankKeeper.GetBalance(ctx, suite.stakingKeeper.GetNotBondedPool(ctx).GetAddress(), bondDenom).Amount
	require.True(sdk.IntEq(suite.T(), newBonded, oldBonded.SubRaw(int64(maxEntries))))
	require.True(sdk.IntEq(suite.T(), newNotBonded, oldNotBonded.AddRaw(int64(maxEntries))))

	oldBonded = suite.bankKeeper.GetBalance(ctx, suite.stakingKeeper.GetBondedPool(ctx).GetAddress(), bondDenom).Amount
	oldNotBonded = suite.bankKeeper.GetBalance(ctx, suite.stakingKeeper.GetNotBondedPool(ctx).GetAddress(), bondDenom).Amount

	// an additional unbond should fail due to max entries
	_, err := suite.stakingKeeper.Undelegate(ctx, addrDels[0], addrVals[0], sdk.NewDec(1))
	suite.Require().Error(err)

	newBonded = suite.bankKeeper.GetBalance(ctx, suite.stakingKeeper.GetBondedPool(ctx).GetAddress(), bondDenom).Amount
	newNotBonded = suite.bankKeeper.GetBalance(ctx, suite.stakingKeeper.GetNotBondedPool(ctx).GetAddress(), bondDenom).Amount

	require.True(sdk.IntEq(suite.T(), newBonded, oldBonded))
	require.True(sdk.IntEq(suite.T(), newNotBonded, oldNotBonded))

	// mature unbonding delegations
	ctx = ctx.WithBlockTime(completionTime)
	_, err = suite.stakingKeeper.CompleteUnbonding(ctx, addrDels[0], addrVals[0])
	suite.Require().NoError(err)

	newBonded = suite.bankKeeper.GetBalance(ctx, suite.stakingKeeper.GetBondedPool(ctx).GetAddress(), bondDenom).Amount
	newNotBonded = suite.bankKeeper.GetBalance(ctx, suite.stakingKeeper.GetNotBondedPool(ctx).GetAddress(), bondDenom).Amount
	require.True(sdk.IntEq(suite.T(), newBonded, oldBonded))
	require.True(sdk.IntEq(suite.T(), newNotBonded, oldNotBonded.SubRaw(int64(maxEntries))))

	oldNotBonded = suite.bankKeeper.GetBalance(ctx, suite.stakingKeeper.GetNotBondedPool(ctx).GetAddress(), bondDenom).Amount

	// unbonding  should work again
	_, err = suite.stakingKeeper.Undelegate(ctx, addrDels[0], addrVals[0], sdk.NewDec(1))
	suite.Require().NoError(err)

	newBonded = suite.bankKeeper.GetBalance(ctx, suite.stakingKeeper.GetBondedPool(ctx).GetAddress(), bondDenom).Amount
	newNotBonded = suite.bankKeeper.GetBalance(ctx, suite.stakingKeeper.GetNotBondedPool(ctx).GetAddress(), bondDenom).Amount
	require.True(sdk.IntEq(suite.T(), newBonded, oldBonded.SubRaw(1)))
	require.True(sdk.IntEq(suite.T(), newNotBonded, oldNotBonded.AddRaw(1)))
}

//// test undelegating self delegation from a validator pushing it below MinSelfDelegation
//// shift it from the bonded to unbonding state and jailed
func (suite *KeeperTestSuite) TestUndelegateSelfDelegationBelowMinSelfDelegation() {
	addrDels := simtestutil.AddTestAddrsIncremental(suite.bankKeeper, suite.stakingKeeper, suite.ctx, 1, sdk.NewInt(10000))
	addrVals := simtestutil.ConvertAddrsToValAddrs(addrDels)
	delTokens := suite.stakingKeeper.TokensFromConsensusPower(suite.ctx, 10)
	delCoins := sdk.NewCoins(sdk.NewCoin(suite.stakingKeeper.BondDenom(suite.ctx), delTokens))

	// create a validator with a self-delegation
	validator := teststaking.NewValidator(suite.T(), addrVals[0], PKs[0])

	validator.MinSelfDelegation = delTokens
	validator, issuedShares := validator.AddTokensFromDel(delTokens)
	suite.Require().Equal(delTokens, issuedShares.RoundInt())

	// add bonded tokens to pool for delegations
	notBondedPool := suite.stakingKeeper.GetNotBondedPool(suite.ctx)
	suite.Require().NoError(testutil.FundModuleAccount(suite.bankKeeper, suite.ctx, notBondedPool.GetName(), delCoins))
	suite.accountKeeper.SetModuleAccount(suite.ctx, notBondedPool)

	validator = keeper.TestingUpdateValidator(suite.stakingKeeper, suite.ctx, validator, true)
	suite.stakingKeeper.SetValidatorByConsAddr(suite.ctx, validator)
	suite.Require().True(validator.IsBonded())

	selfDelegation := types.NewDelegation(sdk.AccAddress(addrVals[0].Bytes()), addrVals[0], issuedShares)
	suite.stakingKeeper.SetDelegation(suite.ctx, selfDelegation)

	// add bonded tokens to pool for delegations
	bondedPool := suite.stakingKeeper.GetBondedPool(suite.ctx)
	suite.Require().NoError(testutil.FundModuleAccount(suite.bankKeeper, suite.ctx, bondedPool.GetName(), delCoins))
	suite.accountKeeper.SetModuleAccount(suite.ctx, bondedPool)

	// create a second delegation to this validator
	suite.stakingKeeper.DeleteValidatorByPowerIndex(suite.ctx, validator)
	validator, issuedShares = validator.AddTokensFromDel(delTokens)
	suite.Require().True(validator.IsBonded())
	suite.Require().Equal(delTokens, issuedShares.RoundInt())

	// add bonded tokens to pool for delegations
	suite.Require().NoError(testutil.FundModuleAccount(suite.bankKeeper, suite.ctx, bondedPool.GetName(), delCoins))
	suite.accountKeeper.SetModuleAccount(suite.ctx, bondedPool)

	validator = keeper.TestingUpdateValidator(suite.stakingKeeper, suite.ctx, validator, true)
	delegation := types.NewDelegation(addrDels[0], addrVals[0], issuedShares)
	suite.stakingKeeper.SetDelegation(suite.ctx, delegation)

	val0AccAddr := sdk.AccAddress(addrVals[0].Bytes())
	_, err := suite.stakingKeeper.Undelegate(suite.ctx, val0AccAddr, addrVals[0], sdk.NewDecFromInt(suite.stakingKeeper.TokensFromConsensusPower(suite.ctx, 6)))
	suite.Require().NoError(err)

	// end block
	applyValidatorSetUpdates(suite.T(), suite.ctx, suite.stakingKeeper, 1)

	validator, found := suite.stakingKeeper.GetValidator(suite.ctx, addrVals[0])
	suite.Require().True(found)
	suite.Require().Equal(suite.stakingKeeper.TokensFromConsensusPower(suite.ctx, 14), validator.Tokens)
	suite.Require().Equal(types.Unbonding, validator.Status)
	suite.Require().True(validator.Jailed)
}

func (suite *KeeperTestSuite) TestUndelegateFromUnbondingValidator() {
	ctx := suite.ctx

	delTokens := suite.stakingKeeper.TokensFromConsensusPower(ctx, 10)
	delCoins := sdk.NewCoins(sdk.NewCoin(suite.stakingKeeper.BondDenom(ctx), delTokens))

	addrDels := simtestutil.AddTestAddrsIncremental(suite.bankKeeper, suite.stakingKeeper, ctx, 2, sdk.NewInt(0))
	addrVals := simtestutil.ConvertAddrsToValAddrs(addrDels)

	// create a validator with a self-delegation
	validator := teststaking.NewValidator(suite.T(), addrVals[0], PKs[0])
	suite.stakingKeeper.SetValidatorByConsAddr(ctx, validator)

	validator, issuedShares := validator.AddTokensFromDel(delTokens)
	suite.Require().Equal(delTokens, issuedShares.RoundInt())

	// add bonded tokens to pool for delegations
	notBondedPool := suite.stakingKeeper.GetNotBondedPool(ctx)
	suite.Require().NoError(testutil.FundModuleAccount(suite.bankKeeper, ctx, notBondedPool.GetName(), delCoins))
	suite.accountKeeper.SetModuleAccount(ctx, notBondedPool)

	validator = keeper.TestingUpdateValidator(suite.stakingKeeper, ctx, validator, true)
	suite.Require().True(validator.IsBonded())

	selfDelegation := types.NewDelegation(addrVals[0].Bytes(), addrVals[0], issuedShares)
	suite.stakingKeeper.SetDelegation(ctx, selfDelegation)

	// add bonded tokens to pool for delegations
	bondedPool := suite.stakingKeeper.GetBondedPool(ctx)
	suite.Require().NoError(testutil.FundModuleAccount(suite.bankKeeper, ctx, bondedPool.GetName(), delCoins))
	suite.accountKeeper.SetModuleAccount(ctx, bondedPool)

	// create a second delegation to this validator
	suite.stakingKeeper.DeleteValidatorByPowerIndex(ctx, validator)

	validator, issuedShares = validator.AddTokensFromDel(delTokens)
	suite.Require().Equal(delTokens, issuedShares.RoundInt())

	suite.Require().NoError(testutil.FundModuleAccount(suite.bankKeeper, ctx, bondedPool.GetName(), delCoins))
	suite.accountKeeper.SetModuleAccount(ctx, bondedPool)

	validator = keeper.TestingUpdateValidator(suite.stakingKeeper, ctx, validator, true)
	delegation := types.NewDelegation(addrDels[1], addrVals[0], issuedShares)
	suite.stakingKeeper.SetDelegation(ctx, delegation)

	suite.Require().NoError(testutil.FundModuleAccount(suite.bankKeeper, ctx, bondedPool.GetName(), delCoins))
	suite.accountKeeper.SetModuleAccount(ctx, bondedPool)

	header := ctx.BlockHeader()
	blockHeight := int64(10)
	header.Height = blockHeight
	blockTime := time.Unix(333, 0)
	header.Time = blockTime
	ctx = ctx.WithBlockHeader(header)

	// unbond the all self-delegation to put validator in unbonding state
	val0AccAddr := sdk.AccAddress(addrVals[0])
	_, err := suite.stakingKeeper.Undelegate(ctx, val0AccAddr, addrVals[0], sdk.NewDecFromInt(delTokens))
	suite.Require().NoError(err)

	// end block
	applyValidatorSetUpdates(suite.T(), ctx, suite.stakingKeeper, 1)

	validator, found := suite.stakingKeeper.GetValidator(ctx, addrVals[0])
	suite.Require().True(found)
	suite.Require().Equal(blockHeight, validator.UnbondingHeight)
	params := suite.stakingKeeper.GetParams(ctx)
	suite.Require().True(blockTime.Add(params.UnbondingTime).Equal(validator.UnbondingTime))

	blockHeight2 := int64(20)
	blockTime2 := time.Unix(444, 0).UTC()
	ctx = ctx.WithBlockHeight(blockHeight2)
	ctx = ctx.WithBlockTime(blockTime2)

	// unbond some of the other delegation's shares
	_, err = suite.stakingKeeper.Undelegate(ctx, addrDels[1], addrVals[0], sdk.NewDec(6))
	suite.Require().NoError(err)

	// retrieve the unbonding delegation
	ubd, found := suite.stakingKeeper.GetUnbondingDelegation(ctx, addrDels[1], addrVals[0])
	suite.Require().True(found)
	suite.Require().Len(ubd.Entries, 1)
	suite.Require().True(ubd.Entries[0].Balance.Equal(sdk.NewInt(6)))
	suite.Require().Equal(blockHeight2, ubd.Entries[0].CreationHeight)
	suite.Require().True(blockTime2.Add(params.UnbondingTime).Equal(ubd.Entries[0].CompletionTime))
}

func (suite *KeeperTestSuite) TestUndelegateFromUnbondedValidator() {
	ctx := suite.ctx

	delTokens := suite.stakingKeeper.TokensFromConsensusPower(ctx, 10)
	delCoins := sdk.NewCoins(sdk.NewCoin(suite.stakingKeeper.BondDenom(ctx), delTokens))

	addrDels := simtestutil.AddTestAddrsIncremental(suite.bankKeeper, suite.stakingKeeper, ctx, 2, sdk.NewInt(0))
	addrVals := simtestutil.ConvertAddrsToValAddrs(addrDels)

	// add bonded tokens to pool for delegations
	notBondedPool := suite.stakingKeeper.GetNotBondedPool(ctx)
	suite.Require().NoError(testutil.FundModuleAccount(suite.bankKeeper, ctx, notBondedPool.GetName(), delCoins))
	suite.accountKeeper.SetModuleAccount(ctx, notBondedPool)

	// create a validator with a self-delegation
	validator := teststaking.NewValidator(suite.T(), addrVals[0], PKs[0])
	suite.stakingKeeper.SetValidatorByConsAddr(ctx, validator)

	valTokens := suite.stakingKeeper.TokensFromConsensusPower(ctx, 10)
	validator, issuedShares := validator.AddTokensFromDel(valTokens)
	suite.Require().Equal(valTokens, issuedShares.RoundInt())
	validator = keeper.TestingUpdateValidator(suite.stakingKeeper, ctx, validator, true)
	suite.Require().True(validator.IsBonded())

	val0AccAddr := sdk.AccAddress(addrVals[0])
	selfDelegation := types.NewDelegation(val0AccAddr, addrVals[0], issuedShares)
	suite.stakingKeeper.SetDelegation(ctx, selfDelegation)

	// add bonded tokens to pool for delegations
	bondedPool := suite.stakingKeeper.GetBondedPool(ctx)
	suite.Require().NoError(testutil.FundModuleAccount(suite.bankKeeper, ctx, bondedPool.GetName(), delCoins))
	suite.accountKeeper.SetModuleAccount(ctx, bondedPool)

	// create a second delegation to this validator
	suite.stakingKeeper.DeleteValidatorByPowerIndex(ctx, validator)
	validator, issuedShares = validator.AddTokensFromDel(delTokens)
	suite.Require().Equal(delTokens, issuedShares.RoundInt())
	validator = keeper.TestingUpdateValidator(suite.stakingKeeper, ctx, validator, true)
	suite.Require().True(validator.IsBonded())
	delegation := types.NewDelegation(addrDels[1], addrVals[0], issuedShares)
	suite.stakingKeeper.SetDelegation(ctx, delegation)

	ctx = ctx.WithBlockHeight(10)
	ctx = ctx.WithBlockTime(time.Unix(333, 0))

	// unbond the all self-delegation to put validator in unbonding state
	_, err := suite.stakingKeeper.Undelegate(ctx, val0AccAddr, addrVals[0], sdk.NewDecFromInt(valTokens))
	suite.Require().NoError(err)

	// end block
	applyValidatorSetUpdates(suite.T(), ctx, suite.stakingKeeper, 1)

	validator, found := suite.stakingKeeper.GetValidator(ctx, addrVals[0])
	suite.Require().True(found)
	suite.Require().Equal(ctx.BlockHeight(), validator.UnbondingHeight)
	params := suite.stakingKeeper.GetParams(ctx)
	suite.Require().True(ctx.BlockHeader().Time.Add(params.UnbondingTime).Equal(validator.UnbondingTime))

	// unbond the validator
	ctx = ctx.WithBlockTime(validator.UnbondingTime)
	suite.stakingKeeper.UnbondAllMatureValidators(ctx)

	// Make sure validator is still in state because there is still an outstanding delegation
	validator, found = suite.stakingKeeper.GetValidator(ctx, addrVals[0])
	suite.Require().True(found)
	suite.Require().Equal(validator.Status, types.Unbonded)

	// unbond some of the other delegation's shares
	unbondTokens := suite.stakingKeeper.TokensFromConsensusPower(ctx, 6)
	_, err = suite.stakingKeeper.Undelegate(ctx, addrDels[1], addrVals[0], sdk.NewDecFromInt(unbondTokens))
	suite.Require().NoError(err)

	// unbond rest of the other delegation's shares
	remainingTokens := delTokens.Sub(unbondTokens)
	_, err = suite.stakingKeeper.Undelegate(ctx, addrDels[1], addrVals[0], sdk.NewDecFromInt(remainingTokens))
	suite.Require().NoError(err)

	//  now validator should be deleted from state
	validator, found = suite.stakingKeeper.GetValidator(ctx, addrVals[0])
	suite.Require().False(found, "%v", validator)
}

func (suite *KeeperTestSuite) TestUnbondingAllDelegationFromValidator() {
	ctx := suite.ctx

	delTokens := suite.stakingKeeper.TokensFromConsensusPower(ctx, 10)
	delCoins := sdk.NewCoins(sdk.NewCoin(suite.stakingKeeper.BondDenom(ctx), delTokens))

	addrDels := simtestutil.AddTestAddrsIncremental(suite.bankKeeper, suite.stakingKeeper, ctx, 2, sdk.NewInt(0))
	addrVals := simtestutil.ConvertAddrsToValAddrs(addrDels)

	// add bonded tokens to pool for delegations
	notBondedPool := suite.stakingKeeper.GetNotBondedPool(ctx)
	suite.Require().NoError(testutil.FundModuleAccount(suite.bankKeeper, ctx, notBondedPool.GetName(), delCoins))
	suite.accountKeeper.SetModuleAccount(ctx, notBondedPool)

	// create a validator with a self-delegation
	validator := teststaking.NewValidator(suite.T(), addrVals[0], PKs[0])
	suite.stakingKeeper.SetValidatorByConsAddr(ctx, validator)

	valTokens := suite.stakingKeeper.TokensFromConsensusPower(ctx, 10)
	validator, issuedShares := validator.AddTokensFromDel(valTokens)
	suite.Require().Equal(valTokens, issuedShares.RoundInt())

	validator = keeper.TestingUpdateValidator(suite.stakingKeeper, ctx, validator, true)
	suite.Require().True(validator.IsBonded())
	val0AccAddr := sdk.AccAddress(addrVals[0].Bytes())

	selfDelegation := types.NewDelegation(val0AccAddr, addrVals[0], issuedShares)
	suite.stakingKeeper.SetDelegation(ctx, selfDelegation)

	// create a second delegation to this validator
	suite.stakingKeeper.DeleteValidatorByPowerIndex(ctx, validator)
	validator, issuedShares = validator.AddTokensFromDel(delTokens)
	suite.Require().Equal(delTokens, issuedShares.RoundInt())

	// add bonded tokens to pool for delegations
	bondedPool := suite.stakingKeeper.GetBondedPool(ctx)
	suite.Require().NoError(testutil.FundModuleAccount(suite.bankKeeper, ctx, bondedPool.GetName(), delCoins))
	suite.accountKeeper.SetModuleAccount(ctx, bondedPool)

	validator = keeper.TestingUpdateValidator(suite.stakingKeeper, ctx, validator, true)
	suite.Require().True(validator.IsBonded())

	delegation := types.NewDelegation(addrDels[1], addrVals[0], issuedShares)
	suite.stakingKeeper.SetDelegation(ctx, delegation)

	ctx = ctx.WithBlockHeight(10)
	ctx = ctx.WithBlockTime(time.Unix(333, 0))

	// unbond the all self-delegation to put validator in unbonding state
	_, err := suite.stakingKeeper.Undelegate(ctx, val0AccAddr, addrVals[0], sdk.NewDecFromInt(valTokens))
	suite.Require().NoError(err)

	// end block
	applyValidatorSetUpdates(suite.T(), ctx, suite.stakingKeeper, 1)

	// unbond all the remaining delegation
	_, err = suite.stakingKeeper.Undelegate(ctx, addrDels[1], addrVals[0], sdk.NewDecFromInt(delTokens))
	suite.Require().NoError(err)

	// validator should still be in state and still be in unbonding state
	validator, found := suite.stakingKeeper.GetValidator(ctx, addrVals[0])
	suite.Require().True(found)
	suite.Require().Equal(validator.Status, types.Unbonding)

	// unbond the validator
	ctx = ctx.WithBlockTime(validator.UnbondingTime)
	suite.stakingKeeper.UnbondAllMatureValidators(ctx)

	// validator should now be deleted from state
	_, found = suite.stakingKeeper.GetValidator(ctx, addrVals[0])
	suite.Require().False(found)
}

// Make sure that that the retrieving the delegations doesn't affect the state
func (suite *KeeperTestSuite) TestGetRedelegationsFromSrcValidator() {
	addrDels := simtestutil.AddTestAddrsIncremental(suite.bankKeeper, suite.stakingKeeper, suite.ctx, 2, sdk.NewInt(0))
	addrVals := simtestutil.ConvertAddrsToValAddrs(addrDels)

	rd := types.NewRedelegation(addrDels[0], addrVals[0], addrVals[1], 0,
		time.Unix(0, 0), sdk.NewInt(5),
		sdk.NewDec(5))

	// set and retrieve a record
	suite.stakingKeeper.SetRedelegation(suite.ctx, rd)
	resBond, found := suite.stakingKeeper.GetRedelegation(suite.ctx, addrDels[0], addrVals[0], addrVals[1])
	suite.Require().True(found)

	// get the redelegations one time
	redelegations := suite.stakingKeeper.GetRedelegationsFromSrcValidator(suite.ctx, addrVals[0])
	suite.Require().Equal(1, len(redelegations))
	suite.Require().Equal(redelegations[0], resBond)

	// get the redelegations a second time, should be exactly the same
	redelegations = suite.stakingKeeper.GetRedelegationsFromSrcValidator(suite.ctx, addrVals[0])
	suite.Require().Equal(1, len(redelegations))
	suite.Require().Equal(redelegations[0], resBond)
}

// tests Get/Set/Remove/Has UnbondingDelegation
func (suite *KeeperTestSuite) TestRedelegation() {
	addrDels := simtestutil.AddTestAddrsIncremental(suite.bankKeeper, suite.stakingKeeper, suite.ctx, 2, sdk.NewInt(0))
	addrVals := simtestutil.ConvertAddrsToValAddrs(addrDels)

	rd := types.NewRedelegation(addrDels[0], addrVals[0], addrVals[1], 0,
		time.Unix(0, 0).UTC(), sdk.NewInt(5),
		sdk.NewDec(5))

	// test shouldn't have and redelegations
	has := suite.stakingKeeper.HasReceivingRedelegation(suite.ctx, addrDels[0], addrVals[1])
	suite.Require().False(has)

	// set and retrieve a record
	suite.stakingKeeper.SetRedelegation(suite.ctx, rd)
	resRed, found := suite.stakingKeeper.GetRedelegation(suite.ctx, addrDels[0], addrVals[0], addrVals[1])
	suite.Require().True(found)

	redelegations := suite.stakingKeeper.GetRedelegationsFromSrcValidator(suite.ctx, addrVals[0])
	suite.Require().Equal(1, len(redelegations))
	suite.Require().Equal(redelegations[0], resRed)

	redelegations = suite.stakingKeeper.GetRedelegations(suite.ctx, addrDels[0], 5)
	suite.Require().Equal(1, len(redelegations))
	suite.Require().Equal(redelegations[0], resRed)

	redelegations = suite.stakingKeeper.GetAllRedelegations(suite.ctx, addrDels[0], nil, nil)
	suite.Require().Equal(1, len(redelegations))
	suite.Require().Equal(redelegations[0], resRed)

	// check if has the redelegation
	has = suite.stakingKeeper.HasReceivingRedelegation(suite.ctx, addrDels[0], addrVals[1])
	suite.Require().True(has)

	// modify a records, save, and retrieve
	rd.Entries[0].SharesDst = sdk.NewDec(21)
	suite.stakingKeeper.SetRedelegation(suite.ctx, rd)

	resRed, found = suite.stakingKeeper.GetRedelegation(suite.ctx, addrDels[0], addrVals[0], addrVals[1])
	suite.Require().True(found)
	suite.Require().Equal(rd, resRed)

	redelegations = suite.stakingKeeper.GetRedelegationsFromSrcValidator(suite.ctx, addrVals[0])
	suite.Require().Equal(1, len(redelegations))
	suite.Require().Equal(redelegations[0], resRed)

	redelegations = suite.stakingKeeper.GetRedelegations(suite.ctx, addrDels[0], 5)
	suite.Require().Equal(1, len(redelegations))
	suite.Require().Equal(redelegations[0], resRed)

	// delete a record
	suite.stakingKeeper.RemoveRedelegation(suite.ctx, rd)
	_, found = suite.stakingKeeper.GetRedelegation(suite.ctx, addrDels[0], addrVals[0], addrVals[1])
	suite.Require().False(found)

	redelegations = suite.stakingKeeper.GetRedelegations(suite.ctx, addrDels[0], 5)
	suite.Require().Equal(0, len(redelegations))

	redelegations = suite.stakingKeeper.GetAllRedelegations(suite.ctx, addrDels[0], nil, nil)
	suite.Require().Equal(0, len(redelegations))
}

func (suite *KeeperTestSuite) TestRedelegateToSameValidator() {
	addrDels := simtestutil.AddTestAddrsIncremental(suite.bankKeeper, suite.stakingKeeper, suite.ctx, 1, sdk.NewInt(0))
	addrVals := simtestutil.ConvertAddrsToValAddrs(addrDels)

	valTokens := suite.stakingKeeper.TokensFromConsensusPower(suite.ctx, 10)
	startCoins := sdk.NewCoins(sdk.NewCoin(suite.stakingKeeper.BondDenom(suite.ctx), valTokens))

	// add bonded tokens to pool for delegations
	notBondedPool := suite.stakingKeeper.GetNotBondedPool(suite.ctx)
	suite.Require().NoError(testutil.FundModuleAccount(suite.bankKeeper, suite.ctx, notBondedPool.GetName(), startCoins))
	suite.accountKeeper.SetModuleAccount(suite.ctx, notBondedPool)

	// create a validator with a self-delegation
	validator := teststaking.NewValidator(suite.T(), addrVals[0], PKs[0])
	validator, issuedShares := validator.AddTokensFromDel(valTokens)
	suite.Require().Equal(valTokens, issuedShares.RoundInt())
	validator = keeper.TestingUpdateValidator(suite.stakingKeeper, suite.ctx, validator, true)
	suite.Require().True(validator.IsBonded())

	val0AccAddr := sdk.AccAddress(addrVals[0].Bytes())
	selfDelegation := types.NewDelegation(val0AccAddr, addrVals[0], issuedShares)
	suite.stakingKeeper.SetDelegation(suite.ctx, selfDelegation)

	_, err := suite.stakingKeeper.BeginRedelegation(suite.ctx, val0AccAddr, addrVals[0], addrVals[0], sdk.NewDec(5))
	suite.Require().Error(err)
}

func (suite *KeeperTestSuite) TestRedelegationMaxEntries() {
	ctx := suite.ctx

	addrDels := simtestutil.AddTestAddrsIncremental(suite.bankKeeper, suite.stakingKeeper, ctx, 2, sdk.NewInt(0))
	addrVals := simtestutil.ConvertAddrsToValAddrs(addrDels)

	startTokens := suite.stakingKeeper.TokensFromConsensusPower(ctx, 20)
	startCoins := sdk.NewCoins(sdk.NewCoin(suite.stakingKeeper.BondDenom(ctx), startTokens))

	// add bonded tokens to pool for delegations
	notBondedPool := suite.stakingKeeper.GetNotBondedPool(ctx)
	suite.Require().NoError(testutil.FundModuleAccount(suite.bankKeeper, ctx, notBondedPool.GetName(), startCoins))
	suite.accountKeeper.SetModuleAccount(ctx, notBondedPool)

	// create a validator with a self-delegation
	validator := teststaking.NewValidator(suite.T(), addrVals[0], PKs[0])
	valTokens := suite.stakingKeeper.TokensFromConsensusPower(ctx, 10)
	validator, issuedShares := validator.AddTokensFromDel(valTokens)
	suite.Require().Equal(valTokens, issuedShares.RoundInt())
	validator = keeper.TestingUpdateValidator(suite.stakingKeeper, ctx, validator, true)
	val0AccAddr := sdk.AccAddress(addrVals[0].Bytes())
	selfDelegation := types.NewDelegation(val0AccAddr, addrVals[0], issuedShares)
	suite.stakingKeeper.SetDelegation(ctx, selfDelegation)

	// create a second validator
	validator2 := teststaking.NewValidator(suite.T(), addrVals[1], PKs[1])
	validator2, issuedShares = validator2.AddTokensFromDel(valTokens)
	suite.Require().Equal(valTokens, issuedShares.RoundInt())

	validator2 = keeper.TestingUpdateValidator(suite.stakingKeeper, ctx, validator2, true)
	suite.Require().Equal(types.Bonded, validator2.Status)

	maxEntries := suite.stakingKeeper.MaxEntries(ctx)

	// redelegations should pass
	var completionTime time.Time
	for i := uint32(0); i < maxEntries; i++ {
		var err error
		completionTime, err = suite.stakingKeeper.BeginRedelegation(ctx, val0AccAddr, addrVals[0], addrVals[1], sdk.NewDec(1))
		suite.Require().NoError(err)
	}

	// an additional redelegation should fail due to max entries
	_, err := suite.stakingKeeper.BeginRedelegation(ctx, val0AccAddr, addrVals[0], addrVals[1], sdk.NewDec(1))
	suite.Require().Error(err)

	// mature redelegations
	ctx = ctx.WithBlockTime(completionTime)
	_, err = suite.stakingKeeper.CompleteRedelegation(ctx, val0AccAddr, addrVals[0], addrVals[1])
	suite.Require().NoError(err)

	// redelegation should work again
	_, err = suite.stakingKeeper.BeginRedelegation(ctx, val0AccAddr, addrVals[0], addrVals[1], sdk.NewDec(1))
	suite.Require().NoError(err)
}

func (suite *KeeperTestSuite) TestRedelegateSelfDelegation() {
	addrDels := simtestutil.AddTestAddrsIncremental(suite.bankKeeper, suite.stakingKeeper, suite.ctx, 2, sdk.NewInt(0))
	addrVals := simtestutil.ConvertAddrsToValAddrs(addrDels)

	startTokens := suite.stakingKeeper.TokensFromConsensusPower(suite.ctx, 30)
	startCoins := sdk.NewCoins(sdk.NewCoin(suite.stakingKeeper.BondDenom(suite.ctx), startTokens))

	// add bonded tokens to pool for delegations
	notBondedPool := suite.stakingKeeper.GetNotBondedPool(suite.ctx)
	suite.Require().NoError(testutil.FundModuleAccount(suite.bankKeeper, suite.ctx, notBondedPool.GetName(), startCoins))
	suite.accountKeeper.SetModuleAccount(suite.ctx, notBondedPool)

	// create a validator with a self-delegation
	validator := teststaking.NewValidator(suite.T(), addrVals[0], PKs[0])
	suite.stakingKeeper.SetValidatorByConsAddr(suite.ctx, validator)

	valTokens := suite.stakingKeeper.TokensFromConsensusPower(suite.ctx, 10)
	validator, issuedShares := validator.AddTokensFromDel(valTokens)
	suite.Require().Equal(valTokens, issuedShares.RoundInt())

	validator = keeper.TestingUpdateValidator(suite.stakingKeeper, suite.ctx, validator, true)

	val0AccAddr := sdk.AccAddress(addrVals[0])
	selfDelegation := types.NewDelegation(val0AccAddr, addrVals[0], issuedShares)
	suite.stakingKeeper.SetDelegation(suite.ctx, selfDelegation)

	// create a second validator
	validator2 := teststaking.NewValidator(suite.T(), addrVals[1], PKs[1])
	validator2, issuedShares = validator2.AddTokensFromDel(valTokens)
	suite.Require().Equal(valTokens, issuedShares.RoundInt())
	validator2 = keeper.TestingUpdateValidator(suite.stakingKeeper, suite.ctx, validator2, true)
	suite.Require().Equal(types.Bonded, validator2.Status)

	// create a second delegation to validator 1
	delTokens := suite.stakingKeeper.TokensFromConsensusPower(suite.ctx, 10)
	validator, issuedShares = validator.AddTokensFromDel(delTokens)
	suite.Require().Equal(delTokens, issuedShares.RoundInt())
	validator = keeper.TestingUpdateValidator(suite.stakingKeeper, suite.ctx, validator, true)

	delegation := types.NewDelegation(addrDels[0], addrVals[0], issuedShares)
	suite.stakingKeeper.SetDelegation(suite.ctx, delegation)

	_, err := suite.stakingKeeper.BeginRedelegation(suite.ctx, val0AccAddr, addrVals[0], addrVals[1], sdk.NewDecFromInt(delTokens))
	suite.Require().NoError(err)

	// end block
	applyValidatorSetUpdates(suite.T(), suite.ctx, suite.stakingKeeper, 2)

	validator, found := suite.stakingKeeper.GetValidator(suite.ctx, addrVals[0])
	suite.Require().True(found)
	suite.Require().Equal(valTokens, validator.Tokens)
	suite.Require().Equal(types.Unbonding, validator.Status)
}

func (suite *KeeperTestSuite) TestRedelegateFromUnbondingValidator() {
	ctx := suite.ctx

	addrDels := simtestutil.AddTestAddrsIncremental(suite.bankKeeper, suite.stakingKeeper, ctx, 2, sdk.NewInt(0))
	addrVals := simtestutil.ConvertAddrsToValAddrs(addrDels)

	startTokens := suite.stakingKeeper.TokensFromConsensusPower(ctx, 30)
	startCoins := sdk.NewCoins(sdk.NewCoin(suite.stakingKeeper.BondDenom(ctx), startTokens))

	// add bonded tokens to pool for delegations
	notBondedPool := suite.stakingKeeper.GetNotBondedPool(ctx)
	suite.Require().NoError(testutil.FundModuleAccount(suite.bankKeeper, ctx, notBondedPool.GetName(), startCoins))
	suite.accountKeeper.SetModuleAccount(ctx, notBondedPool)

	// create a validator with a self-delegation
	validator := teststaking.NewValidator(suite.T(), addrVals[0], PKs[0])
	suite.stakingKeeper.SetValidatorByConsAddr(ctx, validator)

	valTokens := suite.stakingKeeper.TokensFromConsensusPower(ctx, 10)
	validator, issuedShares := validator.AddTokensFromDel(valTokens)
	suite.Require().Equal(valTokens, issuedShares.RoundInt())
	validator = keeper.TestingUpdateValidator(suite.stakingKeeper, ctx, validator, true)
	val0AccAddr := sdk.AccAddress(addrVals[0].Bytes())
	selfDelegation := types.NewDelegation(val0AccAddr, addrVals[0], issuedShares)
	suite.stakingKeeper.SetDelegation(ctx, selfDelegation)

	// create a second delegation to this validator
	suite.stakingKeeper.DeleteValidatorByPowerIndex(ctx, validator)
	delTokens := suite.stakingKeeper.TokensFromConsensusPower(ctx, 10)
	validator, issuedShares = validator.AddTokensFromDel(delTokens)
	suite.Require().Equal(delTokens, issuedShares.RoundInt())
	validator = keeper.TestingUpdateValidator(suite.stakingKeeper, ctx, validator, true)
	delegation := types.NewDelegation(addrDels[1], addrVals[0], issuedShares)
	suite.stakingKeeper.SetDelegation(ctx, delegation)

	// create a second validator
	validator2 := teststaking.NewValidator(suite.T(), addrVals[1], PKs[1])
	validator2, issuedShares = validator2.AddTokensFromDel(valTokens)
	suite.Require().Equal(valTokens, issuedShares.RoundInt())
	validator2 = keeper.TestingUpdateValidator(suite.stakingKeeper, ctx, validator2, true)

	header := ctx.BlockHeader()
	blockHeight := int64(10)
	header.Height = blockHeight
	blockTime := time.Unix(333, 0)
	header.Time = blockTime
	ctx = ctx.WithBlockHeader(header)

	// unbond the all self-delegation to put validator in unbonding state
	_, err := suite.stakingKeeper.Undelegate(ctx, val0AccAddr, addrVals[0], sdk.NewDecFromInt(delTokens))
	suite.Require().NoError(err)

	// end block
	applyValidatorSetUpdates(suite.T(), ctx, suite.stakingKeeper, 1)

	validator, found := suite.stakingKeeper.GetValidator(ctx, addrVals[0])
	suite.Require().True(found)
	suite.Require().Equal(blockHeight, validator.UnbondingHeight)
	params := suite.stakingKeeper.GetParams(ctx)
	suite.Require().True(blockTime.Add(params.UnbondingTime).Equal(validator.UnbondingTime))

	// change the context
	header = ctx.BlockHeader()
	blockHeight2 := int64(20)
	header.Height = blockHeight2
	blockTime2 := time.Unix(444, 0)
	header.Time = blockTime2
	ctx = ctx.WithBlockHeader(header)

	// unbond some of the other delegation's shares
	redelegateTokens := suite.stakingKeeper.TokensFromConsensusPower(ctx, 6)
	_, err = suite.stakingKeeper.BeginRedelegation(ctx, addrDels[1], addrVals[0], addrVals[1], sdk.NewDecFromInt(redelegateTokens))
	suite.Require().NoError(err)

	// retrieve the unbonding delegation
	ubd, found := suite.stakingKeeper.GetRedelegation(ctx, addrDels[1], addrVals[0], addrVals[1])
	suite.Require().True(found)
	suite.Require().Len(ubd.Entries, 1)
	suite.Require().Equal(blockHeight, ubd.Entries[0].CreationHeight)
	suite.Require().True(blockTime.Add(params.UnbondingTime).Equal(ubd.Entries[0].CompletionTime))
}

func (suite *KeeperTestSuite) TestRedelegateFromUnbondedValidator() {
	ctx := suite.ctx

	addrDels := simtestutil.AddTestAddrsIncremental(suite.bankKeeper, suite.stakingKeeper, ctx, 2, sdk.NewInt(0))
	addrVals := simtestutil.ConvertAddrsToValAddrs(addrDels)

	startTokens := suite.stakingKeeper.TokensFromConsensusPower(ctx, 30)
	startCoins := sdk.NewCoins(sdk.NewCoin(suite.stakingKeeper.BondDenom(ctx), startTokens))

	// add bonded tokens to pool for delegations
	notBondedPool := suite.stakingKeeper.GetNotBondedPool(ctx)
	suite.Require().NoError(testutil.FundModuleAccount(suite.bankKeeper, ctx, notBondedPool.GetName(), startCoins))
	suite.accountKeeper.SetModuleAccount(ctx, notBondedPool)

	// create a validator with a self-delegation
	validator := teststaking.NewValidator(suite.T(), addrVals[0], PKs[0])
	suite.stakingKeeper.SetValidatorByConsAddr(ctx, validator)

	valTokens := suite.stakingKeeper.TokensFromConsensusPower(ctx, 10)
	validator, issuedShares := validator.AddTokensFromDel(valTokens)
	suite.Require().Equal(valTokens, issuedShares.RoundInt())
	validator = keeper.TestingUpdateValidator(suite.stakingKeeper, ctx, validator, true)
	val0AccAddr := sdk.AccAddress(addrVals[0].Bytes())
	selfDelegation := types.NewDelegation(val0AccAddr, addrVals[0], issuedShares)
	suite.stakingKeeper.SetDelegation(ctx, selfDelegation)

	// create a second delegation to this validator
	suite.stakingKeeper.DeleteValidatorByPowerIndex(ctx, validator)
	delTokens := suite.stakingKeeper.TokensFromConsensusPower(ctx, 10)
	validator, issuedShares = validator.AddTokensFromDel(delTokens)
	suite.Require().Equal(delTokens, issuedShares.RoundInt())
	validator = keeper.TestingUpdateValidator(suite.stakingKeeper, ctx, validator, true)
	delegation := types.NewDelegation(addrDels[1], addrVals[0], issuedShares)
	suite.stakingKeeper.SetDelegation(ctx, delegation)

	// create a second validator
	validator2 := teststaking.NewValidator(suite.T(), addrVals[1], PKs[1])
	validator2, issuedShares = validator2.AddTokensFromDel(valTokens)
	suite.Require().Equal(valTokens, issuedShares.RoundInt())
	validator2 = keeper.TestingUpdateValidator(suite.stakingKeeper, ctx, validator2, true)
	suite.Require().Equal(types.Bonded, validator2.Status)

	ctx = ctx.WithBlockHeight(10)
	ctx = ctx.WithBlockTime(time.Unix(333, 0))

	// unbond the all self-delegation to put validator in unbonding state
	_, err := suite.stakingKeeper.Undelegate(ctx, val0AccAddr, addrVals[0], sdk.NewDecFromInt(delTokens))
	suite.Require().NoError(err)

	// end block
	applyValidatorSetUpdates(suite.T(), ctx, suite.stakingKeeper, 1)

	validator, found := suite.stakingKeeper.GetValidator(ctx, addrVals[0])
	suite.Require().True(found)
	suite.Require().Equal(ctx.BlockHeight(), validator.UnbondingHeight)
	params := suite.stakingKeeper.GetParams(ctx)
	suite.Require().True(ctx.BlockHeader().Time.Add(params.UnbondingTime).Equal(validator.UnbondingTime))

	// unbond the validator
	suite.stakingKeeper.UnbondingToUnbonded(ctx, validator)

	// redelegate some of the delegation's shares
	redelegationTokens := suite.stakingKeeper.TokensFromConsensusPower(ctx, 6)
	_, err = suite.stakingKeeper.BeginRedelegation(ctx, addrDels[1], addrVals[0], addrVals[1], sdk.NewDecFromInt(redelegationTokens))
	suite.Require().NoError(err)

	// no red should have been found
	red, found := suite.stakingKeeper.GetRedelegation(ctx, addrDels[0], addrVals[0], addrVals[1])
	suite.Require().False(found, "%v", red)
}
