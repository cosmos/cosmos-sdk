package keeper

import (
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// tests GetDelegation, GetDelegatorDelegations, SetDelegation, RemoveDelegation, GetDelegatorDelegations
func TestDelegation(t *testing.T) {
	ctx, _, keeper := CreateTestInput(t, false, 10)
	pool := keeper.GetPool(ctx)

	//construct the validators
	amts := []sdk.Int{sdk.NewInt(9), sdk.NewInt(8), sdk.NewInt(7)}
	var validators [3]types.Validator
	for i, amt := range amts {
		validators[i] = types.NewValidator(addrVals[i], PKs[i], types.Description{})
		validators[i], pool, _ = validators[i].AddTokensFromDel(pool, amt)
	}

	keeper.SetPool(ctx, pool)
	validators[0] = TestingUpdateValidator(keeper, ctx, validators[0], true)
	validators[1] = TestingUpdateValidator(keeper, ctx, validators[1], true)
	validators[2] = TestingUpdateValidator(keeper, ctx, validators[2], true)

	// first add a validators[0] to delegate too

	bond1to1 := types.NewDelegation(addrDels[0], addrVals[0], sdk.NewDec(9))

	// check the empty keeper first
	_, found := keeper.GetDelegation(ctx, addrDels[0], addrVals[0])
	require.False(t, found)

	// set and retrieve a record
	keeper.SetDelegation(ctx, bond1to1)
	resBond, found := keeper.GetDelegation(ctx, addrDels[0], addrVals[0])
	require.True(t, found)
	require.True(t, bond1to1.Equal(resBond))

	// modify a records, save, and retrieve
	bond1to1.Shares = sdk.NewDec(99)
	keeper.SetDelegation(ctx, bond1to1)
	resBond, found = keeper.GetDelegation(ctx, addrDels[0], addrVals[0])
	require.True(t, found)
	require.True(t, bond1to1.Equal(resBond))

	// add some more records
	bond1to2 := types.NewDelegation(addrDels[0], addrVals[1], sdk.NewDec(9))
	bond1to3 := types.NewDelegation(addrDels[0], addrVals[2], sdk.NewDec(9))
	bond2to1 := types.NewDelegation(addrDels[1], addrVals[0], sdk.NewDec(9))
	bond2to2 := types.NewDelegation(addrDels[1], addrVals[1], sdk.NewDec(9))
	bond2to3 := types.NewDelegation(addrDels[1], addrVals[2], sdk.NewDec(9))
	keeper.SetDelegation(ctx, bond1to2)
	keeper.SetDelegation(ctx, bond1to3)
	keeper.SetDelegation(ctx, bond2to1)
	keeper.SetDelegation(ctx, bond2to2)
	keeper.SetDelegation(ctx, bond2to3)

	// test all bond retrieve capabilities
	resBonds := keeper.GetDelegatorDelegations(ctx, addrDels[0], 5)
	require.Equal(t, 3, len(resBonds))
	require.True(t, bond1to1.Equal(resBonds[0]))
	require.True(t, bond1to2.Equal(resBonds[1]))
	require.True(t, bond1to3.Equal(resBonds[2]))
	resBonds = keeper.GetAllDelegatorDelegations(ctx, addrDels[0])
	require.Equal(t, 3, len(resBonds))
	resBonds = keeper.GetDelegatorDelegations(ctx, addrDels[0], 2)
	require.Equal(t, 2, len(resBonds))
	resBonds = keeper.GetDelegatorDelegations(ctx, addrDels[1], 5)
	require.Equal(t, 3, len(resBonds))
	require.True(t, bond2to1.Equal(resBonds[0]))
	require.True(t, bond2to2.Equal(resBonds[1]))
	require.True(t, bond2to3.Equal(resBonds[2]))
	allBonds := keeper.GetAllDelegations(ctx)
	require.Equal(t, 6, len(allBonds))
	require.True(t, bond1to1.Equal(allBonds[0]))
	require.True(t, bond1to2.Equal(allBonds[1]))
	require.True(t, bond1to3.Equal(allBonds[2]))
	require.True(t, bond2to1.Equal(allBonds[3]))
	require.True(t, bond2to2.Equal(allBonds[4]))
	require.True(t, bond2to3.Equal(allBonds[5]))

	resVals := keeper.GetDelegatorValidators(ctx, addrDels[0], 3)
	require.Equal(t, 3, len(resVals))
	resVals = keeper.GetDelegatorValidators(ctx, addrDels[1], 4)
	require.Equal(t, 3, len(resVals))

	for i := 0; i < 3; i++ {

		resVal, err := keeper.GetDelegatorValidator(ctx, addrDels[0], addrVals[i])
		require.Nil(t, err)
		require.Equal(t, addrVals[i], resVal.GetOperator())

		resVal, err = keeper.GetDelegatorValidator(ctx, addrDels[1], addrVals[i])
		require.Nil(t, err)
		require.Equal(t, addrVals[i], resVal.GetOperator())

		resDels := keeper.GetValidatorDelegations(ctx, addrVals[i])
		require.Len(t, resDels, 2)
	}

	// delete a record
	keeper.RemoveDelegation(ctx, bond2to3)
	_, found = keeper.GetDelegation(ctx, addrDels[1], addrVals[2])
	require.False(t, found)
	resBonds = keeper.GetDelegatorDelegations(ctx, addrDels[1], 5)
	require.Equal(t, 2, len(resBonds))
	require.True(t, bond2to1.Equal(resBonds[0]))
	require.True(t, bond2to2.Equal(resBonds[1]))

	resBonds = keeper.GetAllDelegatorDelegations(ctx, addrDels[1])
	require.Equal(t, 2, len(resBonds))

	// delete all the records from delegator 2
	keeper.RemoveDelegation(ctx, bond2to1)
	keeper.RemoveDelegation(ctx, bond2to2)
	_, found = keeper.GetDelegation(ctx, addrDels[1], addrVals[0])
	require.False(t, found)
	_, found = keeper.GetDelegation(ctx, addrDels[1], addrVals[1])
	require.False(t, found)
	resBonds = keeper.GetDelegatorDelegations(ctx, addrDels[1], 5)
	require.Equal(t, 0, len(resBonds))
}

// tests Get/Set/Remove UnbondingDelegation
func TestUnbondingDelegation(t *testing.T) {
	ctx, _, keeper := CreateTestInput(t, false, 0)

	ubd := types.NewUnbondingDelegation(addrDels[0], addrVals[0], 0,
		time.Unix(0, 0), sdk.NewInt(5))

	// set and retrieve a record
	keeper.SetUnbondingDelegation(ctx, ubd)
	resUnbond, found := keeper.GetUnbondingDelegation(ctx, addrDels[0], addrVals[0])
	require.True(t, found)
	require.True(t, ubd.Equal(resUnbond))

	// modify a records, save, and retrieve
	ubd.Entries[0].Balance = sdk.NewInt(21)
	keeper.SetUnbondingDelegation(ctx, ubd)

	resUnbonds := keeper.GetUnbondingDelegations(ctx, addrDels[0], 5)
	require.Equal(t, 1, len(resUnbonds))

	resUnbonds = keeper.GetAllUnbondingDelegations(ctx, addrDels[0])
	require.Equal(t, 1, len(resUnbonds))

	resUnbond, found = keeper.GetUnbondingDelegation(ctx, addrDels[0], addrVals[0])
	require.True(t, found)
	require.True(t, ubd.Equal(resUnbond))

	// delete a record
	keeper.RemoveUnbondingDelegation(ctx, ubd)
	_, found = keeper.GetUnbondingDelegation(ctx, addrDels[0], addrVals[0])
	require.False(t, found)

	resUnbonds = keeper.GetUnbondingDelegations(ctx, addrDels[0], 5)
	require.Equal(t, 0, len(resUnbonds))

	resUnbonds = keeper.GetAllUnbondingDelegations(ctx, addrDels[0])
	require.Equal(t, 0, len(resUnbonds))

}

func TestUnbondDelegation(t *testing.T) {
	ctx, _, keeper := CreateTestInput(t, false, 0)
	pool := keeper.GetPool(ctx)
	startTokens := sdk.TokensFromTendermintPower(10)
	pool.NotBondedTokens = startTokens

	//create a validator and a delegator to that validator
	validator := types.NewValidator(addrVals[0], PKs[0], types.Description{})
	validator, pool, issuedShares := validator.AddTokensFromDel(pool, startTokens)
	require.Equal(t, startTokens, issuedShares.RoundInt())
	keeper.SetPool(ctx, pool)
	validator = TestingUpdateValidator(keeper, ctx, validator, true)

	pool = keeper.GetPool(ctx)
	require.Equal(t, startTokens, pool.BondedTokens)
	require.Equal(t, startTokens, validator.BondedTokens())

	delegation := types.NewDelegation(addrDels[0], addrVals[0], issuedShares)
	keeper.SetDelegation(ctx, delegation)

	bondTokens := sdk.TokensFromTendermintPower(6)
	amount, err := keeper.unbond(ctx, addrDels[0], addrVals[0], sdk.NewDecFromInt(bondTokens))
	require.NoError(t, err)
	require.Equal(t, bondTokens, amount) // shares to be added to an unbonding delegation

	delegation, found := keeper.GetDelegation(ctx, addrDels[0], addrVals[0])
	require.True(t, found)
	validator, found = keeper.GetValidator(ctx, addrVals[0])
	require.True(t, found)
	pool = keeper.GetPool(ctx)

	remainingTokens := startTokens.Sub(bondTokens)
	require.Equal(t, remainingTokens, delegation.Shares.RoundInt())
	require.Equal(t, remainingTokens, validator.BondedTokens())
	require.Equal(t, bondTokens, pool.NotBondedTokens, "%v", pool)
	require.Equal(t, remainingTokens, pool.BondedTokens)
}

func TestUnbondingDelegationsMaxEntries(t *testing.T) {
	ctx, _, keeper := CreateTestInput(t, false, 0)
	pool := keeper.GetPool(ctx)
	startTokens := sdk.TokensFromTendermintPower(10)
	pool.NotBondedTokens = startTokens

	// create a validator and a delegator to that validator
	validator := types.NewValidator(addrVals[0], PKs[0], types.Description{})
	validator, pool, issuedShares := validator.AddTokensFromDel(pool, startTokens)
	require.Equal(t, startTokens, issuedShares.RoundInt())
	keeper.SetPool(ctx, pool)
	validator = TestingUpdateValidator(keeper, ctx, validator, true)

	pool = keeper.GetPool(ctx)
	require.Equal(t, startTokens, pool.BondedTokens)
	require.Equal(t, startTokens, validator.BondedTokens())

	delegation := types.NewDelegation(addrDels[0], addrVals[0], issuedShares)
	keeper.SetDelegation(ctx, delegation)

	maxEntries := keeper.MaxEntries(ctx)

	// should all pass
	var completionTime time.Time
	for i := uint16(0); i < maxEntries; i++ {
		var err error
		completionTime, err = keeper.Undelegate(ctx, addrDels[0], addrVals[0], sdk.NewDec(1))
		require.NoError(t, err)
	}

	// an additional unbond should fail due to max entries
	_, err := keeper.Undelegate(ctx, addrDels[0], addrVals[0], sdk.NewDec(1))
	require.Error(t, err)

	// mature unbonding delegations
	ctx = ctx.WithBlockTime(completionTime)
	err = keeper.CompleteUnbonding(ctx, addrDels[0], addrVals[0])
	require.NoError(t, err)

	// unbonding  should work again
	_, err = keeper.Undelegate(ctx, addrDels[0], addrVals[0], sdk.NewDec(1))
	require.NoError(t, err)
}

// test undelegating self delegation from a validator pushing it below MinSelfDelegation
// shift it from the bonded to unbonding state and jailed
func TestUndelegateSelfDelegationBelowMinSelfDelegation(t *testing.T) {

	ctx, _, keeper := CreateTestInput(t, false, 0)
	pool := keeper.GetPool(ctx)
	startTokens := sdk.TokensFromTendermintPower(20)
	pool.NotBondedTokens = startTokens

	//create a validator with a self-delegation
	validator := types.NewValidator(addrVals[0], PKs[0], types.Description{})

	valTokens := sdk.TokensFromTendermintPower(10)
	validator.MinSelfDelegation = valTokens
	validator, pool, issuedShares := validator.AddTokensFromDel(pool, valTokens)
	require.Equal(t, valTokens, issuedShares.RoundInt())

	keeper.SetPool(ctx, pool)
	validator = TestingUpdateValidator(keeper, ctx, validator, true)
	pool = keeper.GetPool(ctx)
	selfDelegation := types.NewDelegation(sdk.AccAddress(addrVals[0].Bytes()), addrVals[0], issuedShares)
	keeper.SetDelegation(ctx, selfDelegation)

	// create a second delegation to this validator
	keeper.DeleteValidatorByPowerIndex(ctx, validator)
	delTokens := sdk.TokensFromTendermintPower(10)
	validator, pool, issuedShares = validator.AddTokensFromDel(pool, delTokens)
	require.Equal(t, delTokens, issuedShares.RoundInt())
	keeper.SetPool(ctx, pool)
	validator = TestingUpdateValidator(keeper, ctx, validator, true)
	pool = keeper.GetPool(ctx)
	delegation := types.NewDelegation(addrDels[0], addrVals[0], issuedShares)
	keeper.SetDelegation(ctx, delegation)

	val0AccAddr := sdk.AccAddress(addrVals[0].Bytes())
	_, err := keeper.Undelegate(ctx, val0AccAddr, addrVals[0], sdk.NewDecFromInt(sdk.TokensFromTendermintPower(6)))
	require.NoError(t, err)

	// end block
	updates := keeper.ApplyAndReturnValidatorSetUpdates(ctx)
	require.Equal(t, 1, len(updates))

	validator, found := keeper.GetValidator(ctx, addrVals[0])
	require.True(t, found)
	require.Equal(t, sdk.TokensFromTendermintPower(14), validator.Tokens)
	require.Equal(t, sdk.Unbonding, validator.Status)
	require.True(t, validator.Jailed)
}

func TestUndelegateFromUnbondingValidator(t *testing.T) {
	ctx, _, keeper := CreateTestInput(t, false, 0)
	pool := keeper.GetPool(ctx)
	startTokens := sdk.TokensFromTendermintPower(20)
	pool.NotBondedTokens = startTokens

	//create a validator with a self-delegation
	validator := types.NewValidator(addrVals[0], PKs[0], types.Description{})

	valTokens := sdk.TokensFromTendermintPower(10)
	validator, pool, issuedShares := validator.AddTokensFromDel(pool, valTokens)
	require.Equal(t, valTokens, issuedShares.RoundInt())
	keeper.SetPool(ctx, pool)
	validator = TestingUpdateValidator(keeper, ctx, validator, true)
	pool = keeper.GetPool(ctx)
	selfDelegation := types.NewDelegation(sdk.AccAddress(addrVals[0].Bytes()), addrVals[0], issuedShares)
	keeper.SetDelegation(ctx, selfDelegation)

	// create a second delegation to this validator
	keeper.DeleteValidatorByPowerIndex(ctx, validator)
	delTokens := sdk.TokensFromTendermintPower(10)
	validator, pool, issuedShares = validator.AddTokensFromDel(pool, delTokens)
	require.Equal(t, delTokens, issuedShares.RoundInt())
	keeper.SetPool(ctx, pool)
	validator = TestingUpdateValidator(keeper, ctx, validator, true)
	pool = keeper.GetPool(ctx)
	delegation := types.NewDelegation(addrDels[0], addrVals[0], issuedShares)
	keeper.SetDelegation(ctx, delegation)

	header := ctx.BlockHeader()
	blockHeight := int64(10)
	header.Height = blockHeight
	blockTime := time.Unix(333, 0)
	header.Time = blockTime
	ctx = ctx.WithBlockHeader(header)

	// unbond the all self-delegation to put validator in unbonding state
	val0AccAddr := sdk.AccAddress(addrVals[0].Bytes())
	_, err := keeper.Undelegate(ctx, val0AccAddr, addrVals[0], sdk.NewDecFromInt(valTokens))
	require.NoError(t, err)

	// end block
	updates := keeper.ApplyAndReturnValidatorSetUpdates(ctx)
	require.Equal(t, 1, len(updates))

	validator, found := keeper.GetValidator(ctx, addrVals[0])
	require.True(t, found)
	require.Equal(t, blockHeight, validator.UnbondingHeight)
	params := keeper.GetParams(ctx)
	require.True(t, blockTime.Add(params.UnbondingTime).Equal(validator.UnbondingCompletionTime))

	//change the context
	header = ctx.BlockHeader()
	blockHeight2 := int64(20)
	header.Height = blockHeight2
	blockTime2 := time.Unix(444, 0)
	header.Time = blockTime2
	ctx = ctx.WithBlockHeader(header)

	// unbond some of the other delegation's shares
	_, err = keeper.Undelegate(ctx, addrDels[0], addrVals[0], sdk.NewDec(6))
	require.NoError(t, err)

	// retrieve the unbonding delegation
	ubd, found := keeper.GetUnbondingDelegation(ctx, addrDels[0], addrVals[0])
	require.True(t, found)
	require.Len(t, ubd.Entries, 1)
	require.True(t, ubd.Entries[0].Balance.Equal(sdk.NewInt(6)))
	assert.Equal(t, blockHeight, ubd.Entries[0].CreationHeight)
	assert.True(t, blockTime.Add(params.UnbondingTime).Equal(ubd.Entries[0].CompletionTime))
}

func TestUndelegateFromUnbondedValidator(t *testing.T) {
	ctx, _, keeper := CreateTestInput(t, false, 0)
	pool := keeper.GetPool(ctx)
	startTokens := sdk.TokensFromTendermintPower(20)
	pool.NotBondedTokens = startTokens

	//create a validator with a self-delegation
	validator := types.NewValidator(addrVals[0], PKs[0], types.Description{})

	valTokens := sdk.TokensFromTendermintPower(10)
	validator, pool, issuedShares := validator.AddTokensFromDel(pool, valTokens)
	require.Equal(t, valTokens, issuedShares.RoundInt())
	keeper.SetPool(ctx, pool)
	validator = TestingUpdateValidator(keeper, ctx, validator, true)
	pool = keeper.GetPool(ctx)
	val0AccAddr := sdk.AccAddress(addrVals[0].Bytes())
	selfDelegation := types.NewDelegation(val0AccAddr, addrVals[0], issuedShares)
	keeper.SetDelegation(ctx, selfDelegation)

	// create a second delegation to this validator
	keeper.DeleteValidatorByPowerIndex(ctx, validator)
	delTokens := sdk.TokensFromTendermintPower(10)
	validator, pool, issuedShares = validator.AddTokensFromDel(pool, delTokens)
	require.Equal(t, delTokens, issuedShares.RoundInt())
	keeper.SetPool(ctx, pool)
	validator = TestingUpdateValidator(keeper, ctx, validator, true)
	pool = keeper.GetPool(ctx)
	delegation := types.NewDelegation(addrDels[0], addrVals[0], issuedShares)
	keeper.SetDelegation(ctx, delegation)

	ctx = ctx.WithBlockHeight(10)
	ctx = ctx.WithBlockTime(time.Unix(333, 0))

	// unbond the all self-delegation to put validator in unbonding state
	_, err := keeper.Undelegate(ctx, val0AccAddr, addrVals[0], sdk.NewDecFromInt(valTokens))
	require.NoError(t, err)

	// end block
	updates := keeper.ApplyAndReturnValidatorSetUpdates(ctx)
	require.Equal(t, 1, len(updates))

	validator, found := keeper.GetValidator(ctx, addrVals[0])
	require.True(t, found)
	require.Equal(t, ctx.BlockHeight(), validator.UnbondingHeight)
	params := keeper.GetParams(ctx)
	require.True(t, ctx.BlockHeader().Time.Add(params.UnbondingTime).Equal(validator.UnbondingCompletionTime))

	// unbond the validator
	ctx = ctx.WithBlockTime(validator.UnbondingCompletionTime)
	keeper.UnbondAllMatureValidatorQueue(ctx)

	// Make sure validator is still in state because there is still an outstanding delegation
	validator, found = keeper.GetValidator(ctx, addrVals[0])
	require.True(t, found)
	require.Equal(t, validator.Status, sdk.Unbonded)

	// unbond some of the other delegation's shares
	unbondTokens := sdk.TokensFromTendermintPower(6)
	_, err = keeper.Undelegate(ctx, addrDels[0], addrVals[0], sdk.NewDecFromInt(unbondTokens))
	require.NoError(t, err)

	// no ubd should have been found, coins should have been returned direcly to account
	ubd, found := keeper.GetUnbondingDelegation(ctx, addrDels[0], addrVals[0])
	require.False(t, found, "%v", ubd)

	// unbond rest of the other delegation's shares
	remainingTokens := delTokens.Sub(unbondTokens)
	_, err = keeper.Undelegate(ctx, addrDels[0], addrVals[0], sdk.NewDecFromInt(remainingTokens))
	require.NoError(t, err)

	//  now validator should now be deleted from state
	validator, found = keeper.GetValidator(ctx, addrVals[0])
	require.False(t, found, "%v", validator)
}

func TestUnbondingAllDelegationFromValidator(t *testing.T) {
	ctx, _, keeper := CreateTestInput(t, false, 0)
	pool := keeper.GetPool(ctx)
	startTokens := sdk.TokensFromTendermintPower(20)
	pool.NotBondedTokens = startTokens

	//create a validator with a self-delegation
	validator := types.NewValidator(addrVals[0], PKs[0], types.Description{})

	valTokens := sdk.TokensFromTendermintPower(10)
	validator, pool, issuedShares := validator.AddTokensFromDel(pool, valTokens)
	require.Equal(t, valTokens, issuedShares.RoundInt())
	keeper.SetPool(ctx, pool)
	validator = TestingUpdateValidator(keeper, ctx, validator, true)
	pool = keeper.GetPool(ctx)
	val0AccAddr := sdk.AccAddress(addrVals[0].Bytes())
	selfDelegation := types.NewDelegation(val0AccAddr, addrVals[0], issuedShares)
	keeper.SetDelegation(ctx, selfDelegation)

	// create a second delegation to this validator
	keeper.DeleteValidatorByPowerIndex(ctx, validator)
	delTokens := sdk.TokensFromTendermintPower(10)
	validator, pool, issuedShares = validator.AddTokensFromDel(pool, delTokens)
	require.Equal(t, delTokens, issuedShares.RoundInt())
	keeper.SetPool(ctx, pool)
	validator = TestingUpdateValidator(keeper, ctx, validator, true)
	pool = keeper.GetPool(ctx)
	delegation := types.NewDelegation(addrDels[0], addrVals[0], issuedShares)
	keeper.SetDelegation(ctx, delegation)

	ctx = ctx.WithBlockHeight(10)
	ctx = ctx.WithBlockTime(time.Unix(333, 0))

	// unbond the all self-delegation to put validator in unbonding state
	_, err := keeper.Undelegate(ctx, val0AccAddr, addrVals[0], sdk.NewDecFromInt(valTokens))
	require.NoError(t, err)

	// end block
	updates := keeper.ApplyAndReturnValidatorSetUpdates(ctx)
	require.Equal(t, 1, len(updates))

	// unbond all the remaining delegation
	_, err = keeper.Undelegate(ctx, addrDels[0], addrVals[0], sdk.NewDecFromInt(delTokens))
	require.NoError(t, err)

	// validator should still be in state and still be in unbonding state
	validator, found := keeper.GetValidator(ctx, addrVals[0])
	require.True(t, found)
	require.Equal(t, validator.Status, sdk.Unbonding)

	// unbond the validator
	ctx = ctx.WithBlockTime(validator.UnbondingCompletionTime)
	keeper.UnbondAllMatureValidatorQueue(ctx)

	// validator should now be deleted from state
	_, found = keeper.GetValidator(ctx, addrVals[0])
	require.False(t, found)
}

// Make sure that that the retrieving the delegations doesn't affect the state
func TestGetRedelegationsFromValidator(t *testing.T) {
	ctx, _, keeper := CreateTestInput(t, false, 0)

	rd := types.NewRedelegation(addrDels[0], addrVals[0], addrVals[1], 0,
		time.Unix(0, 0), sdk.NewInt(5),
		sdk.NewDec(5))

	// set and retrieve a record
	keeper.SetRedelegation(ctx, rd)
	resBond, found := keeper.GetRedelegation(ctx, addrDels[0], addrVals[0], addrVals[1])
	require.True(t, found)

	// get the redelegations one time
	redelegations := keeper.GetRedelegationsFromValidator(ctx, addrVals[0])
	require.Equal(t, 1, len(redelegations))
	require.True(t, redelegations[0].Equal(resBond))

	// get the redelegations a second time, should be exactly the same
	redelegations = keeper.GetRedelegationsFromValidator(ctx, addrVals[0])
	require.Equal(t, 1, len(redelegations))
	require.True(t, redelegations[0].Equal(resBond))
}

// tests Get/Set/Remove/Has UnbondingDelegation
func TestRedelegation(t *testing.T) {
	ctx, _, keeper := CreateTestInput(t, false, 0)

	rd := types.NewRedelegation(addrDels[0], addrVals[0], addrVals[1], 0,
		time.Unix(0, 0), sdk.NewInt(5),
		sdk.NewDec(5))

	// test shouldn't have and redelegations
	has := keeper.HasReceivingRedelegation(ctx, addrDels[0], addrVals[1])
	require.False(t, has)

	// set and retrieve a record
	keeper.SetRedelegation(ctx, rd)
	resRed, found := keeper.GetRedelegation(ctx, addrDels[0], addrVals[0], addrVals[1])
	require.True(t, found)

	redelegations := keeper.GetRedelegationsFromValidator(ctx, addrVals[0])
	require.Equal(t, 1, len(redelegations))
	require.True(t, redelegations[0].Equal(resRed))

	redelegations = keeper.GetRedelegations(ctx, addrDels[0], 5)
	require.Equal(t, 1, len(redelegations))
	require.True(t, redelegations[0].Equal(resRed))

	redelegations = keeper.GetAllRedelegations(ctx, addrDels[0], nil, nil)
	require.Equal(t, 1, len(redelegations))
	require.True(t, redelegations[0].Equal(resRed))

	// check if has the redelegation
	has = keeper.HasReceivingRedelegation(ctx, addrDels[0], addrVals[1])
	require.True(t, has)

	// modify a records, save, and retrieve
	rd.Entries[0].SharesDst = sdk.NewDec(21)
	keeper.SetRedelegation(ctx, rd)

	resRed, found = keeper.GetRedelegation(ctx, addrDels[0], addrVals[0], addrVals[1])
	require.True(t, found)
	require.True(t, rd.Equal(resRed))

	redelegations = keeper.GetRedelegationsFromValidator(ctx, addrVals[0])
	require.Equal(t, 1, len(redelegations))
	require.True(t, redelegations[0].Equal(resRed))

	redelegations = keeper.GetRedelegations(ctx, addrDels[0], 5)
	require.Equal(t, 1, len(redelegations))
	require.True(t, redelegations[0].Equal(resRed))

	// delete a record
	keeper.RemoveRedelegation(ctx, rd)
	_, found = keeper.GetRedelegation(ctx, addrDels[0], addrVals[0], addrVals[1])
	require.False(t, found)

	redelegations = keeper.GetRedelegations(ctx, addrDels[0], 5)
	require.Equal(t, 0, len(redelegations))

	redelegations = keeper.GetAllRedelegations(ctx, addrDels[0], nil, nil)
	require.Equal(t, 0, len(redelegations))
}

func TestRedelegateToSameValidator(t *testing.T) {
	ctx, _, keeper := CreateTestInput(t, false, 0)
	pool := keeper.GetPool(ctx)
	startTokens := sdk.TokensFromTendermintPower(30)
	pool.NotBondedTokens = startTokens

	// create a validator with a self-delegation
	validator := types.NewValidator(addrVals[0], PKs[0], types.Description{})
	valTokens := sdk.TokensFromTendermintPower(10)
	validator, pool, issuedShares := validator.AddTokensFromDel(pool, valTokens)
	require.Equal(t, valTokens, issuedShares.RoundInt())
	keeper.SetPool(ctx, pool)
	validator = TestingUpdateValidator(keeper, ctx, validator, true)
	pool = keeper.GetPool(ctx)
	val0AccAddr := sdk.AccAddress(addrVals[0].Bytes())
	selfDelegation := types.NewDelegation(val0AccAddr, addrVals[0], issuedShares)
	keeper.SetDelegation(ctx, selfDelegation)

	_, err := keeper.BeginRedelegation(ctx, val0AccAddr, addrVals[0], addrVals[0], sdk.NewDec(5))
	require.Error(t, err)

}

func TestRedelegationMaxEntries(t *testing.T) {
	ctx, _, keeper := CreateTestInput(t, false, 0)
	pool := keeper.GetPool(ctx)
	startTokens := sdk.TokensFromTendermintPower(20)
	pool.NotBondedTokens = startTokens

	// create a validator with a self-delegation
	validator := types.NewValidator(addrVals[0], PKs[0], types.Description{})
	valTokens := sdk.TokensFromTendermintPower(10)
	validator, pool, issuedShares := validator.AddTokensFromDel(pool, valTokens)
	require.Equal(t, valTokens, issuedShares.RoundInt())
	keeper.SetPool(ctx, pool)
	validator = TestingUpdateValidator(keeper, ctx, validator, true)
	pool = keeper.GetPool(ctx)
	val0AccAddr := sdk.AccAddress(addrVals[0].Bytes())
	selfDelegation := types.NewDelegation(val0AccAddr, addrVals[0], issuedShares)
	keeper.SetDelegation(ctx, selfDelegation)

	// create a second validator
	validator2 := types.NewValidator(addrVals[1], PKs[1], types.Description{})
	validator2, pool, issuedShares = validator2.AddTokensFromDel(pool, valTokens)
	require.Equal(t, valTokens, issuedShares.RoundInt())
	pool.BondedTokens = pool.BondedTokens.Add(valTokens)
	keeper.SetPool(ctx, pool)
	validator2 = TestingUpdateValidator(keeper, ctx, validator2, true)
	require.Equal(t, sdk.Bonded, validator2.Status)

	maxEntries := keeper.MaxEntries(ctx)

	// redelegations should pass
	var completionTime time.Time
	for i := uint16(0); i < maxEntries; i++ {
		var err error
		completionTime, err = keeper.BeginRedelegation(ctx, val0AccAddr, addrVals[0], addrVals[1], sdk.NewDec(1))
		require.NoError(t, err)
	}

	// an additional redelegation should fail due to max entries
	_, err := keeper.BeginRedelegation(ctx, val0AccAddr, addrVals[0], addrVals[1], sdk.NewDec(1))
	require.Error(t, err)

	// mature redelegations
	ctx = ctx.WithBlockTime(completionTime)
	err = keeper.CompleteRedelegation(ctx, val0AccAddr, addrVals[0], addrVals[1])
	require.NoError(t, err)

	// redelegation should work again
	_, err = keeper.BeginRedelegation(ctx, val0AccAddr, addrVals[0], addrVals[1], sdk.NewDec(1))
	require.NoError(t, err)
}

func TestRedelegateSelfDelegation(t *testing.T) {
	ctx, _, keeper := CreateTestInput(t, false, 0)
	pool := keeper.GetPool(ctx)
	startTokens := sdk.TokensFromTendermintPower(30)
	pool.NotBondedTokens = startTokens

	//create a validator with a self-delegation
	validator := types.NewValidator(addrVals[0], PKs[0], types.Description{})
	valTokens := sdk.TokensFromTendermintPower(10)
	validator, pool, issuedShares := validator.AddTokensFromDel(pool, valTokens)
	require.Equal(t, valTokens, issuedShares.RoundInt())
	keeper.SetPool(ctx, pool)
	validator = TestingUpdateValidator(keeper, ctx, validator, true)
	pool = keeper.GetPool(ctx)
	val0AccAddr := sdk.AccAddress(addrVals[0].Bytes())
	selfDelegation := types.NewDelegation(val0AccAddr, addrVals[0], issuedShares)
	keeper.SetDelegation(ctx, selfDelegation)

	// create a second validator
	validator2 := types.NewValidator(addrVals[1], PKs[1], types.Description{})
	validator2, pool, issuedShares = validator2.AddTokensFromDel(pool, valTokens)
	require.Equal(t, valTokens, issuedShares.RoundInt())
	pool.BondedTokens = pool.BondedTokens.Add(valTokens)
	keeper.SetPool(ctx, pool)
	validator2 = TestingUpdateValidator(keeper, ctx, validator2, true)
	require.Equal(t, sdk.Bonded, validator2.Status)

	// create a second delegation to validator 1
	delTokens := sdk.TokensFromTendermintPower(10)
	validator, pool, issuedShares = validator.AddTokensFromDel(pool, delTokens)
	require.Equal(t, delTokens, issuedShares.RoundInt())
	keeper.SetPool(ctx, pool)
	validator = TestingUpdateValidator(keeper, ctx, validator, true)
	pool = keeper.GetPool(ctx)

	delegation := types.NewDelegation(addrDels[0], addrVals[0], issuedShares)
	keeper.SetDelegation(ctx, delegation)

	_, err := keeper.BeginRedelegation(ctx, val0AccAddr, addrVals[0], addrVals[1], sdk.NewDecFromInt(delTokens))
	require.NoError(t, err)

	// end block
	updates := keeper.ApplyAndReturnValidatorSetUpdates(ctx)
	require.Equal(t, 2, len(updates))

	validator, found := keeper.GetValidator(ctx, addrVals[0])
	require.True(t, found)
	require.Equal(t, valTokens, validator.Tokens)
	require.Equal(t, sdk.Unbonding, validator.Status)
}

func TestRedelegateFromUnbondingValidator(t *testing.T) {
	ctx, _, keeper := CreateTestInput(t, false, 0)
	pool := keeper.GetPool(ctx)
	startTokens := sdk.TokensFromTendermintPower(30)
	pool.NotBondedTokens = startTokens

	//create a validator with a self-delegation
	validator := types.NewValidator(addrVals[0], PKs[0], types.Description{})

	valTokens := sdk.TokensFromTendermintPower(10)
	validator, pool, issuedShares := validator.AddTokensFromDel(pool, valTokens)
	require.Equal(t, valTokens, issuedShares.RoundInt())
	keeper.SetPool(ctx, pool)
	validator = TestingUpdateValidator(keeper, ctx, validator, true)
	pool = keeper.GetPool(ctx)
	val0AccAddr := sdk.AccAddress(addrVals[0].Bytes())
	selfDelegation := types.NewDelegation(val0AccAddr, addrVals[0], issuedShares)
	keeper.SetDelegation(ctx, selfDelegation)

	// create a second delegation to this validator
	keeper.DeleteValidatorByPowerIndex(ctx, validator)
	delTokens := sdk.TokensFromTendermintPower(10)
	validator, pool, issuedShares = validator.AddTokensFromDel(pool, delTokens)
	require.Equal(t, delTokens, issuedShares.RoundInt())
	keeper.SetPool(ctx, pool)
	validator = TestingUpdateValidator(keeper, ctx, validator, true)
	pool = keeper.GetPool(ctx)
	delegation := types.NewDelegation(addrDels[0], addrVals[0], issuedShares)
	keeper.SetDelegation(ctx, delegation)

	// create a second validator
	validator2 := types.NewValidator(addrVals[1], PKs[1], types.Description{})
	validator2, pool, issuedShares = validator2.AddTokensFromDel(pool, valTokens)
	require.Equal(t, valTokens, issuedShares.RoundInt())
	keeper.SetPool(ctx, pool)
	validator2 = TestingUpdateValidator(keeper, ctx, validator2, true)

	header := ctx.BlockHeader()
	blockHeight := int64(10)
	header.Height = blockHeight
	blockTime := time.Unix(333, 0)
	header.Time = blockTime
	ctx = ctx.WithBlockHeader(header)

	// unbond the all self-delegation to put validator in unbonding state
	_, err := keeper.Undelegate(ctx, val0AccAddr, addrVals[0], sdk.NewDecFromInt(delTokens))
	require.NoError(t, err)

	// end block
	updates := keeper.ApplyAndReturnValidatorSetUpdates(ctx)
	require.Equal(t, 1, len(updates))

	validator, found := keeper.GetValidator(ctx, addrVals[0])
	require.True(t, found)
	require.Equal(t, blockHeight, validator.UnbondingHeight)
	params := keeper.GetParams(ctx)
	require.True(t, blockTime.Add(params.UnbondingTime).Equal(validator.UnbondingCompletionTime))

	//change the context
	header = ctx.BlockHeader()
	blockHeight2 := int64(20)
	header.Height = blockHeight2
	blockTime2 := time.Unix(444, 0)
	header.Time = blockTime2
	ctx = ctx.WithBlockHeader(header)

	// unbond some of the other delegation's shares
	redelegateTokens := sdk.TokensFromTendermintPower(6)
	_, err = keeper.BeginRedelegation(ctx, addrDels[0], addrVals[0], addrVals[1], sdk.NewDecFromInt(redelegateTokens))
	require.NoError(t, err)

	// retrieve the unbonding delegation
	ubd, found := keeper.GetRedelegation(ctx, addrDels[0], addrVals[0], addrVals[1])
	require.True(t, found)
	require.Len(t, ubd.Entries, 1)
	assert.Equal(t, blockHeight, ubd.Entries[0].CreationHeight)
	assert.True(t, blockTime.Add(params.UnbondingTime).Equal(ubd.Entries[0].CompletionTime))
}

func TestRedelegateFromUnbondedValidator(t *testing.T) {
	ctx, _, keeper := CreateTestInput(t, false, 0)
	pool := keeper.GetPool(ctx)
	startTokens := sdk.TokensFromTendermintPower(30)
	pool.NotBondedTokens = startTokens

	//create a validator with a self-delegation
	validator := types.NewValidator(addrVals[0], PKs[0], types.Description{})

	valTokens := sdk.TokensFromTendermintPower(10)
	validator, pool, issuedShares := validator.AddTokensFromDel(pool, valTokens)
	require.Equal(t, valTokens, issuedShares.RoundInt())
	keeper.SetPool(ctx, pool)
	validator = TestingUpdateValidator(keeper, ctx, validator, true)
	pool = keeper.GetPool(ctx)
	val0AccAddr := sdk.AccAddress(addrVals[0].Bytes())
	selfDelegation := types.NewDelegation(val0AccAddr, addrVals[0], issuedShares)
	keeper.SetDelegation(ctx, selfDelegation)

	// create a second delegation to this validator
	keeper.DeleteValidatorByPowerIndex(ctx, validator)
	delTokens := sdk.TokensFromTendermintPower(10)
	validator, pool, issuedShares = validator.AddTokensFromDel(pool, delTokens)
	require.Equal(t, delTokens, issuedShares.RoundInt())
	keeper.SetPool(ctx, pool)
	validator = TestingUpdateValidator(keeper, ctx, validator, true)
	pool = keeper.GetPool(ctx)
	delegation := types.NewDelegation(addrDels[0], addrVals[0], issuedShares)
	keeper.SetDelegation(ctx, delegation)

	// create a second validator
	validator2 := types.NewValidator(addrVals[1], PKs[1], types.Description{})
	validator2, pool, issuedShares = validator2.AddTokensFromDel(pool, valTokens)
	require.Equal(t, valTokens, issuedShares.RoundInt())
	keeper.SetPool(ctx, pool)
	validator2 = TestingUpdateValidator(keeper, ctx, validator2, true)
	require.Equal(t, sdk.Bonded, validator2.Status)

	ctx = ctx.WithBlockHeight(10)
	ctx = ctx.WithBlockTime(time.Unix(333, 0))

	// unbond the all self-delegation to put validator in unbonding state
	_, err := keeper.Undelegate(ctx, val0AccAddr, addrVals[0], sdk.NewDecFromInt(delTokens))
	require.NoError(t, err)

	// end block
	updates := keeper.ApplyAndReturnValidatorSetUpdates(ctx)
	require.Equal(t, 1, len(updates))

	validator, found := keeper.GetValidator(ctx, addrVals[0])
	require.True(t, found)
	require.Equal(t, ctx.BlockHeight(), validator.UnbondingHeight)
	params := keeper.GetParams(ctx)
	require.True(t, ctx.BlockHeader().Time.Add(params.UnbondingTime).Equal(validator.UnbondingCompletionTime))

	// unbond the validator
	keeper.unbondingToUnbonded(ctx, validator)

	// redelegate some of the delegation's shares
	redelegationTokens := sdk.TokensFromTendermintPower(6)
	_, err = keeper.BeginRedelegation(ctx, addrDels[0], addrVals[0], addrVals[1], sdk.NewDecFromInt(redelegationTokens))
	require.NoError(t, err)

	// no red should have been found
	red, found := keeper.GetRedelegation(ctx, addrDels[0], addrVals[0], addrVals[1])
	require.False(t, found, "%v", red)
}
