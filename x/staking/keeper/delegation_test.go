package keeper_test

import (
	"time"

	"cosmossdk.io/math"
	"github.com/golang/mock/gomock"

	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	"github.com/cosmos/cosmos-sdk/x/staking/testutil"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

func createValAddrs(count int) ([]sdk.AccAddress, []sdk.ValAddress) {
	addrs := simtestutil.CreateIncrementalAccounts(count)
	valAddrs := simtestutil.ConvertAddrsToValAddrs(addrs)

	return addrs, valAddrs
}

// tests GetDelegation, GetDelegatorDelegations, SetDelegation, RemoveDelegation, GetDelegatorDelegations
func (s *KeeperTestSuite) TestDelegation() {
	ctx, keeper := s.ctx, s.stakingKeeper
	require := s.Require()

	addrDels, valAddrs := createValAddrs(3)

	// construct the validators
	amts := []math.Int{sdk.NewInt(9), sdk.NewInt(8), sdk.NewInt(7)}
	var validators [3]stakingtypes.Validator
	for i, amt := range amts {
		validators[i] = testutil.NewValidator(s.T(), valAddrs[i], PKs[i])
		validators[i], _ = validators[i].AddTokensFromDel(amt)

		validators[i] = stakingkeeper.TestingUpdateValidator(keeper, ctx, validators[i], true)
	}

	// first add a validators[0] to delegate too
	bond1to1 := stakingtypes.NewDelegation(addrDels[0], valAddrs[0], math.LegacyNewDec(9))

	// check the empty keeper first
	_, found := keeper.GetDelegation(ctx, addrDels[0], valAddrs[0])
	require.False(found)

	// set and retrieve a record
	keeper.SetDelegation(ctx, bond1to1)
	resBond, found := keeper.GetDelegation(ctx, addrDels[0], valAddrs[0])
	require.True(found)
	require.Equal(bond1to1, resBond)

	// modify a records, save, and retrieve
	bond1to1.Shares = math.LegacyNewDec(99)
	keeper.SetDelegation(ctx, bond1to1)
	resBond, found = keeper.GetDelegation(ctx, addrDels[0], valAddrs[0])
	require.True(found)
	require.Equal(bond1to1, resBond)

	// add some more records
	bond1to2 := stakingtypes.NewDelegation(addrDels[0], valAddrs[1], math.LegacyNewDec(9))
	bond1to3 := stakingtypes.NewDelegation(addrDels[0], valAddrs[2], math.LegacyNewDec(9))
	bond2to1 := stakingtypes.NewDelegation(addrDels[1], valAddrs[0], math.LegacyNewDec(9))
	bond2to2 := stakingtypes.NewDelegation(addrDels[1], valAddrs[1], math.LegacyNewDec(9))
	bond2to3 := stakingtypes.NewDelegation(addrDels[1], valAddrs[2], math.LegacyNewDec(9))
	keeper.SetDelegation(ctx, bond1to2)
	keeper.SetDelegation(ctx, bond1to3)
	keeper.SetDelegation(ctx, bond2to1)
	keeper.SetDelegation(ctx, bond2to2)
	keeper.SetDelegation(ctx, bond2to3)

	// test all bond retrieve capabilities
	resBonds := keeper.GetDelegatorDelegations(ctx, addrDels[0], 5)
	require.Equal(3, len(resBonds))
	require.Equal(bond1to1, resBonds[0])
	require.Equal(bond1to2, resBonds[1])
	require.Equal(bond1to3, resBonds[2])
	resBonds = keeper.GetAllDelegatorDelegations(ctx, addrDels[0])
	require.Equal(3, len(resBonds))
	resBonds = keeper.GetDelegatorDelegations(ctx, addrDels[0], 2)
	require.Equal(2, len(resBonds))
	resBonds = keeper.GetDelegatorDelegations(ctx, addrDels[1], 5)
	require.Equal(3, len(resBonds))
	require.Equal(bond2to1, resBonds[0])
	require.Equal(bond2to2, resBonds[1])
	require.Equal(bond2to3, resBonds[2])
	allBonds := keeper.GetAllDelegations(ctx)
	require.Equal(6, len(allBonds))
	require.Equal(bond1to1, allBonds[0])
	require.Equal(bond1to2, allBonds[1])
	require.Equal(bond1to3, allBonds[2])
	require.Equal(bond2to1, allBonds[3])
	require.Equal(bond2to2, allBonds[4])
	require.Equal(bond2to3, allBonds[5])

	resVals := keeper.GetDelegatorValidators(ctx, addrDels[0], 3)
	require.Equal(3, len(resVals))
	resVals = keeper.GetDelegatorValidators(ctx, addrDels[1], 4)
	require.Equal(3, len(resVals))

	for i := 0; i < 3; i++ {
		resVal, err := keeper.GetDelegatorValidator(ctx, addrDels[0], valAddrs[i])
		require.Nil(err)
		require.Equal(valAddrs[i], resVal.GetOperator())

		resVal, err = keeper.GetDelegatorValidator(ctx, addrDels[1], valAddrs[i])
		require.Nil(err)
		require.Equal(valAddrs[i], resVal.GetOperator())

		resDels := keeper.GetValidatorDelegations(ctx, valAddrs[i])
		require.Len(resDels, 2)
	}

	// test total bonded for single delegator
	expBonded := bond1to1.Shares.Add(bond2to1.Shares).Add(bond1to3.Shares)
	resDelBond := keeper.GetDelegatorBonded(ctx, addrDels[0])
	require.Equal(expBonded, sdk.NewDecFromInt(resDelBond))

	// delete a record
	keeper.RemoveDelegation(ctx, bond2to3)
	_, found = keeper.GetDelegation(ctx, addrDels[1], valAddrs[2])
	require.False(found)
	resBonds = keeper.GetDelegatorDelegations(ctx, addrDels[1], 5)
	require.Equal(2, len(resBonds))
	require.Equal(bond2to1, resBonds[0])
	require.Equal(bond2to2, resBonds[1])

	resBonds = keeper.GetAllDelegatorDelegations(ctx, addrDels[1])
	require.Equal(2, len(resBonds))

	// delete all the records from delegator 2
	keeper.RemoveDelegation(ctx, bond2to1)
	keeper.RemoveDelegation(ctx, bond2to2)
	_, found = keeper.GetDelegation(ctx, addrDels[1], valAddrs[0])
	require.False(found)
	_, found = keeper.GetDelegation(ctx, addrDels[1], valAddrs[1])
	require.False(found)
	resBonds = keeper.GetDelegatorDelegations(ctx, addrDels[1], 5)
	require.Equal(0, len(resBonds))
}

func TestTransferDelegation(t *testing.T) {
	_, app, ctx := createTestInput(t)

	addrDels := simapp.AddTestAddrsIncremental(app, ctx, 3, sdk.NewInt(10000))
	valAddrs := simapp.ConvertAddrsToValAddrs(addrDels)

	// construct the validators
	amts := []sdk.Int{sdk.NewInt(9), sdk.NewInt(8), sdk.NewInt(7)}
	var validators [3]types.Validator
	for i, amt := range amts {
		validators[i] = teststaking.NewValidator(t, valAddrs[i], PKs[i])
		validators[i], _ = validators[i].AddTokensFromDel(amt)
	}
	validators[0] = keeper.TestingUpdateValidator(app.StakingKeeper, ctx, validators[0], true)
	validators[1] = keeper.TestingUpdateValidator(app.StakingKeeper, ctx, validators[1], true)
	validators[2] = keeper.TestingUpdateValidator(app.StakingKeeper, ctx, validators[2], true)

	// try a transfer when there's nothing
	transferred := app.StakingKeeper.TransferDelegation(ctx, addrDels[0], addrDels[1], valAddrs[0], sdk.NewDec(1000))
	require.Equal(t, sdk.ZeroDec(), transferred)

	// stake some tokens
	bond1to1 := types.NewDelegation(addrDels[0], valAddrs[0], sdk.NewDec(99))
	app.StakingKeeper.SetDelegation(ctx, bond1to1)
	// stake to an unrelated validator so implementation has to skip it
	bond1to3 := types.NewDelegation(addrDels[0], valAddrs[2], sdk.NewDec(9))
	app.StakingKeeper.SetDelegation(ctx, bond1to3)

	// transfer nothing
	transferred = app.StakingKeeper.TransferDelegation(ctx, addrDels[0], addrDels[1], valAddrs[0], sdk.ZeroDec())
	require.Equal(t, sdk.ZeroDec(), transferred)

	// partial transfer, empty recipient
	transferred = app.StakingKeeper.TransferDelegation(ctx, addrDels[0], addrDels[1], valAddrs[0], sdk.NewDec(10))
	require.Equal(t, sdk.NewDec(10), transferred)
	resBond, found := app.StakingKeeper.GetDelegation(ctx, addrDels[0], valAddrs[0])
	require.True(t, found)
	require.Equal(t, sdk.NewDec(89), resBond.Shares)
	resBond, found = app.StakingKeeper.GetDelegation(ctx, addrDels[1], valAddrs[0])
	require.True(t, found)
	require.Equal(t, sdk.NewDec(10), resBond.Shares)

	// partial transfer, existing recipient
	transferred = app.StakingKeeper.TransferDelegation(ctx, addrDels[0], addrDels[1], valAddrs[0], sdk.NewDec(11))
	require.Equal(t, transferred, sdk.NewDec(11))
	resBond, found = app.StakingKeeper.GetDelegation(ctx, addrDels[0], valAddrs[0])
	require.True(t, found)
	require.Equal(t, sdk.NewDec(78), resBond.Shares)
	resBond, found = app.StakingKeeper.GetDelegation(ctx, addrDels[1], valAddrs[0])
	require.True(t, found)
	require.Equal(t, sdk.NewDec(21), resBond.Shares)

	// full transfer
	transferred = app.StakingKeeper.TransferDelegation(ctx, addrDels[0], addrDels[1], valAddrs[0], sdk.NewDec(9999))
	require.Equal(t, transferred, sdk.NewDec(78))
	resBond, found = app.StakingKeeper.GetDelegation(ctx, addrDels[0], valAddrs[0])
	require.False(t, found)
	resBond, found = app.StakingKeeper.GetDelegation(ctx, addrDels[1], valAddrs[0])
	require.True(t, found)
	require.Equal(t, sdk.NewDec(99), resBond.Shares)

	// simulate redelegate to another validator
	bond1to2 := types.NewDelegation(addrDels[0], valAddrs[1], sdk.NewDec(20))
	app.StakingKeeper.SetDelegation(ctx, bond1to2)
	rd := types.NewRedelegation(addrDels[0], valAddrs[0], valAddrs[1], 0, time.Unix(0, 0).UTC(), sdk.NewInt(20), sdk.NewDec(20))
	app.StakingKeeper.SetRedelegation(ctx, rd)

	// partial transfer from redelegation
	transferred = app.StakingKeeper.TransferDelegation(ctx, addrDels[0], addrDels[1], valAddrs[1], sdk.NewDec(7))
	require.Equal(t, sdk.NewDec(7), transferred)
	resBond, found = app.StakingKeeper.GetDelegation(ctx, addrDels[0], valAddrs[1])
	require.True(t, found)
	require.Equal(t, sdk.NewDec(13), resBond.Shares)
	resBond, found = app.StakingKeeper.GetDelegation(ctx, addrDels[1], valAddrs[1])
	require.True(t, found)
	require.Equal(t, sdk.NewDec(7), resBond.Shares)

	// stake more alongside redelegation
	bond1to2, found = app.StakingKeeper.GetDelegation(ctx, addrDels[0], valAddrs[1])
	require.True(t, found)
	require.Equal(t, sdk.NewDec(13), bond1to2.Shares)
	bond1to2.Shares = sdk.NewDec(47) // add 34 shares
	app.StakingKeeper.SetDelegation(ctx, bond1to2)

	// full transfer from partial redelegation
	transferred = app.StakingKeeper.TransferDelegation(ctx, addrDels[0], addrDels[1], valAddrs[1], sdk.NewDec(9999))
	require.Equal(t, sdk.NewDec(47), transferred)
	resBond, found = app.StakingKeeper.GetDelegation(ctx, addrDels[0], valAddrs[1])
	require.False(t, found)
	resBond, found = app.StakingKeeper.GetDelegation(ctx, addrDels[1], valAddrs[1])
	require.True(t, found)
	require.Equal(t, sdk.NewDec(54), resBond.Shares)
}

// tests Get/Set/Remove UnbondingDelegation
func (s *KeeperTestSuite) TestUnbondingDelegation() {
	ctx, keeper := s.ctx, s.stakingKeeper
	require := s.Require()

	delAddrs, valAddrs := createValAddrs(2)

	ubd := stakingtypes.NewUnbondingDelegation(
		delAddrs[0],
		valAddrs[0],
		0,
		time.Unix(0, 0).UTC(),
		sdk.NewInt(5),
		0,
	)

	// set and retrieve a record
	keeper.SetUnbondingDelegation(ctx, ubd)
	resUnbond, found := keeper.GetUnbondingDelegation(ctx, delAddrs[0], valAddrs[0])
	require.True(found)
	require.Equal(ubd, resUnbond)

	// modify a records, save, and retrieve
	expUnbond := sdk.NewInt(21)
	ubd.Entries[0].Balance = expUnbond
	keeper.SetUnbondingDelegation(ctx, ubd)

	resUnbonds := keeper.GetUnbondingDelegations(ctx, delAddrs[0], 5)
	require.Equal(1, len(resUnbonds))

	resUnbonds = keeper.GetAllUnbondingDelegations(ctx, delAddrs[0])
	require.Equal(1, len(resUnbonds))

	resUnbond, found = keeper.GetUnbondingDelegation(ctx, delAddrs[0], valAddrs[0])
	require.True(found)
	require.Equal(ubd, resUnbond)

	resDelUnbond := keeper.GetDelegatorUnbonding(ctx, delAddrs[0])
	require.Equal(expUnbond, resDelUnbond)

	// delete a record
	keeper.RemoveUnbondingDelegation(ctx, ubd)
	_, found = keeper.GetUnbondingDelegation(ctx, delAddrs[0], valAddrs[0])
	require.False(found)

	resUnbonds = keeper.GetUnbondingDelegations(ctx, delAddrs[0], 5)
	require.Equal(0, len(resUnbonds))

	resUnbonds = keeper.GetAllUnbondingDelegations(ctx, delAddrs[0])
	require.Equal(0, len(resUnbonds))
}

func (s *KeeperTestSuite) TestTransferUnbonding() {
	ctx, keeper := s.ctx, s.stakingKeeper
	require := s.Require()

	delAddrs := simapp.AddTestAddrsIncremental(app, ctx, 2, sdk.NewInt(10000))
	valAddrs := simapp.ConvertAddrsToValAddrs(delAddrs)

	// try to transfer when there's nothing
	transferred := app.StakingKeeper.TransferUnbonding(ctx, delAddrs[0], delAddrs[1], valAddrs[0], sdk.NewInt(30))
	require.Equal(t, sdk.ZeroInt(), transferred)
	_, found := app.StakingKeeper.GetUnbondingDelegation(ctx, delAddrs[1], valAddrs[0])
	require.False(t, found)

	// set an UnbondingDelegation with one entry
	ubd := types.NewUnbondingDelegation(
		delAddrs[0],
		valAddrs[0],
		0,
		time.Unix(0, 0).UTC(),
		sdk.NewInt(5),
	)
	app.StakingKeeper.SetUnbondingDelegation(ctx, ubd)

	// transfer nothing
	transferred = app.StakingKeeper.TransferUnbonding(ctx, delAddrs[0], delAddrs[1], valAddrs[0], sdk.ZeroInt())
	require.Equal(t, sdk.ZeroInt(), transferred)

	// partial transfer
	transferred = app.StakingKeeper.TransferUnbonding(ctx, delAddrs[0], delAddrs[1], valAddrs[0], sdk.NewInt(3))
	require.Equal(t, sdk.NewInt(3), transferred)
	ubd.Entries[0].Balance = sdk.NewInt(2)
	resUnbond, found := app.StakingKeeper.GetUnbondingDelegation(ctx, delAddrs[0], valAddrs[0])
	require.True(t, found)
	require.Equal(t, ubd, resUnbond)
	resUnbond, found = app.StakingKeeper.GetUnbondingDelegation(ctx, delAddrs[1], valAddrs[0])
	require.True(t, found)
	wantDestUnbond := types.NewUnbondingDelegation(
		delAddrs[1],
		valAddrs[0],
		0,
		time.Unix(0, 0).UTC(),
		sdk.NewInt(3),
	)
	require.Equal(t, wantDestUnbond, resUnbond)

	// add another entry
	completionTime := time.Unix(3600, 0).UTC()
	ubdTo := app.StakingKeeper.SetUnbondingDelegationEntry(ctx, delAddrs[0], valAddrs[0], 1, completionTime, sdk.NewInt(57))
	app.StakingKeeper.InsertUBDQueue(ctx, ubdTo, completionTime)

	// full transfer
	transferred = app.StakingKeeper.TransferUnbonding(ctx, delAddrs[0], delAddrs[1], valAddrs[0], sdk.NewInt(999))
	require.Equal(t, sdk.NewInt(59), transferred)
	_, found = app.StakingKeeper.GetUnbondingDelegation(ctx, delAddrs[0], valAddrs[0])
	require.False(t, found)
	resUnbond, found = app.StakingKeeper.GetUnbondingDelegation(ctx, delAddrs[1], valAddrs[0])
	require.True(t, found)
	require.Equal(t, 3, len(resUnbond.Entries))
	require.Equal(t, sdk.NewInt(3), resUnbond.Entries[0].Balance)
	require.Equal(t, sdk.NewInt(2), resUnbond.Entries[1].Balance)
	require.Equal(t, sdk.NewInt(57), resUnbond.Entries[2].Balance)
}

func (s *KeeperTestSuite) TestUnbondDelegation() {
	ctx, keeper := s.ctx, s.stakingKeeper
	require := s.Require()

	delAddrs, valAddrs := createValAddrs(1)
	startTokens := keeper.TokensFromConsensusPower(ctx, 10)
	validator := testutil.NewValidator(s.T(), valAddrs[0], PKs[0])

	validator, issuedShares := validator.AddTokensFromDel(startTokens)
	require.Equal(startTokens, issuedShares.RoundInt())

	s.bankKeeper.EXPECT().SendCoinsFromModuleToModule(gomock.Any(), stakingtypes.NotBondedPoolName, stakingtypes.BondedPoolName, gomock.Any())
	_ = stakingkeeper.TestingUpdateValidator(keeper, ctx, validator, true)

	delegation := stakingtypes.NewDelegation(delAddrs[0], valAddrs[0], issuedShares)
	keeper.SetDelegation(ctx, delegation)

	bondTokens := keeper.TokensFromConsensusPower(ctx, 6)
	amount, err := keeper.Unbond(ctx, delAddrs[0], valAddrs[0], sdk.NewDecFromInt(bondTokens))
	require.NoError(err)
	require.Equal(bondTokens, amount) // shares to be added to an unbonding delegation

	delegation, found := keeper.GetDelegation(ctx, delAddrs[0], valAddrs[0])
	require.True(found)
	validator, found = keeper.GetValidator(ctx, valAddrs[0])
	require.True(found)

	remainingTokens := startTokens.Sub(bondTokens)

	require.Equal(remainingTokens, delegation.Shares.RoundInt())
	require.Equal(remainingTokens, validator.BondedTokens())
}

// // test undelegating self delegation from a validator pushing it below MinSelfDelegation
// // shift it from the bonded to unbonding state and jailed
func (s *KeeperTestSuite) TestUndelegateSelfDelegationBelowMinSelfDelegation() {
	ctx, keeper := s.ctx, s.stakingKeeper
	require := s.Require()

	addrDels, addrVals := createValAddrs(1)
	delTokens := keeper.TokensFromConsensusPower(ctx, 10)

	// create a validator with a self-delegation
	validator := testutil.NewValidator(s.T(), addrVals[0], PKs[0])

	validator.MinSelfDelegation = delTokens
	validator, issuedShares := validator.AddTokensFromDel(delTokens)
	require.Equal(delTokens, issuedShares.RoundInt())

	s.bankKeeper.EXPECT().SendCoinsFromModuleToModule(gomock.Any(), stakingtypes.NotBondedPoolName, stakingtypes.BondedPoolName, gomock.Any())
	validator = stakingkeeper.TestingUpdateValidator(keeper, ctx, validator, true)
	keeper.SetValidatorByConsAddr(ctx, validator)
	require.True(validator.IsBonded())

	selfDelegation := stakingtypes.NewDelegation(sdk.AccAddress(addrVals[0].Bytes()), addrVals[0], issuedShares)
	keeper.SetDelegation(ctx, selfDelegation)

	// create a second delegation to this validator
	keeper.DeleteValidatorByPowerIndex(ctx, validator)
	validator, issuedShares = validator.AddTokensFromDel(delTokens)
	require.True(validator.IsBonded())
	require.Equal(delTokens, issuedShares.RoundInt())

	validator = stakingkeeper.TestingUpdateValidator(keeper, ctx, validator, true)
	delegation := stakingtypes.NewDelegation(addrDels[0], addrVals[0], issuedShares)
	keeper.SetDelegation(ctx, delegation)

	val0AccAddr := sdk.AccAddress(addrVals[0].Bytes())
	s.bankKeeper.EXPECT().SendCoinsFromModuleToModule(gomock.Any(), stakingtypes.BondedPoolName, stakingtypes.NotBondedPoolName, gomock.Any())
	_, err := keeper.Undelegate(ctx, val0AccAddr, addrVals[0], sdk.NewDecFromInt(keeper.TokensFromConsensusPower(ctx, 6)))
	require.NoError(err)

	// end block
	s.bankKeeper.EXPECT().SendCoinsFromModuleToModule(gomock.Any(), stakingtypes.BondedPoolName, stakingtypes.NotBondedPoolName, gomock.Any())
	s.applyValidatorSetUpdates(ctx, keeper, 1)

	validator, found := keeper.GetValidator(ctx, addrVals[0])
	require.True(found)
	require.Equal(keeper.TokensFromConsensusPower(ctx, 14), validator.Tokens)
	require.Equal(stakingtypes.Unbonding, validator.Status)
	require.True(validator.Jailed)
}

func (s *KeeperTestSuite) TestUndelegateFromUnbondingValidator() {
	ctx, keeper := s.ctx, s.stakingKeeper
	require := s.Require()
	delTokens := keeper.TokensFromConsensusPower(ctx, 10)

	addrDels, addrVals := createValAddrs(2)

	// create a validator with a self-delegation
	validator := testutil.NewValidator(s.T(), addrVals[0], PKs[0])
	keeper.SetValidatorByConsAddr(ctx, validator)

	validator, issuedShares := validator.AddTokensFromDel(delTokens)
	require.Equal(delTokens, issuedShares.RoundInt())

	s.bankKeeper.EXPECT().SendCoinsFromModuleToModule(gomock.Any(), stakingtypes.NotBondedPoolName, stakingtypes.BondedPoolName, gomock.Any())
	validator = stakingkeeper.TestingUpdateValidator(keeper, ctx, validator, true)
	require.True(validator.IsBonded())

	selfDelegation := stakingtypes.NewDelegation(addrVals[0].Bytes(), addrVals[0], issuedShares)
	keeper.SetDelegation(ctx, selfDelegation)

	// create a second delegation to this validator
	keeper.DeleteValidatorByPowerIndex(ctx, validator)

	validator, issuedShares = validator.AddTokensFromDel(delTokens)
	require.Equal(delTokens, issuedShares.RoundInt())

	validator = stakingkeeper.TestingUpdateValidator(keeper, ctx, validator, true)
	delegation := stakingtypes.NewDelegation(addrDels[1], addrVals[0], issuedShares)
	keeper.SetDelegation(ctx, delegation)

	header := ctx.BlockHeader()
	blockHeight := int64(10)
	header.Height = blockHeight
	blockTime := time.Unix(333, 0)
	header.Time = blockTime
	ctx = ctx.WithBlockHeader(header)

	// unbond the all self-delegation to put validator in unbonding state
	val0AccAddr := sdk.AccAddress(addrVals[0])
	s.bankKeeper.EXPECT().SendCoinsFromModuleToModule(gomock.Any(), stakingtypes.BondedPoolName, stakingtypes.NotBondedPoolName, gomock.Any())
	_, err := keeper.Undelegate(ctx, val0AccAddr, addrVals[0], sdk.NewDecFromInt(delTokens))
	require.NoError(err)

	// end block
	s.bankKeeper.EXPECT().SendCoinsFromModuleToModule(gomock.Any(), stakingtypes.BondedPoolName, stakingtypes.NotBondedPoolName, gomock.Any())
	s.applyValidatorSetUpdates(ctx, keeper, 1)

	validator, found := keeper.GetValidator(ctx, addrVals[0])
	require.True(found)
	require.Equal(blockHeight, validator.UnbondingHeight)
	params := keeper.GetParams(ctx)
	require.True(blockTime.Add(params.UnbondingTime).Equal(validator.UnbondingTime))

	blockHeight2 := int64(20)
	blockTime2 := time.Unix(444, 0).UTC()
	ctx = ctx.WithBlockHeight(blockHeight2)
	ctx = ctx.WithBlockTime(blockTime2)

	// unbond some of the other delegation's shares
	_, err = keeper.Undelegate(ctx, addrDels[1], addrVals[0], math.LegacyNewDec(6))
	require.NoError(err)

	// retrieve the unbonding delegation
	ubd, found := keeper.GetUnbondingDelegation(ctx, addrDels[1], addrVals[0])
	require.True(found)
	require.Len(ubd.Entries, 1)
	require.True(ubd.Entries[0].Balance.Equal(sdk.NewInt(6)))
	require.Equal(blockHeight2, ubd.Entries[0].CreationHeight)
	require.True(blockTime2.Add(params.UnbondingTime).Equal(ubd.Entries[0].CompletionTime))
}

func (s *KeeperTestSuite) TestUndelegateFromUnbondedValidator() {
	ctx, keeper := s.ctx, s.stakingKeeper
	require := s.Require()

	delTokens := keeper.TokensFromConsensusPower(ctx, 10)
	addrDels, addrVals := createValAddrs(2)

	// create a validator with a self-delegation
	validator := testutil.NewValidator(s.T(), addrVals[0], PKs[0])
	keeper.SetValidatorByConsAddr(ctx, validator)

	valTokens := keeper.TokensFromConsensusPower(ctx, 10)
	validator, issuedShares := validator.AddTokensFromDel(valTokens)
	require.Equal(valTokens, issuedShares.RoundInt())
	s.bankKeeper.EXPECT().SendCoinsFromModuleToModule(gomock.Any(), stakingtypes.NotBondedPoolName, stakingtypes.BondedPoolName, gomock.Any())
	validator = stakingkeeper.TestingUpdateValidator(keeper, ctx, validator, true)
	require.True(validator.IsBonded())

	val0AccAddr := sdk.AccAddress(addrVals[0])
	selfDelegation := stakingtypes.NewDelegation(val0AccAddr, addrVals[0], issuedShares)
	keeper.SetDelegation(ctx, selfDelegation)

	// create a second delegation to this validator
	keeper.DeleteValidatorByPowerIndex(ctx, validator)
	validator, issuedShares = validator.AddTokensFromDel(delTokens)
	require.Equal(delTokens, issuedShares.RoundInt())
	validator = stakingkeeper.TestingUpdateValidator(keeper, ctx, validator, true)
	require.True(validator.IsBonded())
	delegation := stakingtypes.NewDelegation(addrDels[1], addrVals[0], issuedShares)
	keeper.SetDelegation(ctx, delegation)

	ctx = ctx.WithBlockHeight(10)
	ctx = ctx.WithBlockTime(time.Unix(333, 0))

	// unbond the all self-delegation to put validator in unbonding state
	s.bankKeeper.EXPECT().SendCoinsFromModuleToModule(gomock.Any(), stakingtypes.BondedPoolName, stakingtypes.NotBondedPoolName, gomock.Any())
	_, err := keeper.Undelegate(ctx, val0AccAddr, addrVals[0], sdk.NewDecFromInt(valTokens))
	require.NoError(err)

	// end block
	s.bankKeeper.EXPECT().SendCoinsFromModuleToModule(gomock.Any(), stakingtypes.BondedPoolName, stakingtypes.NotBondedPoolName, gomock.Any())
	s.applyValidatorSetUpdates(ctx, keeper, 1)

	validator, found := keeper.GetValidator(ctx, addrVals[0])
	require.True(found)
	require.Equal(ctx.BlockHeight(), validator.UnbondingHeight)
	params := keeper.GetParams(ctx)
	require.True(ctx.BlockHeader().Time.Add(params.UnbondingTime).Equal(validator.UnbondingTime))

	// unbond the validator
	ctx = ctx.WithBlockTime(validator.UnbondingTime)
	keeper.UnbondAllMatureValidators(ctx)

	// Make sure validator is still in state because there is still an outstanding delegation
	validator, found = keeper.GetValidator(ctx, addrVals[0])
	require.True(found)
	require.Equal(validator.Status, stakingtypes.Unbonded)

	// unbond some of the other delegation's shares
	unbondTokens := keeper.TokensFromConsensusPower(ctx, 6)
	_, err = keeper.Undelegate(ctx, addrDels[1], addrVals[0], sdk.NewDecFromInt(unbondTokens))
	require.NoError(err)

	// unbond rest of the other delegation's shares
	remainingTokens := delTokens.Sub(unbondTokens)
	_, err = keeper.Undelegate(ctx, addrDels[1], addrVals[0], sdk.NewDecFromInt(remainingTokens))
	require.NoError(err)

	//  now validator should be deleted from state
	validator, found = keeper.GetValidator(ctx, addrVals[0])
	require.False(found, "%v", validator)
}

func (s *KeeperTestSuite) TestUnbondingAllDelegationFromValidator() {
	ctx, keeper := s.ctx, s.stakingKeeper
	require := s.Require()

	delTokens := keeper.TokensFromConsensusPower(ctx, 10)
	addrDels, addrVals := createValAddrs(2)

	// create a validator with a self-delegation
	validator := testutil.NewValidator(s.T(), addrVals[0], PKs[0])
	keeper.SetValidatorByConsAddr(ctx, validator)

	valTokens := keeper.TokensFromConsensusPower(ctx, 10)
	validator, issuedShares := validator.AddTokensFromDel(valTokens)
	require.Equal(valTokens, issuedShares.RoundInt())

	s.bankKeeper.EXPECT().SendCoinsFromModuleToModule(gomock.Any(), stakingtypes.NotBondedPoolName, stakingtypes.BondedPoolName, gomock.Any())
	validator = stakingkeeper.TestingUpdateValidator(keeper, ctx, validator, true)
	require.True(validator.IsBonded())
	val0AccAddr := sdk.AccAddress(addrVals[0].Bytes())

	selfDelegation := stakingtypes.NewDelegation(val0AccAddr, addrVals[0], issuedShares)
	keeper.SetDelegation(ctx, selfDelegation)

	// create a second delegation to this validator
	keeper.DeleteValidatorByPowerIndex(ctx, validator)
	validator, issuedShares = validator.AddTokensFromDel(delTokens)
	require.Equal(delTokens, issuedShares.RoundInt())

	validator = stakingkeeper.TestingUpdateValidator(keeper, ctx, validator, true)
	require.True(validator.IsBonded())

	delegation := stakingtypes.NewDelegation(addrDels[1], addrVals[0], issuedShares)
	keeper.SetDelegation(ctx, delegation)

	ctx = ctx.WithBlockHeight(10)
	ctx = ctx.WithBlockTime(time.Unix(333, 0))

	// unbond the all self-delegation to put validator in unbonding state
	s.bankKeeper.EXPECT().SendCoinsFromModuleToModule(gomock.Any(), stakingtypes.BondedPoolName, stakingtypes.NotBondedPoolName, gomock.Any())
	_, err := keeper.Undelegate(ctx, val0AccAddr, addrVals[0], sdk.NewDecFromInt(valTokens))
	require.NoError(err)

	// end block
	s.bankKeeper.EXPECT().SendCoinsFromModuleToModule(gomock.Any(), stakingtypes.BondedPoolName, stakingtypes.NotBondedPoolName, gomock.Any())
	s.applyValidatorSetUpdates(ctx, keeper, 1)

	// unbond all the remaining delegation
	_, err = keeper.Undelegate(ctx, addrDels[1], addrVals[0], sdk.NewDecFromInt(delTokens))
	require.NoError(err)

	// validator should still be in state and still be in unbonding state
	validator, found := keeper.GetValidator(ctx, addrVals[0])
	require.True(found)
	require.Equal(validator.Status, stakingtypes.Unbonding)

	// unbond the validator
	ctx = ctx.WithBlockTime(validator.UnbondingTime)
	keeper.UnbondAllMatureValidators(ctx)

	// validator should now be deleted from state
	_, found = keeper.GetValidator(ctx, addrVals[0])
	require.False(found)
}

// Make sure that that the retrieving the delegations doesn't affect the state
func (s *KeeperTestSuite) TestGetRedelegationsFromSrcValidator() {
	ctx, keeper := s.ctx, s.stakingKeeper
	require := s.Require()

	addrDels, addrVals := createValAddrs(2)

	rd := stakingtypes.NewRedelegation(addrDels[0], addrVals[0], addrVals[1], 0,
		time.Unix(0, 0), sdk.NewInt(5),
		math.LegacyNewDec(5), 0)

	// set and retrieve a record
	keeper.SetRedelegation(ctx, rd)
	resBond, found := keeper.GetRedelegation(ctx, addrDels[0], addrVals[0], addrVals[1])
	require.True(found)

	// get the redelegations one time
	redelegations := keeper.GetRedelegationsFromSrcValidator(ctx, addrVals[0])
	require.Equal(1, len(redelegations))
	require.Equal(redelegations[0], resBond)

	// get the redelegations a second time, should be exactly the same
	redelegations = keeper.GetRedelegationsFromSrcValidator(ctx, addrVals[0])
	require.Equal(1, len(redelegations))
	require.Equal(redelegations[0], resBond)
}

// tests Get/Set/Remove/Has UnbondingDelegation
func (s *KeeperTestSuite) TestRedelegation() {
	ctx, keeper := s.ctx, s.stakingKeeper
	require := s.Require()

	addrDels, addrVals := createValAddrs(2)

	rd := stakingtypes.NewRedelegation(addrDels[0], addrVals[0], addrVals[1], 0,
		time.Unix(0, 0).UTC(), sdk.NewInt(5),
		math.LegacyNewDec(5), 0)

	// test shouldn't have and redelegations
	has := keeper.HasReceivingRedelegation(ctx, addrDels[0], addrVals[1])
	require.False(has)

	// set and retrieve a record
	keeper.SetRedelegation(ctx, rd)
	resRed, found := keeper.GetRedelegation(ctx, addrDels[0], addrVals[0], addrVals[1])
	require.True(found)

	redelegations := keeper.GetRedelegationsFromSrcValidator(ctx, addrVals[0])
	require.Equal(1, len(redelegations))
	require.Equal(redelegations[0], resRed)

	redelegations = keeper.GetRedelegations(ctx, addrDels[0], 5)
	require.Equal(1, len(redelegations))
	require.Equal(redelegations[0], resRed)

	redelegations = keeper.GetAllRedelegations(ctx, addrDels[0], nil, nil)
	require.Equal(1, len(redelegations))
	require.Equal(redelegations[0], resRed)

	// check if has the redelegation
	has = keeper.HasReceivingRedelegation(ctx, addrDels[0], addrVals[1])
	require.True(has)

	// modify a records, save, and retrieve
	rd.Entries[0].SharesDst = math.LegacyNewDec(21)
	keeper.SetRedelegation(ctx, rd)

	resRed, found = keeper.GetRedelegation(ctx, addrDels[0], addrVals[0], addrVals[1])
	require.True(found)
	require.Equal(rd, resRed)

	redelegations = keeper.GetRedelegationsFromSrcValidator(ctx, addrVals[0])
	require.Equal(1, len(redelegations))
	require.Equal(redelegations[0], resRed)

	redelegations = keeper.GetRedelegations(ctx, addrDels[0], 5)
	require.Equal(1, len(redelegations))
	require.Equal(redelegations[0], resRed)

	// delete a record
	keeper.RemoveRedelegation(ctx, rd)
	_, found = keeper.GetRedelegation(ctx, addrDels[0], addrVals[0], addrVals[1])
	require.False(found)

	redelegations = keeper.GetRedelegations(ctx, addrDels[0], 5)
	require.Equal(0, len(redelegations))

	redelegations = keeper.GetAllRedelegations(ctx, addrDels[0], nil, nil)
	require.Equal(0, len(redelegations))
}

func (s *KeeperTestSuite) TestRedelegateToSameValidator() {
	ctx, keeper := s.ctx, s.stakingKeeper
	require := s.Require()

	_, addrVals := createValAddrs(1)
	valTokens := keeper.TokensFromConsensusPower(ctx, 10)

	// create a validator with a self-delegation
	validator := testutil.NewValidator(s.T(), addrVals[0], PKs[0])
	validator, issuedShares := validator.AddTokensFromDel(valTokens)
	require.Equal(valTokens, issuedShares.RoundInt())

	s.bankKeeper.EXPECT().SendCoinsFromModuleToModule(gomock.Any(), stakingtypes.NotBondedPoolName, stakingtypes.BondedPoolName, gomock.Any())
	validator = stakingkeeper.TestingUpdateValidator(keeper, ctx, validator, true)
	require.True(validator.IsBonded())

	val0AccAddr := sdk.AccAddress(addrVals[0].Bytes())
	selfDelegation := stakingtypes.NewDelegation(val0AccAddr, addrVals[0], issuedShares)
	keeper.SetDelegation(ctx, selfDelegation)

	_, err := keeper.BeginRedelegation(ctx, val0AccAddr, addrVals[0], addrVals[0], math.LegacyNewDec(5))
	require.Error(err)
}

func (s *KeeperTestSuite) TestRedelegationMaxEntries() {
	ctx, keeper := s.ctx, s.stakingKeeper
	require := s.Require()

	_, addrVals := createValAddrs(2)

	// create a validator with a self-delegation
	validator := testutil.NewValidator(s.T(), addrVals[0], PKs[0])
	valTokens := keeper.TokensFromConsensusPower(ctx, 10)
	validator, issuedShares := validator.AddTokensFromDel(valTokens)
	require.Equal(valTokens, issuedShares.RoundInt())

	s.bankKeeper.EXPECT().SendCoinsFromModuleToModule(gomock.Any(), stakingtypes.NotBondedPoolName, stakingtypes.BondedPoolName, gomock.Any())
	_ = stakingkeeper.TestingUpdateValidator(keeper, ctx, validator, true)
	val0AccAddr := sdk.AccAddress(addrVals[0].Bytes())
	selfDelegation := stakingtypes.NewDelegation(val0AccAddr, addrVals[0], issuedShares)
	keeper.SetDelegation(ctx, selfDelegation)

	// create a second validator
	validator2 := testutil.NewValidator(s.T(), addrVals[1], PKs[1])
	validator2, issuedShares = validator2.AddTokensFromDel(valTokens)
	require.Equal(valTokens, issuedShares.RoundInt())

	s.bankKeeper.EXPECT().SendCoinsFromModuleToModule(gomock.Any(), stakingtypes.NotBondedPoolName, stakingtypes.BondedPoolName, gomock.Any())
	validator2 = stakingkeeper.TestingUpdateValidator(keeper, ctx, validator2, true)
	require.Equal(stakingtypes.Bonded, validator2.Status)

	maxEntries := keeper.MaxEntries(ctx)

	// redelegations should pass
	var completionTime time.Time
	for i := uint32(0); i < maxEntries; i++ {
		var err error
		completionTime, err = keeper.BeginRedelegation(ctx, val0AccAddr, addrVals[0], addrVals[1], math.LegacyNewDec(1))
		require.NoError(err)
	}

	// an additional redelegation should fail due to max entries
	_, err := keeper.BeginRedelegation(ctx, val0AccAddr, addrVals[0], addrVals[1], math.LegacyNewDec(1))
	require.Error(err)

	// mature redelegations
	ctx = ctx.WithBlockTime(completionTime)
	_, err = keeper.CompleteRedelegation(ctx, val0AccAddr, addrVals[0], addrVals[1])
	require.NoError(err)

	// redelegation should work again
	_, err = keeper.BeginRedelegation(ctx, val0AccAddr, addrVals[0], addrVals[1], math.LegacyNewDec(1))
	require.NoError(err)
}

func (s *KeeperTestSuite) TestRedelegateSelfDelegation() {
	ctx, keeper := s.ctx, s.stakingKeeper
	require := s.Require()

	addrDels, addrVals := createValAddrs(2)

	// create a validator with a self-delegation
	validator := testutil.NewValidator(s.T(), addrVals[0], PKs[0])
	keeper.SetValidatorByConsAddr(ctx, validator)

	valTokens := keeper.TokensFromConsensusPower(ctx, 10)
	validator, issuedShares := validator.AddTokensFromDel(valTokens)
	require.Equal(valTokens, issuedShares.RoundInt())

	s.bankKeeper.EXPECT().SendCoinsFromModuleToModule(gomock.Any(), stakingtypes.NotBondedPoolName, stakingtypes.BondedPoolName, gomock.Any())
	validator = stakingkeeper.TestingUpdateValidator(keeper, ctx, validator, true)

	val0AccAddr := sdk.AccAddress(addrVals[0])
	selfDelegation := stakingtypes.NewDelegation(val0AccAddr, addrVals[0], issuedShares)
	keeper.SetDelegation(ctx, selfDelegation)

	// create a second validator
	validator2 := testutil.NewValidator(s.T(), addrVals[1], PKs[1])
	validator2, issuedShares = validator2.AddTokensFromDel(valTokens)
	require.Equal(valTokens, issuedShares.RoundInt())
	s.bankKeeper.EXPECT().SendCoinsFromModuleToModule(gomock.Any(), stakingtypes.NotBondedPoolName, stakingtypes.BondedPoolName, gomock.Any())
	validator2 = stakingkeeper.TestingUpdateValidator(keeper, ctx, validator2, true)
	require.Equal(stakingtypes.Bonded, validator2.Status)

	// create a second delegation to validator 1
	delTokens := keeper.TokensFromConsensusPower(ctx, 10)
	validator, issuedShares = validator.AddTokensFromDel(delTokens)
	require.Equal(delTokens, issuedShares.RoundInt())
	validator = stakingkeeper.TestingUpdateValidator(keeper, ctx, validator, true)

	delegation := stakingtypes.NewDelegation(addrDels[0], addrVals[0], issuedShares)
	keeper.SetDelegation(ctx, delegation)

	_, err := keeper.BeginRedelegation(ctx, val0AccAddr, addrVals[0], addrVals[1], sdk.NewDecFromInt(delTokens))
	require.NoError(err)

	// end block
	s.bankKeeper.EXPECT().SendCoinsFromModuleToModule(gomock.Any(), stakingtypes.BondedPoolName, stakingtypes.NotBondedPoolName, gomock.Any())
	s.applyValidatorSetUpdates(ctx, keeper, 2)

	validator, found := keeper.GetValidator(ctx, addrVals[0])
	require.True(found)
	require.Equal(valTokens, validator.Tokens)
	require.Equal(stakingtypes.Unbonding, validator.Status)
}

func (s *KeeperTestSuite) TestRedelegateFromUnbondingValidator() {
	ctx, keeper := s.ctx, s.stakingKeeper
	require := s.Require()

	addrDels, addrVals := createValAddrs(2)

	// create a validator with a self-delegation
	validator := testutil.NewValidator(s.T(), addrVals[0], PKs[0])
	keeper.SetValidatorByConsAddr(ctx, validator)

	valTokens := keeper.TokensFromConsensusPower(ctx, 10)
	validator, issuedShares := validator.AddTokensFromDel(valTokens)
	require.Equal(valTokens, issuedShares.RoundInt())
	s.bankKeeper.EXPECT().SendCoinsFromModuleToModule(gomock.Any(), stakingtypes.NotBondedPoolName, stakingtypes.BondedPoolName, gomock.Any())
	validator = stakingkeeper.TestingUpdateValidator(keeper, ctx, validator, true)
	val0AccAddr := sdk.AccAddress(addrVals[0].Bytes())
	selfDelegation := stakingtypes.NewDelegation(val0AccAddr, addrVals[0], issuedShares)
	keeper.SetDelegation(ctx, selfDelegation)

	// create a second delegation to this validator
	keeper.DeleteValidatorByPowerIndex(ctx, validator)
	delTokens := keeper.TokensFromConsensusPower(ctx, 10)
	validator, issuedShares = validator.AddTokensFromDel(delTokens)
	require.Equal(delTokens, issuedShares.RoundInt())
	validator = stakingkeeper.TestingUpdateValidator(keeper, ctx, validator, true)
	delegation := stakingtypes.NewDelegation(addrDels[1], addrVals[0], issuedShares)
	keeper.SetDelegation(ctx, delegation)

	// create a second validator
	validator2 := testutil.NewValidator(s.T(), addrVals[1], PKs[1])
	validator2, issuedShares = validator2.AddTokensFromDel(valTokens)
	require.Equal(valTokens, issuedShares.RoundInt())
	s.bankKeeper.EXPECT().SendCoinsFromModuleToModule(gomock.Any(), stakingtypes.NotBondedPoolName, stakingtypes.BondedPoolName, gomock.Any())
	_ = stakingkeeper.TestingUpdateValidator(keeper, ctx, validator2, true)

	header := ctx.BlockHeader()
	blockHeight := int64(10)
	header.Height = blockHeight
	blockTime := time.Unix(333, 0)
	header.Time = blockTime
	ctx = ctx.WithBlockHeader(header)

	// unbond the all self-delegation to put validator in unbonding state
	s.bankKeeper.EXPECT().SendCoinsFromModuleToModule(gomock.Any(), stakingtypes.BondedPoolName, stakingtypes.NotBondedPoolName, gomock.Any())
	_, err := keeper.Undelegate(ctx, val0AccAddr, addrVals[0], sdk.NewDecFromInt(delTokens))
	require.NoError(err)

	// end block
	s.bankKeeper.EXPECT().SendCoinsFromModuleToModule(gomock.Any(), stakingtypes.BondedPoolName, stakingtypes.NotBondedPoolName, gomock.Any())
	s.applyValidatorSetUpdates(ctx, keeper, 1)

	validator, found := keeper.GetValidator(ctx, addrVals[0])
	require.True(found)
	require.Equal(blockHeight, validator.UnbondingHeight)
	params := keeper.GetParams(ctx)
	require.True(blockTime.Add(params.UnbondingTime).Equal(validator.UnbondingTime))

	// change the context
	header = ctx.BlockHeader()
	blockHeight2 := int64(20)
	header.Height = blockHeight2
	blockTime2 := time.Unix(444, 0)
	header.Time = blockTime2
	ctx = ctx.WithBlockHeader(header)

	// unbond some of the other delegation's shares
	redelegateTokens := keeper.TokensFromConsensusPower(ctx, 6)
	s.bankKeeper.EXPECT().SendCoinsFromModuleToModule(gomock.Any(), stakingtypes.NotBondedPoolName, stakingtypes.BondedPoolName, gomock.Any())
	_, err = keeper.BeginRedelegation(ctx, addrDels[1], addrVals[0], addrVals[1], sdk.NewDecFromInt(redelegateTokens))
	require.NoError(err)

	// retrieve the unbonding delegation
	ubd, found := keeper.GetRedelegation(ctx, addrDels[1], addrVals[0], addrVals[1])
	require.True(found)
	require.Len(ubd.Entries, 1)
	require.Equal(blockHeight, ubd.Entries[0].CreationHeight)
	require.True(blockTime.Add(params.UnbondingTime).Equal(ubd.Entries[0].CompletionTime))
}

func (s *KeeperTestSuite) TestRedelegateFromUnbondedValidator() {
	ctx, keeper := s.ctx, s.stakingKeeper
	require := s.Require()

	addrDels, addrVals := createValAddrs(2)

	// create a validator with a self-delegation
	validator := testutil.NewValidator(s.T(), addrVals[0], PKs[0])
	keeper.SetValidatorByConsAddr(ctx, validator)

	valTokens := keeper.TokensFromConsensusPower(ctx, 10)
	validator, issuedShares := validator.AddTokensFromDel(valTokens)
	require.Equal(valTokens, issuedShares.RoundInt())
	s.bankKeeper.EXPECT().SendCoinsFromModuleToModule(gomock.Any(), stakingtypes.NotBondedPoolName, stakingtypes.BondedPoolName, gomock.Any())
	validator = stakingkeeper.TestingUpdateValidator(keeper, ctx, validator, true)
	val0AccAddr := sdk.AccAddress(addrVals[0].Bytes())
	selfDelegation := stakingtypes.NewDelegation(val0AccAddr, addrVals[0], issuedShares)
	keeper.SetDelegation(ctx, selfDelegation)

	// create a second delegation to this validator
	keeper.DeleteValidatorByPowerIndex(ctx, validator)
	delTokens := keeper.TokensFromConsensusPower(ctx, 10)
	validator, issuedShares = validator.AddTokensFromDel(delTokens)
	require.Equal(delTokens, issuedShares.RoundInt())
	validator = stakingkeeper.TestingUpdateValidator(keeper, ctx, validator, true)
	delegation := stakingtypes.NewDelegation(addrDels[1], addrVals[0], issuedShares)
	keeper.SetDelegation(ctx, delegation)

	// create a second validator
	validator2 := testutil.NewValidator(s.T(), addrVals[1], PKs[1])
	validator2, issuedShares = validator2.AddTokensFromDel(valTokens)
	require.Equal(valTokens, issuedShares.RoundInt())
	s.bankKeeper.EXPECT().SendCoinsFromModuleToModule(gomock.Any(), stakingtypes.NotBondedPoolName, stakingtypes.BondedPoolName, gomock.Any())
	validator2 = stakingkeeper.TestingUpdateValidator(keeper, ctx, validator2, true)
	require.Equal(stakingtypes.Bonded, validator2.Status)

	ctx = ctx.WithBlockHeight(10)
	ctx = ctx.WithBlockTime(time.Unix(333, 0))

	// unbond the all self-delegation to put validator in unbonding state
	s.bankKeeper.EXPECT().SendCoinsFromModuleToModule(gomock.Any(), stakingtypes.BondedPoolName, stakingtypes.NotBondedPoolName, gomock.Any())
	_, err := keeper.Undelegate(ctx, val0AccAddr, addrVals[0], sdk.NewDecFromInt(delTokens))
	require.NoError(err)

	// end block
	s.bankKeeper.EXPECT().SendCoinsFromModuleToModule(gomock.Any(), stakingtypes.BondedPoolName, stakingtypes.NotBondedPoolName, gomock.Any())
	s.applyValidatorSetUpdates(ctx, keeper, 1)

	validator, found := keeper.GetValidator(ctx, addrVals[0])
	require.True(found)
	require.Equal(ctx.BlockHeight(), validator.UnbondingHeight)
	params := keeper.GetParams(ctx)
	require.True(ctx.BlockHeader().Time.Add(params.UnbondingTime).Equal(validator.UnbondingTime))

	// unbond the validator
	keeper.UnbondingToUnbonded(ctx, validator)

	// redelegate some of the delegation's shares
	redelegationTokens := keeper.TokensFromConsensusPower(ctx, 6)
	s.bankKeeper.EXPECT().SendCoinsFromModuleToModule(gomock.Any(), stakingtypes.NotBondedPoolName, stakingtypes.BondedPoolName, gomock.Any())
	_, err = keeper.BeginRedelegation(ctx, addrDels[1], addrVals[0], addrVals[1], sdk.NewDecFromInt(redelegationTokens))
	require.NoError(err)

	// no red should have been found
	red, found := keeper.GetRedelegation(ctx, addrDels[0], addrVals[0], addrVals[1])
	require.False(found, "%v", red)
}
