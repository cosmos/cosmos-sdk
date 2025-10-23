package keeper_test

import (
	"time"

	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	cmttime "github.com/cometbft/cometbft/types/time"
	"go.uber.org/mock/gomock"

	"cosmossdk.io/math"
	storetypes "cosmossdk.io/store/types"

	"github.com/cosmos/cosmos-sdk/codec/address"
	"github.com/cosmos/cosmos-sdk/runtime"
	sdktestutil "github.com/cosmos/cosmos-sdk/testutil"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	sdk "github.com/cosmos/cosmos-sdk/types"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
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

	s.accountKeeper.EXPECT().AddressCodec().Return(address.NewBech32Codec("cosmos")).AnyTimes()

	// construct the validators
	amts := []math.Int{math.NewInt(9), math.NewInt(8), math.NewInt(7)}
	var validators [3]stakingtypes.Validator
	for i, amt := range amts {
		validators[i] = testutil.NewValidator(s.T(), valAddrs[i], PKs[i])
		validators[i], _ = validators[i].AddTokensFromDel(amt)

		validators[i] = stakingkeeper.TestingUpdateValidator(keeper, ctx, validators[i], true)
	}

	// first add a validators[0] to delegate to
	bond1to1 := stakingtypes.NewDelegation(addrDels[0].String(), valAddrs[0].String(), math.LegacyNewDec(9))

	// check the empty keeper first
	_, err := keeper.GetDelegation(ctx, addrDels[0], valAddrs[0])
	require.ErrorIs(err, stakingtypes.ErrNoDelegation)

	// set and retrieve a record
	require.NoError(keeper.SetDelegation(ctx, bond1to1))
	resBond, err := keeper.GetDelegation(ctx, addrDels[0], valAddrs[0])
	require.NoError(err)
	require.Equal(bond1to1, resBond)

	// modify a records, save, and retrieve
	bond1to1.Shares = math.LegacyNewDec(99)
	require.NoError(keeper.SetDelegation(ctx, bond1to1))
	resBond, err = keeper.GetDelegation(ctx, addrDels[0], valAddrs[0])
	require.NoError(err)
	require.Equal(bond1to1, resBond)

	// add some more records
	bond1to2 := stakingtypes.NewDelegation(addrDels[0].String(), valAddrs[1].String(), math.LegacyNewDec(9))
	bond1to3 := stakingtypes.NewDelegation(addrDels[0].String(), valAddrs[2].String(), math.LegacyNewDec(9))
	bond2to1 := stakingtypes.NewDelegation(addrDels[1].String(), valAddrs[0].String(), math.LegacyNewDec(9))
	bond2to2 := stakingtypes.NewDelegation(addrDels[1].String(), valAddrs[1].String(), math.LegacyNewDec(9))
	bond2to3 := stakingtypes.NewDelegation(addrDels[1].String(), valAddrs[2].String(), math.LegacyNewDec(9))
	require.NoError(keeper.SetDelegation(ctx, bond1to2))
	require.NoError(keeper.SetDelegation(ctx, bond1to3))
	require.NoError(keeper.SetDelegation(ctx, bond2to1))
	require.NoError(keeper.SetDelegation(ctx, bond2to2))
	require.NoError(keeper.SetDelegation(ctx, bond2to3))

	// test all bond retrieve capabilities
	resBonds, err := keeper.GetDelegatorDelegations(ctx, addrDels[0], 5)
	require.NoError(err)
	require.Equal(3, len(resBonds))
	require.Equal(bond1to1, resBonds[0])
	require.Equal(bond1to2, resBonds[1])
	require.Equal(bond1to3, resBonds[2])
	resBonds, err = keeper.GetAllDelegatorDelegations(ctx, addrDels[0])
	require.NoError(err)
	require.Equal(3, len(resBonds))
	resBonds, err = keeper.GetDelegatorDelegations(ctx, addrDels[0], 2)
	require.NoError(err)
	require.Equal(2, len(resBonds))
	resBonds, err = keeper.GetDelegatorDelegations(ctx, addrDels[1], 5)
	require.NoError(err)
	require.Equal(3, len(resBonds))
	require.Equal(bond2to1, resBonds[0])
	require.Equal(bond2to2, resBonds[1])
	require.Equal(bond2to3, resBonds[2])
	allBonds, err := keeper.GetAllDelegations(ctx)
	require.NoError(err)
	require.Equal(6, len(allBonds))
	require.Equal(bond1to1, allBonds[0])
	require.Equal(bond1to2, allBonds[1])
	require.Equal(bond1to3, allBonds[2])
	require.Equal(bond2to1, allBonds[3])
	require.Equal(bond2to2, allBonds[4])
	require.Equal(bond2to3, allBonds[5])

	resVals, err := keeper.GetDelegatorValidators(ctx, addrDels[0], 3)
	require.NoError(err)
	require.Equal(3, len(resVals.Validators))
	resVals, err = keeper.GetDelegatorValidators(ctx, addrDels[1], 4)
	require.NoError(err)
	require.Equal(3, len(resVals.Validators))

	for i := range 3 {
		resVal, err := keeper.GetDelegatorValidator(ctx, addrDels[0], valAddrs[i])
		require.Nil(err)
		require.Equal(valAddrs[i].String(), resVal.GetOperator())

		resVal, err = keeper.GetDelegatorValidator(ctx, addrDels[1], valAddrs[i])
		require.Nil(err)
		require.Equal(valAddrs[i].String(), resVal.GetOperator())

		resDels, err := keeper.GetValidatorDelegations(ctx, valAddrs[i])
		require.NoError(err)
		require.Len(resDels, 2)
	}

	// test total bonded for single delegator
	expBonded := bond1to1.Shares.Add(bond2to1.Shares).Add(bond1to3.Shares)
	resDelBond, err := keeper.GetDelegatorBonded(ctx, addrDels[0])
	require.NoError(err)
	require.Equal(expBonded, math.LegacyNewDecFromInt(resDelBond))

	// delete a record
	require.NoError(keeper.RemoveDelegation(ctx, bond2to3))
	_, err = keeper.GetDelegation(ctx, addrDels[1], valAddrs[2])
	require.ErrorIs(err, stakingtypes.ErrNoDelegation)
	resBonds, err = keeper.GetDelegatorDelegations(ctx, addrDels[1], 5)
	require.NoError(err)
	require.Equal(2, len(resBonds))
	require.Equal(bond2to1, resBonds[0])
	require.Equal(bond2to2, resBonds[1])

	resBonds, err = keeper.GetAllDelegatorDelegations(ctx, addrDels[1])
	require.NoError(err)
	require.Equal(2, len(resBonds))

	// delete all the records from delegator 2
	require.NoError(keeper.RemoveDelegation(ctx, bond2to1))
	require.NoError(keeper.RemoveDelegation(ctx, bond2to2))
	_, err = keeper.GetDelegation(ctx, addrDels[1], valAddrs[0])
	require.ErrorIs(err, stakingtypes.ErrNoDelegation)
	_, err = keeper.GetDelegation(ctx, addrDels[1], valAddrs[1])
	require.ErrorIs(err, stakingtypes.ErrNoDelegation)
	resBonds, err = keeper.GetDelegatorDelegations(ctx, addrDels[1], 5)
	require.NoError(err)
	require.Equal(0, len(resBonds))
}

func (s *KeeperTestSuite) TestDelegationsByValIndex() {
	ctx, keeper := s.ctx, s.stakingKeeper
	require := s.Require()

	addrDels, valAddrs := createValAddrs(3)

	for _, addr := range addrDels {
		s.bankKeeper.EXPECT().DelegateCoinsFromAccountToModule(gomock.Any(), addr, gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
	}
	s.accountKeeper.EXPECT().AddressCodec().Return(address.NewBech32Codec("cosmos")).AnyTimes()

	// construct the validators
	amts := []math.Int{math.NewInt(9), math.NewInt(8), math.NewInt(7)}
	var validators [3]stakingtypes.Validator
	for i, amt := range amts {
		validators[i] = testutil.NewValidator(s.T(), valAddrs[i], PKs[i])
		validators[i], _ = validators[i].AddTokensFromDel(amt)

		validators[i] = stakingkeeper.TestingUpdateValidator(keeper, ctx, validators[i], true)
	}

	// delegate 2 tokens
	//
	// total delegations after delegating: del1 -> 2stake
	_, err := s.msgServer.Delegate(ctx, stakingtypes.NewMsgDelegate(addrDels[0].String(), valAddrs[0].String(), sdk.NewCoin(sdk.DefaultBondDenom, math.NewInt(2))))
	require.NoError(err)

	dels, err := s.stakingKeeper.GetValidatorDelegations(ctx, valAddrs[0])
	require.NoError(err)
	require.Len(dels, 1)

	// delegate 4 tokens
	//
	// total delegations after delegating: del1 -> 2stake, del2 -> 4stake
	_, err = s.msgServer.Delegate(ctx, stakingtypes.NewMsgDelegate(addrDels[1].String(), valAddrs[0].String(), sdk.NewCoin(sdk.DefaultBondDenom, math.NewInt(4))))
	require.NoError(err)

	dels, err = s.stakingKeeper.GetValidatorDelegations(ctx, valAddrs[0])
	require.NoError(err)
	require.Len(dels, 2)

	// undelegate 1 token from del1
	//
	// total delegations after undelegating: del1 -> 1stake, del2 -> 4stake
	_, err = s.msgServer.Undelegate(ctx, stakingtypes.NewMsgUndelegate(addrDels[0].String(), valAddrs[0].String(), sdk.NewCoin(sdk.DefaultBondDenom, math.NewInt(1))))
	require.NoError(err)

	dels, err = s.stakingKeeper.GetValidatorDelegations(ctx, valAddrs[0])
	require.NoError(err)
	require.Len(dels, 2)

	// undelegate 1 token from del1
	//
	// total delegations after undelegating: del2 -> 4stake
	_, err = s.msgServer.Undelegate(ctx, stakingtypes.NewMsgUndelegate(addrDels[0].String(), valAddrs[0].String(), sdk.NewCoin(sdk.DefaultBondDenom, math.NewInt(1))))
	require.NoError(err)

	dels, err = s.stakingKeeper.GetValidatorDelegations(ctx, valAddrs[0])
	require.NoError(err)
	require.Len(dels, 1)

	// undelegate 2 tokens from del2
	//
	// total delegations after undelegating: del2 -> 2stake
	_, err = s.msgServer.Undelegate(ctx, stakingtypes.NewMsgUndelegate(addrDels[1].String(), valAddrs[0].String(), sdk.NewCoin(sdk.DefaultBondDenom, math.NewInt(2))))
	require.NoError(err)

	dels, err = s.stakingKeeper.GetValidatorDelegations(ctx, valAddrs[0])
	require.NoError(err)
	require.Len(dels, 1)

	// undelegate 2 tokens from del2
	//
	// total delegations after undelegating: []
	_, err = s.msgServer.Undelegate(ctx, stakingtypes.NewMsgUndelegate(addrDels[1].String(), valAddrs[0].String(), sdk.NewCoin(sdk.DefaultBondDenom, math.NewInt(2))))
	require.NoError(err)

	dels, err = s.stakingKeeper.GetValidatorDelegations(ctx, valAddrs[0])
	require.NoError(err)
	require.Len(dels, 0)
}

// tests Get/Set/Remove UnbondingDelegation
func (s *KeeperTestSuite) TestUnbondingDelegation() {
	ctx, keeper := s.ctx, s.stakingKeeper
	require := s.Require()

	delAddrs, valAddrs := createValAddrs(2)

	s.accountKeeper.EXPECT().AddressCodec().Return(address.NewBech32Codec("cosmos")).AnyTimes()

	ubd := stakingtypes.NewUnbondingDelegation(
		delAddrs[0],
		valAddrs[0],
		0,
		time.Unix(0, 0).UTC(),
		math.NewInt(5),
		0,
		address.NewBech32Codec("cosmosvaloper"), address.NewBech32Codec("cosmos"),
	)

	// set and retrieve a record
	require.NoError(keeper.SetUnbondingDelegation(ctx, ubd))
	resUnbond, err := keeper.GetUnbondingDelegation(ctx, delAddrs[0], valAddrs[0])
	require.NoError(err)
	require.Equal(ubd, resUnbond)

	// modify a records, save, and retrieve
	expUnbond := math.NewInt(21)
	ubd.Entries[0].Balance = expUnbond
	require.NoError(keeper.SetUnbondingDelegation(ctx, ubd))

	resUnbonds, err := keeper.GetUnbondingDelegations(ctx, delAddrs[0], 5)
	require.NoError(err)
	require.Equal(1, len(resUnbonds))

	resUnbonds, err = keeper.GetAllUnbondingDelegations(ctx, delAddrs[0])
	require.NoError(err)
	require.Equal(1, len(resUnbonds))

	resUnbond, err = keeper.GetUnbondingDelegation(ctx, delAddrs[0], valAddrs[0])
	require.NoError(err)
	require.Equal(ubd, resUnbond)

	resDelUnbond, err := keeper.GetDelegatorUnbonding(ctx, delAddrs[0])
	require.NoError(err)
	require.Equal(expUnbond, resDelUnbond)

	// delete a record
	require.NoError(keeper.RemoveUnbondingDelegation(ctx, ubd))
	_, err = keeper.GetUnbondingDelegation(ctx, delAddrs[0], valAddrs[0])
	require.ErrorIs(err, stakingtypes.ErrNoUnbondingDelegation)

	resUnbonds, err = keeper.GetUnbondingDelegations(ctx, delAddrs[0], 5)
	require.NoError(err)
	require.Equal(0, len(resUnbonds))

	resUnbonds, err = keeper.GetAllUnbondingDelegations(ctx, delAddrs[0])
	require.NoError(err)
	require.Equal(0, len(resUnbonds))
}

func (s *KeeperTestSuite) TestUnbondingDelegationsFromValidator() {
	ctx, keeper := s.ctx, s.stakingKeeper
	require := s.Require()

	delAddrs, valAddrs := createValAddrs(2)

	ubd := stakingtypes.NewUnbondingDelegation(
		delAddrs[0],
		valAddrs[0],
		0,
		time.Unix(0, 0).UTC(),
		math.NewInt(5),
		0,
		address.NewBech32Codec("cosmosvaloper"), address.NewBech32Codec("cosmos"),
	)

	// set and retrieve a record
	require.NoError(keeper.SetUnbondingDelegation(ctx, ubd))
	resUnbond, err := keeper.GetUnbondingDelegation(ctx, delAddrs[0], valAddrs[0])
	require.NoError(err)
	require.Equal(ubd, resUnbond)

	// modify a records, save, and retrieve
	expUnbond := math.NewInt(21)
	ubd.Entries[0].Balance = expUnbond
	require.NoError(keeper.SetUnbondingDelegation(ctx, ubd))

	resUnbonds, err := keeper.GetUnbondingDelegations(ctx, delAddrs[0], 5)
	require.NoError(err)
	require.Equal(1, len(resUnbonds))

	resUnbonds, err = keeper.GetAllUnbondingDelegations(ctx, delAddrs[0])
	require.NoError(err)
	require.Equal(1, len(resUnbonds))

	resUnbonds, err = keeper.GetUnbondingDelegationsFromValidator(ctx, valAddrs[0])
	require.NoError(err)
	require.Equal(1, len(resUnbonds))

	resUnbond, err = keeper.GetUnbondingDelegation(ctx, delAddrs[0], valAddrs[0])
	require.NoError(err)
	require.Equal(ubd, resUnbond)

	resDelUnbond, err := keeper.GetDelegatorUnbonding(ctx, delAddrs[0])
	require.NoError(err)
	require.Equal(expUnbond, resDelUnbond)

	// delete a record
	require.NoError(keeper.RemoveUnbondingDelegation(ctx, ubd))
	_, err = keeper.GetUnbondingDelegation(ctx, delAddrs[0], valAddrs[0])
	require.ErrorIs(err, stakingtypes.ErrNoUnbondingDelegation)

	resUnbonds, err = keeper.GetUnbondingDelegations(ctx, delAddrs[0], 5)
	require.NoError(err)
	require.Equal(0, len(resUnbonds))

	resUnbonds, err = keeper.GetAllUnbondingDelegations(ctx, delAddrs[0])
	require.NoError(err)
	require.Equal(0, len(resUnbonds))

	resUnbonds, err = keeper.GetUnbondingDelegationsFromValidator(ctx, valAddrs[0])
	require.NoError(err)
	require.Equal(0, len(resUnbonds))
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

	delegation := stakingtypes.NewDelegation(delAddrs[0].String(), valAddrs[0].String(), issuedShares)
	require.NoError(keeper.SetDelegation(ctx, delegation))

	bondTokens := keeper.TokensFromConsensusPower(ctx, 6)
	amount, err := keeper.Unbond(ctx, delAddrs[0], valAddrs[0], math.LegacyNewDecFromInt(bondTokens))
	require.NoError(err)
	require.Equal(bondTokens, amount) // shares to be added to an unbonding delegation

	delegation, err = keeper.GetDelegation(ctx, delAddrs[0], valAddrs[0])
	require.NoError(err)
	validator, err = keeper.GetValidator(ctx, valAddrs[0])
	require.NoError(err)

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
	require.NoError(keeper.SetValidatorByConsAddr(ctx, validator))
	require.True(validator.IsBonded())

	selfDelegation := stakingtypes.NewDelegation(sdk.AccAddress(addrVals[0].Bytes()).String(), addrVals[0].String(), issuedShares)
	require.NoError(keeper.SetDelegation(ctx, selfDelegation))

	// create a second delegation to this validator
	require.NoError(keeper.DeleteValidatorByPowerIndex(ctx, validator))
	validator, issuedShares = validator.AddTokensFromDel(delTokens)
	require.True(validator.IsBonded())
	require.Equal(delTokens, issuedShares.RoundInt())

	validator = stakingkeeper.TestingUpdateValidator(keeper, ctx, validator, true)
	delegation := stakingtypes.NewDelegation(addrDels[0].String(), addrVals[0].String(), issuedShares)
	require.NoError(keeper.SetDelegation(ctx, delegation))

	val0AccAddr := sdk.AccAddress(addrVals[0].Bytes())
	s.bankKeeper.EXPECT().SendCoinsFromModuleToModule(gomock.Any(), stakingtypes.BondedPoolName, stakingtypes.NotBondedPoolName, gomock.Any())
	_, _, err := keeper.Undelegate(ctx, val0AccAddr, addrVals[0], math.LegacyNewDecFromInt(keeper.TokensFromConsensusPower(ctx, 6)))
	require.NoError(err)

	// end block
	s.bankKeeper.EXPECT().SendCoinsFromModuleToModule(gomock.Any(), stakingtypes.BondedPoolName, stakingtypes.NotBondedPoolName, gomock.Any())
	s.applyValidatorSetUpdates(ctx, keeper, 1)

	validator, err = keeper.GetValidator(ctx, addrVals[0])
	require.NoError(err)
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
	require.NoError(keeper.SetValidatorByConsAddr(ctx, validator))

	validator, issuedShares := validator.AddTokensFromDel(delTokens)
	require.Equal(delTokens, issuedShares.RoundInt())

	s.bankKeeper.EXPECT().SendCoinsFromModuleToModule(gomock.Any(), stakingtypes.NotBondedPoolName, stakingtypes.BondedPoolName, gomock.Any())
	validator = stakingkeeper.TestingUpdateValidator(keeper, ctx, validator, true)
	require.True(validator.IsBonded())

	selfDelegation := stakingtypes.NewDelegation(addrDels[0].String(), addrVals[0].String(), issuedShares)
	require.NoError(keeper.SetDelegation(ctx, selfDelegation))

	// create a second delegation to this validator
	require.NoError(keeper.DeleteValidatorByPowerIndex(ctx, validator))

	validator, issuedShares = validator.AddTokensFromDel(delTokens)
	require.Equal(delTokens, issuedShares.RoundInt())

	stakingkeeper.TestingUpdateValidator(keeper, ctx, validator, true)
	delegation := stakingtypes.NewDelegation(addrDels[1].String(), addrVals[0].String(), issuedShares)
	require.NoError(keeper.SetDelegation(ctx, delegation))

	header := ctx.BlockHeader()
	blockHeight := int64(10)
	header.Height = blockHeight
	blockTime := time.Unix(333, 0)
	header.Time = blockTime
	ctx = ctx.WithBlockHeader(header)

	// unbond the all self-delegation to put validator in unbonding state
	val0AccAddr := sdk.AccAddress(addrVals[0])
	s.bankKeeper.EXPECT().SendCoinsFromModuleToModule(gomock.Any(), stakingtypes.BondedPoolName, stakingtypes.NotBondedPoolName, gomock.Any())
	_, amount, err := keeper.Undelegate(ctx, val0AccAddr, addrVals[0], math.LegacyNewDecFromInt(delTokens))
	require.NoError(err)
	require.Equal(amount, delTokens)

	// end block
	s.bankKeeper.EXPECT().SendCoinsFromModuleToModule(gomock.Any(), stakingtypes.BondedPoolName, stakingtypes.NotBondedPoolName, gomock.Any())
	s.applyValidatorSetUpdates(ctx, keeper, 1)

	validator, err = keeper.GetValidator(ctx, addrVals[0])
	require.NoError(err)
	require.Equal(blockHeight, validator.UnbondingHeight)
	params, err := keeper.GetParams(ctx)
	require.NoError(err)
	require.True(blockTime.Add(params.UnbondingTime).Equal(validator.UnbondingTime))

	blockHeight2 := int64(20)
	blockTime2 := time.Unix(444, 0).UTC()
	ctx = ctx.WithBlockHeight(blockHeight2)
	ctx = ctx.WithBlockTime(blockTime2)

	// unbond some of the other delegation's shares
	undelegateAmount := math.LegacyNewDec(6)
	_, undelegatedAmount, err := keeper.Undelegate(ctx, addrDels[1], addrVals[0], undelegateAmount)
	require.NoError(err)
	require.Equal(math.LegacyNewDecFromInt(undelegatedAmount), undelegateAmount)

	// retrieve the unbonding delegation
	ubd, err := keeper.GetUnbondingDelegation(ctx, addrDels[1], addrVals[0])
	require.NoError(err)
	require.Len(ubd.Entries, 1)
	require.True(ubd.Entries[0].Balance.Equal(math.NewInt(6)))
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
	require.NoError(keeper.SetValidatorByConsAddr(ctx, validator))

	valTokens := keeper.TokensFromConsensusPower(ctx, 10)
	validator, issuedShares := validator.AddTokensFromDel(valTokens)
	require.Equal(valTokens, issuedShares.RoundInt())
	s.bankKeeper.EXPECT().SendCoinsFromModuleToModule(gomock.Any(), stakingtypes.NotBondedPoolName, stakingtypes.BondedPoolName, gomock.Any())
	validator = stakingkeeper.TestingUpdateValidator(keeper, ctx, validator, true)
	require.True(validator.IsBonded())

	val0AccAddr := sdk.AccAddress(addrVals[0])
	selfDelegation := stakingtypes.NewDelegation(val0AccAddr.String(), addrVals[0].String(), issuedShares)
	require.NoError(keeper.SetDelegation(ctx, selfDelegation))

	// create a second delegation to this validator
	require.NoError(keeper.DeleteValidatorByPowerIndex(ctx, validator))
	validator, issuedShares = validator.AddTokensFromDel(delTokens)
	require.Equal(delTokens, issuedShares.RoundInt())
	validator = stakingkeeper.TestingUpdateValidator(keeper, ctx, validator, true)
	require.True(validator.IsBonded())
	delegation := stakingtypes.NewDelegation(addrDels[1].String(), addrVals[0].String(), issuedShares)
	require.NoError(keeper.SetDelegation(ctx, delegation))

	ctx = ctx.WithBlockHeight(10)
	ctx = ctx.WithBlockTime(time.Unix(333, 0))

	// unbond the all self-delegation to put validator in unbonding state
	s.bankKeeper.EXPECT().SendCoinsFromModuleToModule(gomock.Any(), stakingtypes.BondedPoolName, stakingtypes.NotBondedPoolName, gomock.Any())
	_, amount, err := keeper.Undelegate(ctx, val0AccAddr, addrVals[0], math.LegacyNewDecFromInt(valTokens))
	require.NoError(err)
	require.Equal(amount, valTokens)

	// end block
	s.bankKeeper.EXPECT().SendCoinsFromModuleToModule(gomock.Any(), stakingtypes.BondedPoolName, stakingtypes.NotBondedPoolName, gomock.Any())
	s.applyValidatorSetUpdates(ctx, keeper, 1)

	validator, err = keeper.GetValidator(ctx, addrVals[0])
	require.NoError(err)
	require.Equal(ctx.BlockHeight(), validator.UnbondingHeight)
	params, err := keeper.GetParams(ctx)
	require.NoError(err)
	require.True(ctx.BlockHeader().Time.Add(params.UnbondingTime).Equal(validator.UnbondingTime))

	// unbond the validator
	ctx = ctx.WithBlockTime(validator.UnbondingTime)
	err = keeper.UnbondAllMatureValidators(ctx)
	require.NoError(err)

	// Make sure validator is still in state because there is still an outstanding delegation
	validator, err = keeper.GetValidator(ctx, addrVals[0])
	require.NoError(err)
	require.Equal(validator.Status, stakingtypes.Unbonded)

	// unbond some of the other delegation's shares
	unbondTokens := keeper.TokensFromConsensusPower(ctx, 6)
	_, amount2, err := keeper.Undelegate(ctx, addrDels[1], addrVals[0], math.LegacyNewDecFromInt(unbondTokens))
	require.NoError(err)
	require.Equal(amount2, unbondTokens)

	// unbond rest of the other delegation's shares
	remainingTokens := delTokens.Sub(unbondTokens)
	_, amount3, err := keeper.Undelegate(ctx, addrDels[1], addrVals[0], math.LegacyNewDecFromInt(remainingTokens))
	require.NoError(err)
	require.Equal(amount3, remainingTokens)

	//  now validator should be deleted from state
	validator, err = keeper.GetValidator(ctx, addrVals[0])
	require.ErrorIs(err, stakingtypes.ErrNoValidatorFound)
}

func (s *KeeperTestSuite) TestUnbondingAllDelegationFromValidator() {
	ctx, keeper := s.ctx, s.stakingKeeper
	require := s.Require()

	delTokens := keeper.TokensFromConsensusPower(ctx, 10)
	addrDels, addrVals := createValAddrs(2)

	// create a validator with a self-delegation
	validator := testutil.NewValidator(s.T(), addrVals[0], PKs[0])
	require.NoError(keeper.SetValidatorByConsAddr(ctx, validator))

	valTokens := keeper.TokensFromConsensusPower(ctx, 10)
	validator, issuedShares := validator.AddTokensFromDel(valTokens)
	require.Equal(valTokens, issuedShares.RoundInt())

	s.bankKeeper.EXPECT().SendCoinsFromModuleToModule(gomock.Any(), stakingtypes.NotBondedPoolName, stakingtypes.BondedPoolName, gomock.Any())
	validator = stakingkeeper.TestingUpdateValidator(keeper, ctx, validator, true)
	require.True(validator.IsBonded())
	val0AccAddr := sdk.AccAddress(addrVals[0].Bytes())

	selfDelegation := stakingtypes.NewDelegation(val0AccAddr.String(), addrVals[0].String(), issuedShares)
	require.NoError(keeper.SetDelegation(ctx, selfDelegation))

	// create a second delegation to this validator
	require.NoError(keeper.DeleteValidatorByPowerIndex(ctx, validator))
	validator, issuedShares = validator.AddTokensFromDel(delTokens)
	require.Equal(delTokens, issuedShares.RoundInt())

	validator = stakingkeeper.TestingUpdateValidator(keeper, ctx, validator, true)
	require.True(validator.IsBonded())

	delegation := stakingtypes.NewDelegation(addrDels[1].String(), addrVals[0].String(), issuedShares)
	require.NoError(keeper.SetDelegation(ctx, delegation))

	ctx = ctx.WithBlockHeight(10)
	ctx = ctx.WithBlockTime(time.Unix(333, 0))

	// unbond the all self-delegation to put validator in unbonding state
	s.bankKeeper.EXPECT().SendCoinsFromModuleToModule(gomock.Any(), stakingtypes.BondedPoolName, stakingtypes.NotBondedPoolName, gomock.Any())
	_, amount, err := keeper.Undelegate(ctx, val0AccAddr, addrVals[0], math.LegacyNewDecFromInt(valTokens))
	require.NoError(err)
	require.Equal(amount, valTokens)

	// end block
	s.bankKeeper.EXPECT().SendCoinsFromModuleToModule(gomock.Any(), stakingtypes.BondedPoolName, stakingtypes.NotBondedPoolName, gomock.Any())
	s.applyValidatorSetUpdates(ctx, keeper, 1)

	// unbond all the remaining delegation
	_, amount2, err := keeper.Undelegate(ctx, addrDels[1], addrVals[0], math.LegacyNewDecFromInt(delTokens))
	require.NoError(err)
	require.Equal(amount2, delTokens)

	// validator should still be in state and still be in unbonding state
	validator, err = keeper.GetValidator(ctx, addrVals[0])
	require.NoError(err)
	require.Equal(validator.Status, stakingtypes.Unbonding)

	// unbond the validator
	ctx = ctx.WithBlockTime(validator.UnbondingTime)
	err = keeper.UnbondAllMatureValidators(ctx)
	require.NoError(err)

	// validator should now be deleted from state
	_, err = keeper.GetValidator(ctx, addrVals[0])
	require.ErrorIs(err, stakingtypes.ErrNoValidatorFound)
}

// Make sure that that the retrieving the delegations doesn't affect the state
func (s *KeeperTestSuite) TestGetRedelegationsFromSrcValidator() {
	ctx, keeper := s.ctx, s.stakingKeeper
	require := s.Require()

	addrDels, addrVals := createValAddrs(2)

	rd := stakingtypes.NewRedelegation(addrDels[0], addrVals[0], addrVals[1], 0,
		time.Unix(0, 0), math.NewInt(5),
		math.LegacyNewDec(5), 0, address.NewBech32Codec("cosmosvaloper"), address.NewBech32Codec("cosmos"))

	// set and retrieve a record
	err := keeper.SetRedelegation(ctx, rd)
	require.NoError(err)
	resBond, err := keeper.GetRedelegation(ctx, addrDels[0], addrVals[0], addrVals[1])
	require.NoError(err)

	// get the redelegations one time
	redelegations, err := keeper.GetRedelegationsFromSrcValidator(ctx, addrVals[0])
	require.NoError(err)
	require.Equal(1, len(redelegations))
	require.Equal(redelegations[0], resBond)

	// get the redelegations a second time, should be exactly the same
	redelegations, err = keeper.GetRedelegationsFromSrcValidator(ctx, addrVals[0])
	require.NoError(err)
	require.Equal(1, len(redelegations))
	require.Equal(redelegations[0], resBond)
}

// tests Get/Set/Remove/Has UnbondingDelegation
func (s *KeeperTestSuite) TestRedelegation() {
	ctx, keeper := s.ctx, s.stakingKeeper
	require := s.Require()

	addrDels, addrVals := createValAddrs(2)

	rd := stakingtypes.NewRedelegation(addrDels[0], addrVals[0], addrVals[1], 0,
		time.Unix(0, 0).UTC(), math.NewInt(5),
		math.LegacyNewDec(5), 0, address.NewBech32Codec("cosmosvaloper"), address.NewBech32Codec("cosmos"))

	// test shouldn't have and redelegations
	has, err := keeper.HasReceivingRedelegation(ctx, addrDels[0], addrVals[1])
	require.NoError(err)
	require.False(has)

	// set and retrieve a record
	err = keeper.SetRedelegation(ctx, rd)
	require.NoError(err)
	resRed, err := keeper.GetRedelegation(ctx, addrDels[0], addrVals[0], addrVals[1])
	require.NoError(err)

	redelegations, err := keeper.GetRedelegationsFromSrcValidator(ctx, addrVals[0])
	require.NoError(err)
	require.Equal(1, len(redelegations))
	require.Equal(redelegations[0], resRed)

	redelegations, err = keeper.GetRedelegations(ctx, addrDels[0], 5)
	require.NoError(err)
	require.Equal(1, len(redelegations))
	require.Equal(redelegations[0], resRed)

	redelegations, err = keeper.GetAllRedelegations(ctx, addrDels[0], nil, nil)
	require.NoError(err)
	require.Equal(1, len(redelegations))
	require.Equal(redelegations[0], resRed)

	// check if has the redelegation
	has, err = keeper.HasReceivingRedelegation(ctx, addrDels[0], addrVals[1])
	require.NoError(err)
	require.True(has)

	// modify a records, save, and retrieve
	rd.Entries[0].SharesDst = math.LegacyNewDec(21)
	err = keeper.SetRedelegation(ctx, rd)
	require.NoError(err)

	resRed, err = keeper.GetRedelegation(ctx, addrDels[0], addrVals[0], addrVals[1])
	require.NoError(err)
	require.Equal(rd, resRed)

	redelegations, err = keeper.GetRedelegationsFromSrcValidator(ctx, addrVals[0])
	require.NoError(err)
	require.Equal(1, len(redelegations))
	require.Equal(redelegations[0], resRed)

	redelegations, err = keeper.GetRedelegations(ctx, addrDels[0], 5)
	require.NoError(err)
	require.Equal(1, len(redelegations))
	require.Equal(redelegations[0], resRed)

	// delete a record
	err = keeper.RemoveRedelegation(ctx, rd)
	require.NoError(err)
	_, err = keeper.GetRedelegation(ctx, addrDels[0], addrVals[0], addrVals[1])
	require.ErrorIs(err, stakingtypes.ErrNoRedelegation)

	redelegations, err = keeper.GetRedelegations(ctx, addrDels[0], 5)
	require.NoError(err)
	require.Equal(0, len(redelegations))

	redelegations, err = keeper.GetAllRedelegations(ctx, addrDels[0], nil, nil)
	require.NoError(err)
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

	selfDelegation := stakingtypes.NewDelegation(val0AccAddr.String(), addrVals[0].String(), issuedShares)
	require.NoError(keeper.SetDelegation(ctx, selfDelegation))

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
	selfDelegation := stakingtypes.NewDelegation(val0AccAddr.String(), addrVals[0].String(), issuedShares)
	require.NoError(keeper.SetDelegation(ctx, selfDelegation))

	// create a second validator
	validator2 := testutil.NewValidator(s.T(), addrVals[1], PKs[1])
	validator2, issuedShares = validator2.AddTokensFromDel(valTokens)
	require.Equal(valTokens, issuedShares.RoundInt())

	s.bankKeeper.EXPECT().SendCoinsFromModuleToModule(gomock.Any(), stakingtypes.NotBondedPoolName, stakingtypes.BondedPoolName, gomock.Any())
	validator2 = stakingkeeper.TestingUpdateValidator(keeper, ctx, validator2, true)
	require.Equal(stakingtypes.Bonded, validator2.Status)

	maxEntries, err := keeper.MaxEntries(ctx)
	require.NoError(err)

	// redelegations should pass
	var completionTime time.Time
	for i := uint32(0); i < maxEntries; i++ {
		var err error
		completionTime, err = keeper.BeginRedelegation(ctx, val0AccAddr, addrVals[0], addrVals[1], math.LegacyNewDec(1))
		require.NoError(err)
	}

	// an additional redelegation should fail due to max entries
	_, err = keeper.BeginRedelegation(ctx, val0AccAddr, addrVals[0], addrVals[1], math.LegacyNewDec(1))
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
	require.NoError(keeper.SetValidatorByConsAddr(ctx, validator))

	valTokens := keeper.TokensFromConsensusPower(ctx, 10)
	validator, issuedShares := validator.AddTokensFromDel(valTokens)
	require.Equal(valTokens, issuedShares.RoundInt())

	s.bankKeeper.EXPECT().SendCoinsFromModuleToModule(gomock.Any(), stakingtypes.NotBondedPoolName, stakingtypes.BondedPoolName, gomock.Any())
	validator = stakingkeeper.TestingUpdateValidator(keeper, ctx, validator, true)

	val0AccAddr := sdk.AccAddress(addrVals[0])
	selfDelegation := stakingtypes.NewDelegation(val0AccAddr.String(), addrVals[0].String(), issuedShares)
	require.NoError(keeper.SetDelegation(ctx, selfDelegation))

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
	stakingkeeper.TestingUpdateValidator(keeper, ctx, validator, true)

	delegation := stakingtypes.NewDelegation(addrDels[0].String(), addrVals[0].String(), issuedShares)
	require.NoError(keeper.SetDelegation(ctx, delegation))

	_, err := keeper.BeginRedelegation(ctx, val0AccAddr, addrVals[0], addrVals[1], math.LegacyNewDecFromInt(delTokens))
	require.NoError(err)

	// end block
	s.bankKeeper.EXPECT().SendCoinsFromModuleToModule(gomock.Any(), stakingtypes.BondedPoolName, stakingtypes.NotBondedPoolName, gomock.Any())
	s.applyValidatorSetUpdates(ctx, keeper, 2)

	validator, err = keeper.GetValidator(ctx, addrVals[0])
	require.NoError(err)
	require.Equal(valTokens, validator.Tokens)
	require.Equal(stakingtypes.Unbonding, validator.Status)
}

func (s *KeeperTestSuite) TestRedelegateFromUnbondingValidator() {
	ctx, keeper := s.ctx, s.stakingKeeper
	require := s.Require()

	addrDels, addrVals := createValAddrs(2)

	// create a validator with a self-delegation
	validator := testutil.NewValidator(s.T(), addrVals[0], PKs[0])
	require.NoError(keeper.SetValidatorByConsAddr(ctx, validator))

	valTokens := keeper.TokensFromConsensusPower(ctx, 10)
	validator, issuedShares := validator.AddTokensFromDel(valTokens)
	require.Equal(valTokens, issuedShares.RoundInt())
	s.bankKeeper.EXPECT().SendCoinsFromModuleToModule(gomock.Any(), stakingtypes.NotBondedPoolName, stakingtypes.BondedPoolName, gomock.Any())
	validator = stakingkeeper.TestingUpdateValidator(keeper, ctx, validator, true)
	val0AccAddr := sdk.AccAddress(addrVals[0].Bytes())
	selfDelegation := stakingtypes.NewDelegation(val0AccAddr.String(), addrVals[0].String(), issuedShares)
	require.NoError(keeper.SetDelegation(ctx, selfDelegation))

	// create a second delegation to this validator
	require.NoError(keeper.DeleteValidatorByPowerIndex(ctx, validator))
	delTokens := keeper.TokensFromConsensusPower(ctx, 10)
	validator, issuedShares = validator.AddTokensFromDel(delTokens)
	require.Equal(delTokens, issuedShares.RoundInt())
	stakingkeeper.TestingUpdateValidator(keeper, ctx, validator, true)
	delegation := stakingtypes.NewDelegation(addrDels[1].String(), addrVals[0].String(), issuedShares)
	require.NoError(keeper.SetDelegation(ctx, delegation))

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
	_, amount, err := keeper.Undelegate(ctx, val0AccAddr, addrVals[0], math.LegacyNewDecFromInt(delTokens))
	require.NoError(err)
	require.Equal(amount, delTokens)

	// end block
	s.bankKeeper.EXPECT().SendCoinsFromModuleToModule(gomock.Any(), stakingtypes.BondedPoolName, stakingtypes.NotBondedPoolName, gomock.Any())
	s.applyValidatorSetUpdates(ctx, keeper, 1)

	validator, err = keeper.GetValidator(ctx, addrVals[0])
	require.NoError(err)
	require.Equal(blockHeight, validator.UnbondingHeight)
	params, err := keeper.GetParams(ctx)
	require.NoError(err)
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
	_, err = keeper.BeginRedelegation(ctx, addrDels[1], addrVals[0], addrVals[1], math.LegacyNewDecFromInt(redelegateTokens))
	require.NoError(err)

	// retrieve the unbonding delegation
	ubd, err := keeper.GetRedelegation(ctx, addrDels[1], addrVals[0], addrVals[1])
	require.NoError(err)
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
	require.NoError(keeper.SetValidatorByConsAddr(ctx, validator))

	valTokens := keeper.TokensFromConsensusPower(ctx, 10)
	validator, issuedShares := validator.AddTokensFromDel(valTokens)
	require.Equal(valTokens, issuedShares.RoundInt())
	s.bankKeeper.EXPECT().SendCoinsFromModuleToModule(gomock.Any(), stakingtypes.NotBondedPoolName, stakingtypes.BondedPoolName, gomock.Any())
	validator = stakingkeeper.TestingUpdateValidator(keeper, ctx, validator, true)
	val0AccAddr := sdk.AccAddress(addrVals[0].Bytes())
	selfDelegation := stakingtypes.NewDelegation(val0AccAddr.String(), addrVals[0].String(), issuedShares)
	require.NoError(keeper.SetDelegation(ctx, selfDelegation))

	// create a second delegation to this validator
	require.NoError(keeper.DeleteValidatorByPowerIndex(ctx, validator))
	delTokens := keeper.TokensFromConsensusPower(ctx, 10)
	validator, issuedShares = validator.AddTokensFromDel(delTokens)
	require.Equal(delTokens, issuedShares.RoundInt())
	stakingkeeper.TestingUpdateValidator(keeper, ctx, validator, true)
	delegation := stakingtypes.NewDelegation(addrDels[1].String(), addrVals[0].String(), issuedShares)
	require.NoError(keeper.SetDelegation(ctx, delegation))

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
	_, amount, err := keeper.Undelegate(ctx, val0AccAddr, addrVals[0], math.LegacyNewDecFromInt(delTokens))
	require.NoError(err)
	require.Equal(amount, delTokens)

	// end block
	s.bankKeeper.EXPECT().SendCoinsFromModuleToModule(gomock.Any(), stakingtypes.BondedPoolName, stakingtypes.NotBondedPoolName, gomock.Any())
	s.applyValidatorSetUpdates(ctx, keeper, 1)

	validator, err = keeper.GetValidator(ctx, addrVals[0])
	require.NoError(err)
	require.Equal(ctx.BlockHeight(), validator.UnbondingHeight)
	params, err := keeper.GetParams(ctx)
	require.NoError(err)
	require.True(ctx.BlockHeader().Time.Add(params.UnbondingTime).Equal(validator.UnbondingTime))

	// unbond the validator
	_, err = keeper.UnbondingToUnbonded(ctx, validator)
	require.NoError(err)

	// redelegate some of the delegation's shares
	redelegationTokens := keeper.TokensFromConsensusPower(ctx, 6)
	s.bankKeeper.EXPECT().SendCoinsFromModuleToModule(gomock.Any(), stakingtypes.NotBondedPoolName, stakingtypes.BondedPoolName, gomock.Any())
	_, err = keeper.BeginRedelegation(ctx, addrDels[1], addrVals[0], addrVals[1], math.LegacyNewDecFromInt(redelegationTokens))
	require.NoError(err)

	// no red should have been found
	red, err := keeper.GetRedelegation(ctx, addrDels[0], addrVals[0], addrVals[1])
	require.ErrorIs(err, stakingtypes.ErrNoRedelegation, "%v", red)
}

func (s *KeeperTestSuite) TestUnbondingDelegationAddEntry() {
	require := s.Require()

	delAddrs, valAddrs := createValAddrs(1)

	delAddr := delAddrs[0]
	valAddr := valAddrs[0]
	creationHeight := int64(10)
	ubd := stakingtypes.NewUnbondingDelegation(
		delAddr,
		valAddr,
		creationHeight,
		time.Unix(0, 0).UTC(),
		math.NewInt(10),
		0,
		address.NewBech32Codec("cosmosvaloper"), address.NewBech32Codec("cosmos"),
	)
	var initialEntries []stakingtypes.UnbondingDelegationEntry
	initialEntries = append(initialEntries, ubd.Entries...)
	require.Len(initialEntries, 1)

	isNew := ubd.AddEntry(creationHeight, time.Unix(0, 0).UTC(), math.NewInt(5), 1)
	require.False(isNew)
	require.Len(ubd.Entries, 1) // entry was merged
	require.NotEqual(initialEntries, ubd.Entries)
	require.Equal(creationHeight, ubd.Entries[0].CreationHeight)
	require.Equal(initialEntries[0].UnbondingId, ubd.Entries[0].UnbondingId) // unbondingID remains unchanged
	require.Equal(ubd.Entries[0].Balance, math.NewInt(15))                   // 10 from previous + 5 from merged

	newCreationHeight := int64(11)
	isNew = ubd.AddEntry(newCreationHeight, time.Unix(1, 0).UTC(), math.NewInt(5), 2)
	require.True(isNew)
	require.Len(ubd.Entries, 2) // entry was appended
	require.NotEqual(initialEntries, ubd.Entries)
	require.Equal(creationHeight, ubd.Entries[0].CreationHeight)
	require.Equal(newCreationHeight, ubd.Entries[1].CreationHeight)
	require.Equal(ubd.Entries[0].Balance, math.NewInt(15))
	require.Equal(ubd.Entries[1].Balance, math.NewInt(5))
	require.NotEqual(ubd.Entries[0].UnbondingId, ubd.Entries[1].UnbondingId) // appended entry has a new unbondingID
}

func (s *KeeperTestSuite) TestSetUnbondingDelegationEntry() {
	ctx, keeper := s.ctx, s.stakingKeeper
	require := s.Require()

	delAddrs, valAddrs := createValAddrs(1)

	delAddr := delAddrs[0]
	valAddr := valAddrs[0]
	creationHeight := int64(0)
	ubd := stakingtypes.NewUnbondingDelegation(
		delAddr,
		valAddr,
		creationHeight,
		time.Unix(0, 0).UTC(),
		math.NewInt(5),
		0,
		address.NewBech32Codec("cosmosvaloper"), address.NewBech32Codec("cosmos"),
	)

	// set and retrieve a record
	require.NoError(keeper.SetUnbondingDelegation(ctx, ubd))
	resUnbond, err := keeper.GetUnbondingDelegation(ctx, delAddr, valAddr)
	require.NoError(err)
	require.Equal(ubd, resUnbond)

	initialEntries := ubd.Entries
	require.Len(initialEntries, 1)
	require.Equal(initialEntries[0].Balance, math.NewInt(5))
	require.Equal(initialEntries[0].UnbondingId, uint64(0)) // initial unbondingID

	// set unbonding delegation entry for existing creationHeight
	// entries are expected to be merged
	_, err = keeper.SetUnbondingDelegationEntry(
		ctx,
		delAddr,
		valAddr,
		creationHeight,
		time.Unix(0, 0).UTC(),
		math.NewInt(5),
	)
	require.NoError(err)
	resUnbonding, err := keeper.GetUnbondingDelegation(ctx, delAddr, valAddr)
	require.NoError(err)
	require.Len(resUnbonding.Entries, 1)
	require.NotEqual(initialEntries, resUnbonding.Entries)
	require.Equal(creationHeight, resUnbonding.Entries[0].CreationHeight)
	require.Equal(initialEntries[0].UnbondingId, resUnbonding.Entries[0].UnbondingId) // initial unbondingID remains unchanged
	require.Equal(resUnbonding.Entries[0].Balance, math.NewInt(10))                   // 5 from previous entry + 5 from merged entry

	// set unbonding delegation entry for newCreationHeight
	// new entry is expected to be appended to the existing entries
	newCreationHeight := int64(1)
	_, err = keeper.SetUnbondingDelegationEntry(
		ctx,
		delAddr,
		valAddr,
		newCreationHeight,
		time.Unix(1, 0).UTC(),
		math.NewInt(10),
	)
	require.NoError(err)
	resUnbonding, err = keeper.GetUnbondingDelegation(ctx, delAddr, valAddr)
	require.NoError(err)
	require.Len(resUnbonding.Entries, 2)
	require.NotEqual(initialEntries, resUnbonding.Entries)
	require.NotEqual(resUnbonding.Entries[0], resUnbonding.Entries[1])
	require.Equal(creationHeight, resUnbonding.Entries[0].CreationHeight)
	require.Equal(newCreationHeight, resUnbonding.Entries[1].CreationHeight)

	// unbondingID is incremented on every call to SetUnbondingDelegationEntry
	// unbondingID == 1 was skipped because the entry was merged with the existing entry with unbondingID == 0
	// unbondingID comes from a global counter -> gaps in unbondingIDs are OK as long as every unbondingID is unique
	require.Equal(uint64(2), resUnbonding.Entries[1].UnbondingId)
}

func (s *KeeperTestSuite) TestGetUBDQueueTimeSlice() {
	testCases := []struct {
		name         string
		maxCacheSize int
		description  string
	}{
		{
			name:         "cache size < 0 (cache disabled)",
			maxCacheSize: -1,
			description:  "should always read from store when cache is not initialized",
		},
		{
			name:         "cache size = 0 (unlimited cache)",
			maxCacheSize: 0,
			description:  "should use unlimited cache with no size restrictions",
		},
		{
			name:         "cache size > unbonding delegation entries",
			maxCacheSize: 10,
			description:  "should use cache when cache is large enough",
		},
		{
			name:         "cache size == unbonding delegation entries",
			maxCacheSize: 3,
			description:  "should use cache when cache size matches entries",
		},
		{
			name:         "cache size < unbonding delegation entries",
			maxCacheSize: 1,
			description:  "should fallback to store when cache size is exceeded",
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			key := storetypes.NewKVStoreKey(stakingtypes.StoreKey)
			storeService := runtime.NewKVStoreService(key)
			testCtx := sdktestutil.DefaultContextWithDB(s.T(), key, storetypes.NewTransientStoreKey("transient_test"))
			ctx := testCtx.Ctx.WithBlockHeader(cmtproto.Header{Time: cmttime.Now()})
			encCfg := moduletestutil.MakeTestEncodingConfig()

			ctrl := gomock.NewController(s.T())
			accountKeeper := testutil.NewMockAccountKeeper(ctrl)
			accountKeeper.EXPECT().GetModuleAddress(stakingtypes.BondedPoolName).Return(bondedAcc.GetAddress()).AnyTimes()
			accountKeeper.EXPECT().GetModuleAddress(stakingtypes.NotBondedPoolName).Return(notBondedAcc.GetAddress()).AnyTimes()
			accountKeeper.EXPECT().AddressCodec().Return(address.NewBech32Codec("cosmos")).AnyTimes()

			bankKeeper := testutil.NewMockBankKeeper(ctrl)

			keeper := stakingkeeper.NewKeeper(
				encCfg.Codec,
				storeService,
				accountKeeper,
				bankKeeper,
				authtypes.NewModuleAddress(govtypes.ModuleName).String(),
				address.NewBech32Codec("cosmosvaloper"),
				address.NewBech32Codec("cosmosvalcons"),
				tc.maxCacheSize,
			)
			s.Require().NoError(keeper.SetParams(ctx, stakingtypes.DefaultParams()))

			blockTime := time.Now().UTC()
			blockHeight := int64(1000)
			ctx = ctx.WithBlockHeight(blockHeight).WithBlockTime(blockTime)

			delAddrs, valAddrs := createValAddrs(3)

			// Create multiple unbonding delegations with different completion times
			time1 := blockTime
			ubd1 := stakingtypes.NewUnbondingDelegation(
				delAddrs[0],
				valAddrs[0],
				blockHeight,
				time1,
				math.NewInt(10),
				0,
				address.NewBech32Codec("cosmosvaloper"),
				address.NewBech32Codec("cosmos"),
			)
			s.Require().NoError(keeper.InsertUBDQueue(ctx, ubd1, time1))

			time2 := blockTime.Add(1 * time.Hour)
			ubd2 := stakingtypes.NewUnbondingDelegation(
				delAddrs[1],
				valAddrs[1],
				blockHeight,
				time2,
				math.NewInt(20),
				0,
				address.NewBech32Codec("cosmosvaloper"),
				address.NewBech32Codec("cosmos"),
			)
			s.Require().NoError(keeper.InsertUBDQueue(ctx, ubd2, time2))

			time3 := blockTime.Add(2 * time.Hour)
			ubd3 := stakingtypes.NewUnbondingDelegation(
				delAddrs[2],
				valAddrs[2],
				blockHeight,
				time3,
				math.NewInt(30),
				0,
				address.NewBech32Codec("cosmosvaloper"),
				address.NewBech32Codec("cosmos"),
			)
			s.Require().NoError(keeper.InsertUBDQueue(ctx, ubd3, time3))

			// Test GetUBDQueueTimeSlice for time1
			slice1, err := keeper.GetUBDQueueTimeSlice(ctx, time1)
			s.Require().NoError(err)
			s.Require().Equal(1, len(slice1), "should have 1 entry at time1")
			s.Require().Equal(ubd1.DelegatorAddress, slice1[0].DelegatorAddress)
			s.Require().Equal(ubd1.ValidatorAddress, slice1[0].ValidatorAddress)

			// Test GetUBDQueueTimeSlice for time2
			slice2, err := keeper.GetUBDQueueTimeSlice(ctx, time2)
			s.Require().NoError(err)
			s.Require().Equal(1, len(slice2), "should have 1 entry at time2")
			s.Require().Equal(ubd2.DelegatorAddress, slice2[0].DelegatorAddress)
			s.Require().Equal(ubd2.ValidatorAddress, slice2[0].ValidatorAddress)

			// Test GetUBDQueueTimeSlice for time3
			slice3, err := keeper.GetUBDQueueTimeSlice(ctx, time3)
			s.Require().NoError(err)
			s.Require().Equal(1, len(slice3), "should have 1 entry at time3")
			s.Require().Equal(ubd3.DelegatorAddress, slice3[0].DelegatorAddress)
			s.Require().Equal(ubd3.ValidatorAddress, slice3[0].ValidatorAddress)

			// Test calling again to verify cache consistency
			slice1Again, err := keeper.GetUBDQueueTimeSlice(ctx, time1)
			s.Require().NoError(err)
			s.Require().Equal(len(slice1), len(slice1Again), "repeated call should return same number of entries")
			s.Require().Equal(slice1[0].DelegatorAddress, slice1Again[0].DelegatorAddress)

			// Test for non-existent time (should return empty slice)
			emptyTime := blockTime.Add(-1 * time.Hour)
			emptySlice, err := keeper.GetUBDQueueTimeSlice(ctx, emptyTime)
			s.Require().NoError(err)
			s.Require().Equal(0, len(emptySlice), "should have 0 entries at non-existent time")
		})
	}
}

func (s *KeeperTestSuite) TestGetAllUnbondingDelegations() {
	testCases := []struct {
		name         string
		maxCacheSize int
		description  string
	}{
		{
			name:         "cache size < 0 (cache disabled)",
			maxCacheSize: -1,
			description:  "should always read from store when cache is not initialized",
		},
		{
			name:         "cache size = 0 (unlimited cache)",
			maxCacheSize: 0,
			description:  "should use unlimited cache with no size restrictions",
		},
		{
			name:         "cache size > unbonding delegation entries",
			maxCacheSize: 10,
			description:  "should use cache when cache is large enough",
		},
		{
			name:         "cache size == unbonding delegation entries",
			maxCacheSize: 2,
			description:  "should use cache when cache size matches entries",
		},
		{
			name:         "cache size < unbonding delegation entries",
			maxCacheSize: 1,
			description:  "should fallback to store when cache size is exceeded",
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			key := storetypes.NewKVStoreKey(stakingtypes.StoreKey)
			storeService := runtime.NewKVStoreService(key)
			testCtx := sdktestutil.DefaultContextWithDB(s.T(), key, storetypes.NewTransientStoreKey("transient_test"))
			ctx := testCtx.Ctx.WithBlockHeader(cmtproto.Header{Time: cmttime.Now()})
			encCfg := moduletestutil.MakeTestEncodingConfig()

			ctrl := gomock.NewController(s.T())
			accountKeeper := testutil.NewMockAccountKeeper(ctrl)
			accountKeeper.EXPECT().GetModuleAddress(stakingtypes.BondedPoolName).Return(bondedAcc.GetAddress()).AnyTimes()
			accountKeeper.EXPECT().GetModuleAddress(stakingtypes.NotBondedPoolName).Return(notBondedAcc.GetAddress()).AnyTimes()
			accountKeeper.EXPECT().AddressCodec().Return(address.NewBech32Codec("cosmos")).AnyTimes()

			bankKeeper := testutil.NewMockBankKeeper(ctrl)

			keeper := stakingkeeper.NewKeeper(
				encCfg.Codec,
				storeService,
				accountKeeper,
				bankKeeper,
				authtypes.NewModuleAddress(govtypes.ModuleName).String(),
				address.NewBech32Codec("cosmosvaloper"),
				address.NewBech32Codec("cosmosvalcons"),
				tc.maxCacheSize,
			)
			s.Require().NoError(keeper.SetParams(ctx, stakingtypes.DefaultParams()))

			blockTime := time.Now().UTC()
			blockHeight := int64(1000)
			ctx = ctx.WithBlockHeight(blockHeight).WithBlockTime(blockTime)

			delAddrs, valAddrs := createValAddrs(2)

			// insert unbonding delegation
			ubd := stakingtypes.NewUnbondingDelegation(
				delAddrs[0],
				valAddrs[0],
				blockHeight,
				blockTime,
				math.NewInt(10),
				0,
				address.NewBech32Codec("cosmosvaloper"),
				address.NewBech32Codec("cosmos"),
			)

			t := blockTime
			s.Require().NoError(keeper.InsertUBDQueue(ctx, ubd, t))

			// add another unbonding delegation
			ubd1 := stakingtypes.NewUnbondingDelegation(
				delAddrs[1],
				valAddrs[1],
				blockHeight,
				blockTime,
				math.NewInt(10),
				0,
				address.NewBech32Codec("cosmosvaloper"),
				address.NewBech32Codec("cosmos"),
			)
			t1 := blockTime.Add(-1 * time.Minute)
			s.Require().NoError(keeper.InsertUBDQueue(ctx, ubd1, t1))

			// get all unbonding delegations should return the inserted unbonding delegations
			unbondingDelegations, err := keeper.GetUBDs(ctx, blockTime)
			s.Require().NoError(err)
			s.Require().Equal(2, len(unbondingDelegations))
			s.Require().Equal(ubd.DelegatorAddress, unbondingDelegations[sdk.FormatTimeString(t)][0].DelegatorAddress)
			s.Require().Equal(ubd.ValidatorAddress, unbondingDelegations[sdk.FormatTimeString(t)][0].ValidatorAddress)
			s.Require().Equal(ubd1.DelegatorAddress, unbondingDelegations[sdk.FormatTimeString(t1)][0].DelegatorAddress)
			s.Require().Equal(ubd1.ValidatorAddress, unbondingDelegations[sdk.FormatTimeString(t1)][0].ValidatorAddress)

			// Test calling again to verify cache consistency
			unbondingDelegations2, err := keeper.GetUBDs(ctx, blockTime)
			s.Require().NoError(err)
			s.Require().Equal(len(unbondingDelegations), len(unbondingDelegations2), "repeated call should return same number of entries")
		})
	}
}

func (s *KeeperTestSuite) TestInsertUBDQueue() {
	testCases := []struct {
		name         string
		maxCacheSize int
		description  string
	}{
		{
			name:         "cache size < 0 (cache disabled)",
			maxCacheSize: -1,
			description:  "should always write to store when cache is not initialized",
		},
		{
			name:         "cache size = 0 (unlimited cache)",
			maxCacheSize: 0,
			description:  "should use unlimited cache with no size restrictions",
		},
		{
			name:         "cache size > unbonding delegation entries",
			maxCacheSize: 10,
			description:  "should use cache when cache is large enough",
		},
		{
			name:         "cache size == unbonding delegation entries",
			maxCacheSize: 2,
			description:  "should use cache when cache size matches entries",
		},
		{
			name:         "cache size < unbonding delegation entries",
			maxCacheSize: 1,
			description:  "should fallback to store when cache size is exceeded",
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			key := storetypes.NewKVStoreKey(stakingtypes.StoreKey)
			storeService := runtime.NewKVStoreService(key)
			testCtx := sdktestutil.DefaultContextWithDB(s.T(), key, storetypes.NewTransientStoreKey("transient_test"))
			ctx := testCtx.Ctx.WithBlockHeader(cmtproto.Header{Time: cmttime.Now()})
			encCfg := moduletestutil.MakeTestEncodingConfig()

			ctrl := gomock.NewController(s.T())
			accountKeeper := testutil.NewMockAccountKeeper(ctrl)
			accountKeeper.EXPECT().GetModuleAddress(stakingtypes.BondedPoolName).Return(bondedAcc.GetAddress()).AnyTimes()
			accountKeeper.EXPECT().GetModuleAddress(stakingtypes.NotBondedPoolName).Return(notBondedAcc.GetAddress()).AnyTimes()
			accountKeeper.EXPECT().AddressCodec().Return(address.NewBech32Codec("cosmos")).AnyTimes()

			bankKeeper := testutil.NewMockBankKeeper(ctrl)

			keeper := stakingkeeper.NewKeeper(
				encCfg.Codec,
				storeService,
				accountKeeper,
				bankKeeper,
				authtypes.NewModuleAddress(govtypes.ModuleName).String(),
				address.NewBech32Codec("cosmosvaloper"),
				address.NewBech32Codec("cosmosvalcons"),
				tc.maxCacheSize,
			)
			s.Require().NoError(keeper.SetParams(ctx, stakingtypes.DefaultParams()))

			blockTime := time.Now().UTC()
			blockHeight := int64(1000)
			ctx = ctx.WithBlockHeight(blockHeight).WithBlockTime(blockTime)

			iterator, err := keeper.UBDQueueIterator(ctx, blockTime)
			s.Require().NoError(err)
			defer iterator.Close()
			count := 0
			for ; iterator.Valid(); iterator.Next() {
				count++
			}
			// no unbonding delegations in the queue initially
			s.Require().Equal(0, count)

			delAddrs, valAddrs := createValAddrs(3)

			// insert unbonding delegation
			ubd := stakingtypes.NewUnbondingDelegation(
				delAddrs[0],
				valAddrs[0],
				blockHeight,
				blockTime,
				math.NewInt(10),
				0,
				address.NewBech32Codec("cosmosvaloper"),
				address.NewBech32Codec("cosmos"),
			)

			t := blockTime
			s.Require().NoError(keeper.InsertUBDQueue(ctx, ubd, t))

			// insert another unbonding delegation
			ubd1 := stakingtypes.NewUnbondingDelegation(
				delAddrs[1],
				valAddrs[1],
				blockHeight,
				blockTime,
				math.NewInt(10),
				0,
				address.NewBech32Codec("cosmosvaloper"),
				address.NewBech32Codec("cosmos"),
			)

			s.Require().NoError(keeper.InsertUBDQueue(ctx, ubd1, t))

			iterator1, err := keeper.UBDQueueIterator(ctx, blockTime)
			s.Require().NoError(err)
			defer iterator1.Close()
			count1 := 0
			for ; iterator1.Valid(); iterator1.Next() {
				count1++
			}

			// unbonding delegation should be retrieved
			// count 1 due to same unbonding time
			s.Require().Equal(1, count1)

			// Verify GetUBDQueueTimeSlice returns the correct unbonding delegations after insertion
			ubds, err := keeper.GetUBDQueueTimeSlice(ctx, blockTime)
			s.Require().NoError(err)
			s.Require().Equal(2, len(ubds), "should have 2 unbonding delegations at same time")

			// insert unbonding delegation with different unbonding time and height
			ubd2 := stakingtypes.NewUnbondingDelegation(
				delAddrs[2],
				valAddrs[2],
				blockHeight,
				blockTime,
				math.NewInt(10),
				0,
				address.NewBech32Codec("cosmosvaloper"),
				address.NewBech32Codec("cosmos"),
			)
			t1 := blockTime.Add(-1 * time.Minute)
			s.Require().NoError(keeper.InsertUBDQueue(ctx, ubd2, t1))

			iterator2, err := keeper.UBDQueueIterator(ctx, blockTime)
			s.Require().NoError(err)
			defer iterator2.Close()
			count2 := 0
			for ; iterator2.Valid(); iterator2.Next() {
				count2++
			}

			// unbonding delegation should be retrieved
			s.Require().Equal(2, count2)

			// Verify the new unbonding delegation was inserted at the correct time
			ubds2, err := keeper.GetUBDQueueTimeSlice(ctx, t1)
			s.Require().NoError(err)
			s.Require().Equal(1, len(ubds2), "should have 1 unbonding delegation at different time")
		})
	}
}

func (s *KeeperTestSuite) TestDequeueAllMatureUBDQueue() {
	testCases := []struct {
		name                    string
		maxCacheSize            int
		numUnbondingDelegations int
	}{
		{
			name:                    "cache size < 0 (cache disabled)",
			maxCacheSize:            -1,
			numUnbondingDelegations: 3,
		},
		{
			name:                    "cache size = 0 (unlimited cache)",
			maxCacheSize:            0,
			numUnbondingDelegations: 3,
		},
		{
			name:                    "cache size > unbonding delegations",
			maxCacheSize:            5,
			numUnbondingDelegations: 2,
		},
		{
			name:                    "cache size == unbonding delegations",
			maxCacheSize:            2,
			numUnbondingDelegations: 2,
		},
		{
			name:                    "cache size < unbonding delegations",
			maxCacheSize:            1,
			numUnbondingDelegations: 3,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			key := storetypes.NewKVStoreKey(stakingtypes.StoreKey)
			storeService := runtime.NewKVStoreService(key)
			testCtx := sdktestutil.DefaultContextWithDB(s.T(), key, storetypes.NewTransientStoreKey("transient_test"))
			ctx := testCtx.Ctx.WithBlockHeader(cmtproto.Header{Time: cmttime.Now()})
			encCfg := moduletestutil.MakeTestEncodingConfig()

			ctrl := gomock.NewController(s.T())
			accountKeeper := testutil.NewMockAccountKeeper(ctrl)
			accountKeeper.EXPECT().GetModuleAddress(stakingtypes.BondedPoolName).Return(bondedAcc.GetAddress()).AnyTimes()
			accountKeeper.EXPECT().GetModuleAddress(stakingtypes.NotBondedPoolName).Return(notBondedAcc.GetAddress()).AnyTimes()
			accountKeeper.EXPECT().AddressCodec().Return(address.NewBech32Codec("cosmos")).AnyTimes()

			bankKeeper := testutil.NewMockBankKeeper(ctrl)
			bankKeeper.EXPECT().DelegateCoinsFromAccountToModule(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
			bankKeeper.EXPECT().SendCoinsFromModuleToModule(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
			bankKeeper.EXPECT().UndelegateCoinsFromModuleToAccount(gomock.Any(), stakingtypes.NotBondedPoolName, gomock.Any(), gomock.Any()).AnyTimes()

			// Initialize keeper with specific cache size
			keeper := stakingkeeper.NewKeeper(
				encCfg.Codec,
				storeService,
				accountKeeper,
				bankKeeper,
				authtypes.NewModuleAddress(govtypes.ModuleName).String(),
				address.NewBech32Codec("cosmosvaloper"),
				address.NewBech32Codec("cosmosvalcons"),
				tc.maxCacheSize,
			)
			params := stakingtypes.DefaultParams()
			params.UnbondingTime = 1 * time.Second
			s.Require().NoError(keeper.SetParams(ctx, params))

			blockTime := time.Now().UTC()
			ctx = ctx.WithBlockTime(blockTime)

			// Create validator
			valAddr := sdk.ValAddress(PKs[0].Address())
			validator := testutil.NewValidator(s.T(), valAddr, PKs[0])
			validator, _ = validator.AddTokensFromDel(keeper.TokensFromConsensusPower(ctx, 100))
			validator = stakingkeeper.TestingUpdateValidator(keeper, ctx, validator, true)

			// Create multiple unbonding delegations
			delAddrs, _ := createValAddrs(tc.numUnbondingDelegations)
			for i := 0; i < tc.numUnbondingDelegations; i++ {
				// Delegate
				bondAmt := keeper.TokensFromConsensusPower(ctx, 10)
				_, err := keeper.Delegate(ctx, delAddrs[i], bondAmt, stakingtypes.Unbonded, validator, true)
				s.Require().NoError(err)

				// Undelegate
				_, _, err = keeper.Undelegate(ctx, delAddrs[i], valAddr, math.LegacyNewDec(5))
				s.Require().NoError(err)
			}

			// Verify unbonding delegations were created
			for i := 0; i < tc.numUnbondingDelegations; i++ {
				_, err := keeper.GetUnbondingDelegation(ctx, delAddrs[i], valAddr)
				s.Require().NoError(err)
			}

			// Fast-forward time to maturity
			ctx = ctx.WithBlockTime(blockTime.Add(params.UnbondingTime))

			// Verify GetUBDs returns the expected number of unbonding delegations
			allUBDs, err := keeper.GetUBDs(ctx, ctx.BlockTime())
			s.Require().NoError(err)
			s.Require().NotEmpty(allUBDs)

			// Verify GetUBDQueueTimeSlice returns the expected number of unbonding delegations
			// In this case, it should return all unbonding delegations as all unbonding delegations are at the same time.
			ubds, err := keeper.GetUBDQueueTimeSlice(ctx, ctx.BlockTime())
			s.Require().NoError(err)
			s.Require().Equal(tc.numUnbondingDelegations, len(ubds))

			// Dequeue and complete all mature unbonding delegations
			matureUnbonds, err := keeper.DequeueAllMatureUBDQueue(ctx, ctx.BlockTime())
			s.Require().NoError(err)
			s.Require().Equal(tc.numUnbondingDelegations, len(matureUnbonds), "all unbonding delegations should be mature")

			// Complete the unbonding delegations
			for _, dvPair := range matureUnbonds {
				delAddr, err := accountKeeper.AddressCodec().StringToBytes(dvPair.DelegatorAddress)
				s.Require().NoError(err)
				valAddr, err := keeper.ValidatorAddressCodec().StringToBytes(dvPair.ValidatorAddress)
				s.Require().NoError(err)
				_, err = keeper.CompleteUnbonding(ctx, delAddr, valAddr)
				s.Require().NoError(err)
			}

			// Verify all unbonding delegations were completed (removed from store)
			for i := 0; i < tc.numUnbondingDelegations; i++ {
				_, err := keeper.GetUnbondingDelegation(ctx, delAddrs[i], valAddr)
				s.Require().ErrorIs(err, stakingtypes.ErrNoUnbondingDelegation, "unbonding delegation should be completed and removed")
			}
		})
	}
}

func (s *KeeperTestSuite) TestUnbondingDelegationQueueCacheRecovery() {
	// This test verifies that when the cache is initially too small (exceeded),
	// and then entries are dequeued, the cache can recover and be used again
	// Cache size is based on the number of unique timestamps (keys), not individual entries
	key := storetypes.NewKVStoreKey(stakingtypes.StoreKey)
	storeService := runtime.NewKVStoreService(key)
	testCtx := sdktestutil.DefaultContextWithDB(s.T(), key, storetypes.NewTransientStoreKey("transient_test"))
	ctx := testCtx.Ctx.WithBlockHeader(cmtproto.Header{Time: cmttime.Now()})
	encCfg := moduletestutil.MakeTestEncodingConfig()

	ctrl := gomock.NewController(s.T())
	accountKeeper := testutil.NewMockAccountKeeper(ctrl)
	accountKeeper.EXPECT().GetModuleAddress(stakingtypes.BondedPoolName).Return(bondedAcc.GetAddress()).AnyTimes()
	accountKeeper.EXPECT().GetModuleAddress(stakingtypes.NotBondedPoolName).Return(notBondedAcc.GetAddress()).AnyTimes()
	accountKeeper.EXPECT().AddressCodec().Return(address.NewBech32Codec("cosmos")).AnyTimes()

	bankKeeper := testutil.NewMockBankKeeper(ctrl)
	bankKeeper.EXPECT().DelegateCoinsFromAccountToModule(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
	bankKeeper.EXPECT().SendCoinsFromModuleToModule(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	bankKeeper.EXPECT().UndelegateCoinsFromModuleToAccount(gomock.Any(), stakingtypes.NotBondedPoolName, gomock.Any(), gomock.Any()).AnyTimes()

	// Initialize keeper with small cache size (2 timestamps)
	maxCacheSize := 2
	keeper := stakingkeeper.NewKeeper(
		encCfg.Codec,
		storeService,
		accountKeeper,
		bankKeeper,
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
		address.NewBech32Codec("cosmosvaloper"),
		address.NewBech32Codec("cosmosvalcons"),
		maxCacheSize,
	)
	params := stakingtypes.DefaultParams()
	params.UnbondingTime = 1 * time.Hour // Use a long enough time (1 hr) so we can create different timestamps
	s.Require().NoError(keeper.SetParams(ctx, params))

	baseTime := time.Now().UTC()
	ctx = ctx.WithBlockTime(baseTime)

	// Create validator
	valAddr := sdk.ValAddress(PKs[0].Address())
	validator := testutil.NewValidator(s.T(), valAddr, PKs[0])
	validator, _ = validator.AddTokensFromDel(keeper.TokensFromConsensusPower(ctx, 100))
	validator = stakingkeeper.TestingUpdateValidator(keeper, ctx, validator, true)

	// Create unbonding delegations at 5 different timestamps (exceeds cache size of 2)
	numTimestamps := 5
	delAddrs, _ := createValAddrs(numTimestamps)
	completionTimes := make([]time.Time, numTimestamps)

	for i := 0; i < numTimestamps; i++ {
		// Set different block times to create different completion timestamps
		currentTime := baseTime.Add(time.Duration(i) * time.Second)
		ctx = ctx.WithBlockTime(currentTime)
		completionTimes[i] = currentTime.Add(params.UnbondingTime)

		// Delegate
		bondAmt := keeper.TokensFromConsensusPower(ctx, 10)
		_, err := keeper.Delegate(ctx, delAddrs[i], bondAmt, stakingtypes.Unbonded, validator, true)
		s.Require().NoError(err)

		// Undelegate (will create unbonding delegation with unique completion time)
		_, _, err = keeper.Undelegate(ctx, delAddrs[i], valAddr, math.LegacyNewDec(5))
		s.Require().NoError(err)
	}

	// Verify all unbonding delegations were created
	for i := 0; i < numTimestamps; i++ {
		_, err := keeper.GetUnbondingDelegation(ctx, delAddrs[i], valAddr)
		s.Require().NoError(err)
	}

	// At this point, cache should be exceeded (5 timestamps > maxCacheSize of 2)
	// GetUBDs should still work, but will read from store instead of cache
	ctx = ctx.WithBlockTime(completionTimes[numTimestamps-1])
	allUBDs, err := keeper.GetUBDs(ctx, ctx.BlockTime())
	s.Require().NoError(err)
	s.Require().Equal(5, len(allUBDs), "should have 5 different timestamps")

	// Fast-forward time to mature the first 3 timestamps and dequeue them
	ctx = ctx.WithBlockTime(completionTimes[2])
	matureUnbonds, err := keeper.DequeueAllMatureUBDQueue(ctx, ctx.BlockTime())
	s.Require().NoError(err)
	s.Require().Equal(3, len(matureUnbonds), "should dequeue 3 unbonding delegations from 3 timestamps")

	// Complete the first 3 unbonding delegations
	for i := 0; i < 3; i++ {
		_, err = keeper.CompleteUnbonding(ctx, delAddrs[i], valAddr)
		s.Require().NoError(err)
	}

	// Verify the first 3 were removed
	for i := 0; i < 3; i++ {
		_, err := keeper.GetUnbondingDelegation(ctx, delAddrs[i], valAddr)
		s.Require().ErrorIs(err, stakingtypes.ErrNoUnbondingDelegation, "unbonding delegation should be completed and removed")
	}

	// Verify the last 2 still exist
	for i := 3; i < numTimestamps; i++ {
		_, err := keeper.GetUnbondingDelegation(ctx, delAddrs[i], valAddr)
		s.Require().NoError(err, "unbonding delegation should still exist")
	}

	// Now only 2 timestamps remain (completionTimes[3] and completionTimes[4])
	// This fits in the cache (2 timestamps == maxCacheSize)
	// GetUBDs should now be able to use the cache
	ctx = ctx.WithBlockTime(completionTimes[4])
	remainingUBDs, err := keeper.GetUBDs(ctx, ctx.BlockTime())
	s.Require().NoError(err)
	s.Require().Equal(2, len(remainingUBDs), "should have 2 timestamps in cache")

	// Dequeue the remaining 2
	finalMatureUnbonds, err := keeper.DequeueAllMatureUBDQueue(ctx, ctx.BlockTime())
	s.Require().NoError(err)
	s.Require().Equal(2, len(finalMatureUnbonds), "should have 2 mature unbonding delegations")

	// Complete them
	for i := 3; i < numTimestamps; i++ {
		_, err = keeper.CompleteUnbonding(ctx, delAddrs[i], valAddr)
		s.Require().NoError(err)
	}

	// Verify all unbonding delegations are now completed
	for i := 3; i < numTimestamps; i++ {
		_, err := keeper.GetUnbondingDelegation(ctx, delAddrs[i], valAddr)
		s.Require().ErrorIs(err, stakingtypes.ErrNoUnbondingDelegation, "all unbonding delegations should be completed")
	}
}

func (s *KeeperTestSuite) TestGetAndParseUnbondingDelegationTimeKey() {
	require := s.Require()

	blockTime := time.Now().UTC()
	key := stakingtypes.GetUnbondingDelegationTimeKey(blockTime)
	time, err := stakingtypes.ParseUnbondingDelegationTimeKey(key)
	require.NoError(err)
	require.Equal(blockTime, time)
}

func (s *KeeperTestSuite) TestGetRedelegationQueueTimeSlice() {
	testCases := []struct {
		name         string
		maxCacheSize int
		description  string
	}{
		{
			name:         "cache size < 0 (cache disabled)",
			maxCacheSize: -1,
			description:  "should always read from store when cache is not initialized",
		},
		{
			name:         "cache size = 0 (unlimited cache)",
			maxCacheSize: 0,
			description:  "should use unlimited cache with no size restrictions",
		},
		{
			name:         "cache size > redelegation entries",
			maxCacheSize: 10,
			description:  "should use cache when cache is large enough",
		},
		{
			name:         "cache size == redelegation entries",
			maxCacheSize: 3,
			description:  "should use cache when cache size matches entries",
		},
		{
			name:         "cache size < redelegation entries",
			maxCacheSize: 1,
			description:  "should fallback to store when cache size is exceeded",
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			key := storetypes.NewKVStoreKey(stakingtypes.StoreKey)
			storeService := runtime.NewKVStoreService(key)
			testCtx := sdktestutil.DefaultContextWithDB(s.T(), key, storetypes.NewTransientStoreKey("transient_test"))
			ctx := testCtx.Ctx.WithBlockHeader(cmtproto.Header{Time: cmttime.Now()})
			encCfg := moduletestutil.MakeTestEncodingConfig()

			ctrl := gomock.NewController(s.T())
			accountKeeper := testutil.NewMockAccountKeeper(ctrl)
			accountKeeper.EXPECT().GetModuleAddress(stakingtypes.BondedPoolName).Return(bondedAcc.GetAddress()).AnyTimes()
			accountKeeper.EXPECT().GetModuleAddress(stakingtypes.NotBondedPoolName).Return(notBondedAcc.GetAddress()).AnyTimes()
			accountKeeper.EXPECT().AddressCodec().Return(address.NewBech32Codec("cosmos")).AnyTimes()

			bankKeeper := testutil.NewMockBankKeeper(ctrl)

			keeper := stakingkeeper.NewKeeper(
				encCfg.Codec,
				storeService,
				accountKeeper,
				bankKeeper,
				authtypes.NewModuleAddress(govtypes.ModuleName).String(),
				address.NewBech32Codec("cosmosvaloper"),
				address.NewBech32Codec("cosmosvalcons"),
				tc.maxCacheSize,
			)
			s.Require().NoError(keeper.SetParams(ctx, stakingtypes.DefaultParams()))

			blockTime := time.Now().UTC()
			blockHeight := int64(1000)
			ctx = ctx.WithBlockHeight(blockHeight).WithBlockTime(blockTime)

			delAddrs, valAddrs := createValAddrs(4)

			// Create multiple redelegations with different completion times
			time1 := blockTime
			red1 := stakingtypes.Redelegation{
				DelegatorAddress:    delAddrs[0].String(),
				ValidatorSrcAddress: valAddrs[0].String(),
				ValidatorDstAddress: valAddrs[1].String(),
			}
			s.Require().NoError(keeper.InsertRedelegationQueue(ctx, red1, time1))

			time2 := blockTime.Add(1 * time.Hour)
			red2 := stakingtypes.Redelegation{
				DelegatorAddress:    delAddrs[1].String(),
				ValidatorSrcAddress: valAddrs[1].String(),
				ValidatorDstAddress: valAddrs[2].String(),
			}
			s.Require().NoError(keeper.InsertRedelegationQueue(ctx, red2, time2))

			time3 := blockTime.Add(2 * time.Hour)
			red3 := stakingtypes.Redelegation{
				DelegatorAddress:    delAddrs[2].String(),
				ValidatorSrcAddress: valAddrs[2].String(),
				ValidatorDstAddress: valAddrs[3].String(),
			}
			s.Require().NoError(keeper.InsertRedelegationQueue(ctx, red3, time3))

			// Test GetRedelegationQueueTimeSlice for time1
			slice1, err := keeper.GetRedelegationQueueTimeSlice(ctx, time1)
			s.Require().NoError(err)
			s.Require().Equal(1, len(slice1), "should have 1 entry at time1")
			s.Require().Equal(red1.DelegatorAddress, slice1[0].DelegatorAddress)
			s.Require().Equal(red1.ValidatorSrcAddress, slice1[0].ValidatorSrcAddress)
			s.Require().Equal(red1.ValidatorDstAddress, slice1[0].ValidatorDstAddress)

			// Test GetRedelegationQueueTimeSlice for time2
			slice2, err := keeper.GetRedelegationQueueTimeSlice(ctx, time2)
			s.Require().NoError(err)
			s.Require().Equal(1, len(slice2), "should have 1 entry at time2")
			s.Require().Equal(red2.DelegatorAddress, slice2[0].DelegatorAddress)
			s.Require().Equal(red2.ValidatorSrcAddress, slice2[0].ValidatorSrcAddress)
			s.Require().Equal(red2.ValidatorDstAddress, slice2[0].ValidatorDstAddress)

			// Test GetRedelegationQueueTimeSlice for time3
			slice3, err := keeper.GetRedelegationQueueTimeSlice(ctx, time3)
			s.Require().NoError(err)
			s.Require().Equal(1, len(slice3), "should have 1 entry at time3")
			s.Require().Equal(red3.DelegatorAddress, slice3[0].DelegatorAddress)
			s.Require().Equal(red3.ValidatorSrcAddress, slice3[0].ValidatorSrcAddress)
			s.Require().Equal(red3.ValidatorDstAddress, slice3[0].ValidatorDstAddress)

			// Test calling again to verify cache consistency
			slice1Again, err := keeper.GetRedelegationQueueTimeSlice(ctx, time1)
			s.Require().NoError(err)
			s.Require().Equal(len(slice1), len(slice1Again), "repeated call should return same number of entries")
			s.Require().Equal(slice1[0].DelegatorAddress, slice1Again[0].DelegatorAddress)

			// Test for non-existent time (should return empty slice)
			emptyTime := blockTime.Add(-1 * time.Hour)
			emptySlice, err := keeper.GetRedelegationQueueTimeSlice(ctx, emptyTime)
			s.Require().NoError(err)
			s.Require().Equal(0, len(emptySlice), "should have 0 entries at non-existent time")
		})
	}
}

func (s *KeeperTestSuite) TestGetPendingRedelegations() {
	testCases := []struct {
		name         string
		maxCacheSize int
		description  string
	}{
		{
			name:         "cache size < 0 (cache disabled)",
			maxCacheSize: -1,
			description:  "should always read from store when cache is not initialized",
		},
		{
			name:         "cache size = 0 (unlimited cache)",
			maxCacheSize: 0,
			description:  "should use unlimited cache with no size restrictions",
		},
		{
			name:         "cache size > redelegation entries",
			maxCacheSize: 10,
			description:  "should use cache when cache is large enough",
		},
		{
			name:         "cache size == redelegation entries",
			maxCacheSize: 2,
			description:  "should use cache when cache size matches entries",
		},
		{
			name:         "cache size < redelegation entries",
			maxCacheSize: 1,
			description:  "should fallback to store when cache size is exceeded",
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			key := storetypes.NewKVStoreKey(stakingtypes.StoreKey)
			storeService := runtime.NewKVStoreService(key)
			testCtx := sdktestutil.DefaultContextWithDB(s.T(), key, storetypes.NewTransientStoreKey("transient_test"))
			ctx := testCtx.Ctx.WithBlockHeader(cmtproto.Header{Time: cmttime.Now()})
			encCfg := moduletestutil.MakeTestEncodingConfig()

			ctrl := gomock.NewController(s.T())
			accountKeeper := testutil.NewMockAccountKeeper(ctrl)
			accountKeeper.EXPECT().GetModuleAddress(stakingtypes.BondedPoolName).Return(bondedAcc.GetAddress()).AnyTimes()
			accountKeeper.EXPECT().GetModuleAddress(stakingtypes.NotBondedPoolName).Return(notBondedAcc.GetAddress()).AnyTimes()
			accountKeeper.EXPECT().AddressCodec().Return(address.NewBech32Codec("cosmos")).AnyTimes()

			bankKeeper := testutil.NewMockBankKeeper(ctrl)

			keeper := stakingkeeper.NewKeeper(
				encCfg.Codec,
				storeService,
				accountKeeper,
				bankKeeper,
				authtypes.NewModuleAddress(govtypes.ModuleName).String(),
				address.NewBech32Codec("cosmosvaloper"),
				address.NewBech32Codec("cosmosvalcons"),
				tc.maxCacheSize,
			)
			s.Require().NoError(keeper.SetParams(ctx, stakingtypes.DefaultParams()))

			blockTime := time.Now().UTC()
			blockHeight := int64(1000)
			ctx = ctx.WithBlockHeight(blockHeight).WithBlockTime(blockTime)

			delAddrs, valAddrs := createValAddrs(2)

			// insert redelegation
			red := stakingtypes.Redelegation{
				DelegatorAddress:    delAddrs[0].String(),
				ValidatorSrcAddress: valAddrs[0].String(),
				ValidatorDstAddress: valAddrs[1].String(),
			}

			t := blockTime
			s.Require().NoError(keeper.InsertRedelegationQueue(ctx, red, t))

			// add another redelegation
			red1 := stakingtypes.Redelegation{
				DelegatorAddress:    delAddrs[1].String(),
				ValidatorSrcAddress: valAddrs[1].String(),
				ValidatorDstAddress: valAddrs[0].String(),
			}
			t1 := blockTime.Add(-1 * time.Minute)
			s.Require().NoError(keeper.InsertRedelegationQueue(ctx, red1, t1))

			// get all redelegations should return the inserted redelegations
			redelegations, err := keeper.GetPendingRedelegations(ctx, blockTime)
			s.Require().NoError(err)
			s.Require().Equal(2, len(redelegations))
			s.Require().Equal(red.DelegatorAddress, redelegations[sdk.FormatTimeString(t)][0].DelegatorAddress)
			s.Require().Equal(red.ValidatorSrcAddress, redelegations[sdk.FormatTimeString(t)][0].ValidatorSrcAddress)
			s.Require().Equal(red.ValidatorDstAddress, redelegations[sdk.FormatTimeString(t)][0].ValidatorDstAddress)
			s.Require().Equal(red1.DelegatorAddress, redelegations[sdk.FormatTimeString(t1)][0].DelegatorAddress)
			s.Require().Equal(red1.ValidatorSrcAddress, redelegations[sdk.FormatTimeString(t1)][0].ValidatorSrcAddress)
			s.Require().Equal(red1.ValidatorDstAddress, redelegations[sdk.FormatTimeString(t1)][0].ValidatorDstAddress)

			// Test calling again to verify cache consistency
			redelegations2, err := keeper.GetPendingRedelegations(ctx, blockTime)
			s.Require().NoError(err)
			s.Require().Equal(len(redelegations), len(redelegations2), "repeated call should return same number of entries")
		})
	}
}

func (s *KeeperTestSuite) TestInsertRedelegationQueue() {
	testCases := []struct {
		name         string
		maxCacheSize int
		description  string
	}{
		{
			name:         "cache size < 0 (cache disabled)",
			maxCacheSize: -1,
			description:  "should always write to store when cache is not initialized",
		},
		{
			name:         "cache size = 0 (unlimited cache)",
			maxCacheSize: 0,
			description:  "should use unlimited cache with no size restrictions",
		},
		{
			name:         "cache size > redelegation entries",
			maxCacheSize: 10,
			description:  "should use cache when cache is large enough",
		},
		{
			name:         "cache size == redelegation entries",
			maxCacheSize: 2,
			description:  "should use cache when cache size matches entries",
		},
		{
			name:         "cache size < redelegation entries",
			maxCacheSize: 1,
			description:  "should fallback to store when cache size is exceeded",
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			key := storetypes.NewKVStoreKey(stakingtypes.StoreKey)
			storeService := runtime.NewKVStoreService(key)
			testCtx := sdktestutil.DefaultContextWithDB(s.T(), key, storetypes.NewTransientStoreKey("transient_test"))
			ctx := testCtx.Ctx.WithBlockHeader(cmtproto.Header{Time: cmttime.Now()})
			encCfg := moduletestutil.MakeTestEncodingConfig()

			ctrl := gomock.NewController(s.T())
			accountKeeper := testutil.NewMockAccountKeeper(ctrl)
			accountKeeper.EXPECT().GetModuleAddress(stakingtypes.BondedPoolName).Return(bondedAcc.GetAddress()).AnyTimes()
			accountKeeper.EXPECT().GetModuleAddress(stakingtypes.NotBondedPoolName).Return(notBondedAcc.GetAddress()).AnyTimes()
			accountKeeper.EXPECT().AddressCodec().Return(address.NewBech32Codec("cosmos")).AnyTimes()

			bankKeeper := testutil.NewMockBankKeeper(ctrl)

			keeper := stakingkeeper.NewKeeper(
				encCfg.Codec,
				storeService,
				accountKeeper,
				bankKeeper,
				authtypes.NewModuleAddress(govtypes.ModuleName).String(),
				address.NewBech32Codec("cosmosvaloper"),
				address.NewBech32Codec("cosmosvalcons"),
				tc.maxCacheSize,
			)
			s.Require().NoError(keeper.SetParams(ctx, stakingtypes.DefaultParams()))

			blockTime := time.Now().UTC()
			blockHeight := int64(1000)
			ctx = ctx.WithBlockHeight(blockHeight).WithBlockTime(blockTime)

			iterator, err := keeper.RedelegationQueueIterator(ctx, blockTime)
			s.Require().NoError(err)
			defer iterator.Close()
			count := 0
			for ; iterator.Valid(); iterator.Next() {
				count++
			}
			// no redelegations in the queue initially
			s.Require().Equal(0, count)

			delAddrs, valAddrs := createValAddrs(3)

			// insert redelegation
			red := stakingtypes.NewRedelegation(delAddrs[0], valAddrs[0], valAddrs[1], 0,
				time.Unix(0, 0), math.NewInt(5),
				math.LegacyNewDec(5), 0, address.NewBech32Codec("cosmosvaloper"), address.NewBech32Codec("cosmos"))

			t := blockTime
			s.Require().NoError(keeper.InsertRedelegationQueue(ctx, red, t))

			// insert another redelegation
			red1 := stakingtypes.NewRedelegation(delAddrs[1], valAddrs[1], valAddrs[0], 0,
				time.Unix(0, 0), math.NewInt(5),
				math.LegacyNewDec(5), 0, address.NewBech32Codec("cosmosvaloper"), address.NewBech32Codec("cosmos"))

			s.Require().NoError(keeper.InsertRedelegationQueue(ctx, red1, t))

			iterator1, err := keeper.RedelegationQueueIterator(ctx, blockTime)
			s.Require().NoError(err)
			defer iterator1.Close()
			count1 := 0
			for ; iterator1.Valid(); iterator1.Next() {
				count1++
			}

			// redelegation should be retrieved
			// count 1 due to same redelegation time
			s.Require().Equal(1, count1)

			// Verify GetRedelegationQueueTimeSlice returns the correct redelegations after insertion
			reds, err := keeper.GetRedelegationQueueTimeSlice(ctx, blockTime)
			s.Require().NoError(err)
			s.Require().Equal(2, len(reds), "should have 2 redelegations at same time")

			// insert another redelegation with different redelegation time and height
			red2 := stakingtypes.NewRedelegation(delAddrs[2], valAddrs[2], valAddrs[0], 0,
				time.Unix(0, 0), math.NewInt(5),
				math.LegacyNewDec(5), 0, address.NewBech32Codec("cosmosvaloper"), address.NewBech32Codec("cosmos"))
			t2 := blockTime.Add(-1 * time.Minute)
			s.Require().NoError(keeper.InsertRedelegationQueue(ctx, red2, t2))

			iterator2, err := keeper.RedelegationQueueIterator(ctx, blockTime)
			s.Require().NoError(err)
			defer iterator2.Close()
			count2 := 0
			for ; iterator2.Valid(); iterator2.Next() {
				count2++
			}

			// redelegation should be retrieved
			s.Require().Equal(2, count2)

			// Verify the new redelegation was inserted at the correct time
			reds2, err := keeper.GetRedelegationQueueTimeSlice(ctx, t2)
			s.Require().NoError(err)
			s.Require().Equal(1, len(reds2), "should have 1 redelegation at different time")
		})
	}
}

func (s *KeeperTestSuite) TestDequeueAllMatureRedelegationQueue() {
	testCases := []struct {
		name             string
		maxCacheSize     int
		numRedelegations int
	}{
		{
			name:             "cache size < 0 (cache disabled)",
			maxCacheSize:     -1,
			numRedelegations: 3,
		},
		{
			name:             "cache size = 0 (unlimited cache)",
			maxCacheSize:     0,
			numRedelegations: 3,
		},
		{
			name:             "cache size > redelegations",
			maxCacheSize:     5,
			numRedelegations: 2,
		},
		{
			name:             "cache size == redelegations",
			maxCacheSize:     2,
			numRedelegations: 2,
		},
		{
			name:             "cache size < redelegations",
			maxCacheSize:     1,
			numRedelegations: 3,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			key := storetypes.NewKVStoreKey(stakingtypes.StoreKey)
			storeService := runtime.NewKVStoreService(key)
			testCtx := sdktestutil.DefaultContextWithDB(s.T(), key, storetypes.NewTransientStoreKey("transient_test"))
			ctx := testCtx.Ctx.WithBlockHeader(cmtproto.Header{Time: cmttime.Now()})
			encCfg := moduletestutil.MakeTestEncodingConfig()

			ctrl := gomock.NewController(s.T())
			accountKeeper := testutil.NewMockAccountKeeper(ctrl)
			accountKeeper.EXPECT().GetModuleAddress(stakingtypes.BondedPoolName).Return(bondedAcc.GetAddress()).AnyTimes()
			accountKeeper.EXPECT().GetModuleAddress(stakingtypes.NotBondedPoolName).Return(notBondedAcc.GetAddress()).AnyTimes()
			accountKeeper.EXPECT().AddressCodec().Return(address.NewBech32Codec("cosmos")).AnyTimes()

			bankKeeper := testutil.NewMockBankKeeper(ctrl)
			bankKeeper.EXPECT().DelegateCoinsFromAccountToModule(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
			bankKeeper.EXPECT().SendCoinsFromModuleToModule(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()

			// Initialize keeper with specific cache size
			keeper := stakingkeeper.NewKeeper(
				encCfg.Codec,
				storeService,
				accountKeeper,
				bankKeeper,
				authtypes.NewModuleAddress(govtypes.ModuleName).String(),
				address.NewBech32Codec("cosmosvaloper"),
				address.NewBech32Codec("cosmosvalcons"),
				tc.maxCacheSize,
			)
			params := stakingtypes.DefaultParams()
			params.UnbondingTime = 1 * time.Second // Short unbonding time for testing
			s.Require().NoError(keeper.SetParams(ctx, params))

			blockTime := time.Now().UTC()
			ctx = ctx.WithBlockTime(blockTime)

			// Create 2 validators
			valAddr1 := sdk.ValAddress(PKs[0].Address())
			validator1 := testutil.NewValidator(s.T(), valAddr1, PKs[0])
			validator1, _ = validator1.AddTokensFromDel(keeper.TokensFromConsensusPower(ctx, 100))
			validator1 = stakingkeeper.TestingUpdateValidator(keeper, ctx, validator1, true)

			valAddr2 := sdk.ValAddress(PKs[1].Address())
			validator2 := testutil.NewValidator(s.T(), valAddr2, PKs[1])
			validator2, _ = validator2.AddTokensFromDel(keeper.TokensFromConsensusPower(ctx, 100))
			stakingkeeper.TestingUpdateValidator(keeper, ctx, validator2, true)

			// Create multiple redelegations
			delAddrs, _ := createValAddrs(tc.numRedelegations)
			for i := 0; i < tc.numRedelegations; i++ {
				// Delegate to validator1
				bondAmt := keeper.TokensFromConsensusPower(ctx, 10)
				_, err := keeper.Delegate(ctx, delAddrs[i], bondAmt, stakingtypes.Unbonded, validator1, true)
				s.Require().NoError(err)

				// Redelegate from validator1 to validator2
				_, err = keeper.BeginRedelegation(ctx, delAddrs[i], valAddr1, valAddr2, math.LegacyNewDec(5))
				s.Require().NoError(err)
			}

			// Verify redelegations were created
			for i := 0; i < tc.numRedelegations; i++ {
				_, err := keeper.GetRedelegation(ctx, delAddrs[i], valAddr1, valAddr2)
				s.Require().NoError(err)
			}

			// Fast-forward time to maturity
			ctx = ctx.WithBlockTime(blockTime.Add(params.UnbondingTime))

			// Verify GetPendingRedelegations returns the expected number of redelegations
			allReds, err := keeper.GetPendingRedelegations(ctx, ctx.BlockTime())
			s.Require().NoError(err)
			s.Require().NotEmpty(allReds, "should have pending redelegations")

			// Verify GetRedelegationQueueTimeSlice returns the expected number of redelegations
			// In this case, it should return all redelegations as all redelegations mature at the same time.
			reds, err := keeper.GetRedelegationQueueTimeSlice(ctx, ctx.BlockTime())
			s.Require().NoError(err)
			s.Require().Equal(tc.numRedelegations, len(reds))

			// Dequeue and complete all mature redelegations
			matureRedelegations, err := keeper.DequeueAllMatureRedelegationQueue(ctx, ctx.BlockTime())
			s.Require().NoError(err)
			s.Require().Equal(tc.numRedelegations, len(matureRedelegations), "all redelegations should be mature")

			// Complete the redelegations
			for _, dvvTriplet := range matureRedelegations {
				delAddr, err := accountKeeper.AddressCodec().StringToBytes(dvvTriplet.DelegatorAddress)
				s.Require().NoError(err)
				valSrcAddr, err := keeper.ValidatorAddressCodec().StringToBytes(dvvTriplet.ValidatorSrcAddress)
				s.Require().NoError(err)
				valDstAddr, err := keeper.ValidatorAddressCodec().StringToBytes(dvvTriplet.ValidatorDstAddress)
				s.Require().NoError(err)
				_, err = keeper.CompleteRedelegation(ctx, delAddr, valSrcAddr, valDstAddr)
				s.Require().NoError(err)
			}

			// Verify all redelegations were completed (removed from store)
			for i := 0; i < tc.numRedelegations; i++ {
				_, err := keeper.GetRedelegation(ctx, delAddrs[i], valAddr1, valAddr2)
				s.Require().ErrorIs(err, stakingtypes.ErrNoRedelegation, "redelegation should be completed and removed")
			}
		})
	}
}

func (s *KeeperTestSuite) TestRedelegationQueueCacheRecovery() {
	// This test verifies that when the cache is initially too small (exceeded),
	// and then entries are dequeued, the cache can recover and be used again
	// Cache size is based on the number of unique timestamps (keys), not individual entries
	key := storetypes.NewKVStoreKey(stakingtypes.StoreKey)
	storeService := runtime.NewKVStoreService(key)
	testCtx := sdktestutil.DefaultContextWithDB(s.T(), key, storetypes.NewTransientStoreKey("transient_test"))
	ctx := testCtx.Ctx.WithBlockHeader(cmtproto.Header{Time: cmttime.Now()})
	encCfg := moduletestutil.MakeTestEncodingConfig()

	ctrl := gomock.NewController(s.T())
	accountKeeper := testutil.NewMockAccountKeeper(ctrl)
	accountKeeper.EXPECT().GetModuleAddress(stakingtypes.BondedPoolName).Return(bondedAcc.GetAddress()).AnyTimes()
	accountKeeper.EXPECT().GetModuleAddress(stakingtypes.NotBondedPoolName).Return(notBondedAcc.GetAddress()).AnyTimes()
	accountKeeper.EXPECT().AddressCodec().Return(address.NewBech32Codec("cosmos")).AnyTimes()

	bankKeeper := testutil.NewMockBankKeeper(ctrl)
	bankKeeper.EXPECT().DelegateCoinsFromAccountToModule(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
	bankKeeper.EXPECT().SendCoinsFromModuleToModule(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()

	// Initialize keeper with small cache size (2 timestamps)
	maxCacheSize := 2
	keeper := stakingkeeper.NewKeeper(
		encCfg.Codec,
		storeService,
		accountKeeper,
		bankKeeper,
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
		address.NewBech32Codec("cosmosvaloper"),
		address.NewBech32Codec("cosmosvalcons"),
		maxCacheSize,
	)
	params := stakingtypes.DefaultParams()
	params.UnbondingTime = 1 * time.Hour // Use 1 hour so we can create different timestamps
	s.Require().NoError(keeper.SetParams(ctx, params))

	baseTime := time.Now().UTC()
	ctx = ctx.WithBlockTime(baseTime)

	// Create 2 validators
	valAddr1 := sdk.ValAddress(PKs[0].Address())
	validator1 := testutil.NewValidator(s.T(), valAddr1, PKs[0])
	validator1, _ = validator1.AddTokensFromDel(keeper.TokensFromConsensusPower(ctx, 100))
	validator1 = stakingkeeper.TestingUpdateValidator(keeper, ctx, validator1, true)

	valAddr2 := sdk.ValAddress(PKs[1].Address())
	validator2 := testutil.NewValidator(s.T(), valAddr2, PKs[1])
	validator2, _ = validator2.AddTokensFromDel(keeper.TokensFromConsensusPower(ctx, 100))
	stakingkeeper.TestingUpdateValidator(keeper, ctx, validator2, true)

	// Create redelegations at 5 different timestamps (exceeds cache size of 2)
	numTimestamps := 5
	delAddrs, _ := createValAddrs(numTimestamps)
	completionTimes := make([]time.Time, numTimestamps)

	for i := 0; i < numTimestamps; i++ {
		// Set different block times to create different completion timestamps
		currentTime := baseTime.Add(time.Duration(i) * time.Second)
		ctx = ctx.WithBlockTime(currentTime)
		completionTimes[i] = currentTime.Add(params.UnbondingTime)

		// Delegate to validator1
		bondAmt := keeper.TokensFromConsensusPower(ctx, 10)
		_, err := keeper.Delegate(ctx, delAddrs[i], bondAmt, stakingtypes.Unbonded, validator1, true)
		s.Require().NoError(err)

		// Redelegate from validator1 to validator2 (will create redelegation with unique completion time)
		_, err = keeper.BeginRedelegation(ctx, delAddrs[i], valAddr1, valAddr2, math.LegacyNewDec(5))
		s.Require().NoError(err)
	}

	// Verify all redelegations were created
	for i := 0; i < numTimestamps; i++ {
		_, err := keeper.GetRedelegation(ctx, delAddrs[i], valAddr1, valAddr2)
		s.Require().NoError(err)
	}

	// At this point, cache should be exceeded (5 timestamps > maxCacheSize of 2)
	// GetPendingRedelegations should still work, but will read from store
	ctx = ctx.WithBlockTime(completionTimes[numTimestamps-1])
	allReds, err := keeper.GetPendingRedelegations(ctx, ctx.BlockTime())
	s.Require().NoError(err)
	s.Require().Equal(5, len(allReds), "should have 5 different timestamps")

	// Fast-forward time to mature the first 3 timestamps and dequeue them
	ctx = ctx.WithBlockTime(completionTimes[2])
	matureRedelegations, err := keeper.DequeueAllMatureRedelegationQueue(ctx, ctx.BlockTime())
	s.Require().NoError(err)
	s.Require().Equal(3, len(matureRedelegations), "should dequeue 3 redelegations from 3 timestamps")

	// Complete the first 3 redelegations
	for i := 0; i < 3; i++ {
		_, err = keeper.CompleteRedelegation(ctx, delAddrs[i], valAddr1, valAddr2)
		s.Require().NoError(err)
	}

	// Verify the first 3 were removed
	for i := 0; i < 3; i++ {
		_, err := keeper.GetRedelegation(ctx, delAddrs[i], valAddr1, valAddr2)
		s.Require().ErrorIs(err, stakingtypes.ErrNoRedelegation, "redelegation should be completed and removed")
	}

	// Verify the last 2 still exist
	for i := 3; i < numTimestamps; i++ {
		_, err := keeper.GetRedelegation(ctx, delAddrs[i], valAddr1, valAddr2)
		s.Require().NoError(err, "redelegation should still exist")
	}

	// Now only 2 timestamps remain (completionTimes[3] and completionTimes[4])
	// This fits in the cache (2 timestamps == maxCacheSize)
	// GetPendingRedelegations should now be able to use the cache
	ctx = ctx.WithBlockTime(completionTimes[4])
	remainingReds, err := keeper.GetPendingRedelegations(ctx, ctx.BlockTime())
	s.Require().NoError(err)
	s.Require().Equal(2, len(remainingReds), "should have 2 timestamps in cache")

	// Dequeue the remaining 2
	finalMatureRedelegations, err := keeper.DequeueAllMatureRedelegationQueue(ctx, ctx.BlockTime())
	s.Require().NoError(err)
	s.Require().Equal(2, len(finalMatureRedelegations), "should have 2 mature redelegations")

	// Complete them
	for i := 3; i < numTimestamps; i++ {
		_, err = keeper.CompleteRedelegation(ctx, delAddrs[i], valAddr1, valAddr2)
		s.Require().NoError(err)
	}

	// Verify all redelegations are now completed
	for i := 3; i < numTimestamps; i++ {
		_, err := keeper.GetRedelegation(ctx, delAddrs[i], valAddr1, valAddr2)
		s.Require().ErrorIs(err, stakingtypes.ErrNoRedelegation, "all redelegations should be completed")
	}
}

func (s *KeeperTestSuite) TestGetAndParseRedelegationTimeKey() {
	require := s.Require()

	blockTime := time.Now().UTC()
	key := stakingtypes.GetRedelegationTimeKey(blockTime)
	time, err := stakingtypes.ParseRedelegationTimeKey(key)
	require.NoError(err)
	require.Equal(blockTime, time)
}

func (s *KeeperTestSuite) TestSortRedelegationQueueKeysByAscendingOrder() {
	require := s.Require()

	currentTime := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)
	oneHourLater := currentTime.Add(1 * time.Hour)
	oneHourBefore := currentTime.Add(-1 * time.Hour)

	keys := []string{
		sdk.FormatTimeString(oneHourLater),
		sdk.FormatTimeString(oneHourBefore),
		sdk.FormatTimeString(currentTime),
	}

	stakingtypes.SortTimestampsByAscendingOrder(keys)

	// Verify sorting is correct - should be sorted by timestamp ascending order
	for i := 0; i < len(keys)-1; i++ {
		t1, err := sdk.ParseTime(keys[i])
		require.NoError(err)
		t2, err := sdk.ParseTime(keys[i+1])
		require.NoError(err)

		// Current entry should be before or equal to next entry
		require.True(t1.Before(t2) || t1.Equal(t2), "timestamps should be in ascending order")

	}

	firstTime, err := sdk.ParseTime(keys[0])
	require.NoError(err)
	require.Equal(oneHourBefore, firstTime)

	lastTime, err := sdk.ParseTime(keys[len(keys)-1])
	require.NoError(err)
	require.Equal(oneHourLater, lastTime)
}
