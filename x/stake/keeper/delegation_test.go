package keeper

import (
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/stake/types"

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
	validators[0] = testingUpdateValidator(keeper, ctx, validators[0])
	validators[1] = testingUpdateValidator(keeper, ctx, validators[1])
	validators[2] = testingUpdateValidator(keeper, ctx, validators[2])

	// first add a validators[0] to delegate too

	bond1to1 := types.Delegation{
		DelegatorAddr: addrDels[0],
		ValidatorAddr: addrVals[0],
		Shares:        sdk.NewDec(9),
	}

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
	bond1to2 := types.Delegation{addrDels[0], addrVals[1], sdk.NewDec(9), 0}
	bond1to3 := types.Delegation{addrDels[0], addrVals[2], sdk.NewDec(9), 1}
	bond2to1 := types.Delegation{addrDels[1], addrVals[0], sdk.NewDec(9), 2}
	bond2to2 := types.Delegation{addrDels[1], addrVals[1], sdk.NewDec(9), 3}
	bond2to3 := types.Delegation{addrDels[1], addrVals[2], sdk.NewDec(9), 4}
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

	ubd := types.UnbondingDelegation{
		DelegatorAddr:  addrDels[0],
		ValidatorAddr:  addrVals[0],
		CreationHeight: 0,
		MinTime:        time.Unix(0, 0),
		Balance:        sdk.NewInt64Coin("steak", 5),
	}

	// set and retrieve a record
	keeper.SetUnbondingDelegation(ctx, ubd)
	resUnbond, found := keeper.GetUnbondingDelegation(ctx, addrDels[0], addrVals[0])
	require.True(t, found)
	require.True(t, ubd.Equal(resUnbond))

	// modify a records, save, and retrieve
	ubd.Balance = sdk.NewInt64Coin("steak", 21)
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
	pool.LooseTokens = sdk.NewDec(10)

	//create a validator and a delegator to that validator
	validator := types.NewValidator(addrVals[0], PKs[0], types.Description{})
	validator, pool, issuedShares := validator.AddTokensFromDel(pool, sdk.NewInt(10))
	require.Equal(t, int64(10), issuedShares.RoundInt64())
	keeper.SetPool(ctx, pool)
	validator = testingUpdateValidator(keeper, ctx, validator)

	pool = keeper.GetPool(ctx)
	require.Equal(t, int64(10), pool.BondedTokens.RoundInt64())
	require.Equal(t, int64(10), validator.BondedTokens().RoundInt64())

	delegation := types.Delegation{
		DelegatorAddr: addrDels[0],
		ValidatorAddr: addrVals[0],
		Shares:        issuedShares,
	}
	keeper.SetDelegation(ctx, delegation)

	amount, err := keeper.unbond(ctx, addrDels[0], addrVals[0], sdk.NewDec(6))
	require.NoError(t, err)
	require.Equal(t, int64(6), amount.RoundInt64()) // shares to be added to an unbonding delegation / redelegation

	delegation, found := keeper.GetDelegation(ctx, addrDels[0], addrVals[0])
	require.True(t, found)
	validator, found = keeper.GetValidator(ctx, addrVals[0])
	require.True(t, found)
	pool = keeper.GetPool(ctx)

	require.Equal(t, int64(4), delegation.Shares.RoundInt64())
	require.Equal(t, int64(4), validator.BondedTokens().RoundInt64())
	require.Equal(t, int64(6), pool.LooseTokens.RoundInt64(), "%v", pool)
	require.Equal(t, int64(4), pool.BondedTokens.RoundInt64())
}

// test removing all self delegation from a validator which should
// shift it from the bonded to unbonded state
func TestUndelegateSelfDelegation(t *testing.T) {

	ctx, _, keeper := CreateTestInput(t, false, 0)
	pool := keeper.GetPool(ctx)
	pool.LooseTokens = sdk.NewDec(20)

	//create a validator with a self-delegation
	validator := types.NewValidator(addrVals[0], PKs[0], types.Description{})
	validator, pool, issuedShares := validator.AddTokensFromDel(pool, sdk.NewInt(10))
	require.Equal(t, int64(10), issuedShares.RoundInt64())
	keeper.SetPool(ctx, pool)
	validator = testingUpdateValidator(keeper, ctx, validator)
	pool = keeper.GetPool(ctx)
	selfDelegation := types.Delegation{
		DelegatorAddr: sdk.AccAddress(addrVals[0].Bytes()),
		ValidatorAddr: addrVals[0],
		Shares:        issuedShares,
	}
	keeper.SetDelegation(ctx, selfDelegation)

	// create a second delegation to this validator
	keeper.DeleteValidatorByPowerIndex(ctx, validator, pool)
	validator, pool, issuedShares = validator.AddTokensFromDel(pool, sdk.NewInt(10))
	require.Equal(t, int64(10), issuedShares.RoundInt64())
	keeper.SetPool(ctx, pool)
	validator = testingUpdateValidator(keeper, ctx, validator)
	pool = keeper.GetPool(ctx)
	delegation := types.Delegation{
		DelegatorAddr: addrDels[0],
		ValidatorAddr: addrVals[0],
		Shares:        issuedShares,
	}
	keeper.SetDelegation(ctx, delegation)

	val0AccAddr := sdk.AccAddress(addrVals[0].Bytes())
	_, err := keeper.BeginUnbonding(ctx, val0AccAddr, addrVals[0], sdk.NewDec(10))
	require.NoError(t, err)

	// end block
	updates := keeper.ApplyAndReturnValidatorSetUpdates(ctx)
	require.Equal(t, 1, len(updates))

	validator, found := keeper.GetValidator(ctx, addrVals[0])
	require.True(t, found)
	require.Equal(t, int64(10), validator.Tokens.RoundInt64())
	require.Equal(t, sdk.Unbonding, validator.Status)
}

func TestUndelegateFromUnbondingValidator(t *testing.T) {
	ctx, _, keeper := CreateTestInput(t, false, 0)
	pool := keeper.GetPool(ctx)
	pool.LooseTokens = sdk.NewDec(20)

	//create a validator with a self-delegation
	validator := types.NewValidator(addrVals[0], PKs[0], types.Description{})

	validator, pool, issuedShares := validator.AddTokensFromDel(pool, sdk.NewInt(10))
	require.Equal(t, int64(10), issuedShares.RoundInt64())
	keeper.SetPool(ctx, pool)
	validator = testingUpdateValidator(keeper, ctx, validator)
	pool = keeper.GetPool(ctx)
	selfDelegation := types.Delegation{
		DelegatorAddr: sdk.AccAddress(addrVals[0].Bytes()),
		ValidatorAddr: addrVals[0],
		Shares:        issuedShares,
	}
	keeper.SetDelegation(ctx, selfDelegation)

	// create a second delegation to this validator
	keeper.DeleteValidatorByPowerIndex(ctx, validator, pool)
	validator, pool, issuedShares = validator.AddTokensFromDel(pool, sdk.NewInt(10))
	require.Equal(t, int64(10), issuedShares.RoundInt64())
	keeper.SetPool(ctx, pool)
	validator = testingUpdateValidator(keeper, ctx, validator)
	pool = keeper.GetPool(ctx)
	delegation := types.Delegation{
		DelegatorAddr: addrDels[0],
		ValidatorAddr: addrVals[0],
		Shares:        issuedShares,
	}
	keeper.SetDelegation(ctx, delegation)

	header := ctx.BlockHeader()
	blockHeight := int64(10)
	header.Height = blockHeight
	blockTime := time.Unix(333, 0)
	header.Time = blockTime
	ctx = ctx.WithBlockHeader(header)

	// unbond the all self-delegation to put validator in unbonding state
	val0AccAddr := sdk.AccAddress(addrVals[0].Bytes())
	_, err := keeper.BeginUnbonding(ctx, val0AccAddr, addrVals[0], sdk.NewDec(10))
	require.NoError(t, err)

	// end block
	updates := keeper.ApplyAndReturnValidatorSetUpdates(ctx)
	require.Equal(t, 1, len(updates))

	validator, found := keeper.GetValidator(ctx, addrVals[0])
	require.True(t, found)
	require.Equal(t, blockHeight, validator.UnbondingHeight)
	params := keeper.GetParams(ctx)
	require.True(t, blockTime.Add(params.UnbondingTime).Equal(validator.UnbondingMinTime))

	//change the context
	header = ctx.BlockHeader()
	blockHeight2 := int64(20)
	header.Height = blockHeight2
	blockTime2 := time.Unix(444, 0)
	header.Time = blockTime2
	ctx = ctx.WithBlockHeader(header)

	// unbond some of the other delegation's shares
	_, err = keeper.BeginUnbonding(ctx, addrDels[0], addrVals[0], sdk.NewDec(6))
	require.NoError(t, err)

	// retrieve the unbonding delegation
	ubd, found := keeper.GetUnbondingDelegation(ctx, addrDels[0], addrVals[0])
	require.True(t, found)
	require.True(t, ubd.Balance.IsEqual(sdk.NewInt64Coin(params.BondDenom, 6)))
	assert.Equal(t, blockHeight, ubd.CreationHeight)
	assert.True(t, blockTime.Add(params.UnbondingTime).Equal(ubd.MinTime))
}

func TestUndelegateFromUnbondedValidator(t *testing.T) {
	ctx, _, keeper := CreateTestInput(t, false, 0)
	pool := keeper.GetPool(ctx)
	pool.LooseTokens = sdk.NewDec(20)

	//create a validator with a self-delegation
	validator := types.NewValidator(addrVals[0], PKs[0], types.Description{})

	validator, pool, issuedShares := validator.AddTokensFromDel(pool, sdk.NewInt(10))
	require.Equal(t, int64(10), issuedShares.RoundInt64())
	keeper.SetPool(ctx, pool)
	validator = testingUpdateValidator(keeper, ctx, validator)
	pool = keeper.GetPool(ctx)
	val0AccAddr := sdk.AccAddress(addrVals[0].Bytes())
	selfDelegation := types.Delegation{
		DelegatorAddr: val0AccAddr,
		ValidatorAddr: addrVals[0],
		Shares:        issuedShares,
	}
	keeper.SetDelegation(ctx, selfDelegation)

	// create a second delegation to this validator
	keeper.DeleteValidatorByPowerIndex(ctx, validator, pool)
	validator, pool, issuedShares = validator.AddTokensFromDel(pool, sdk.NewInt(10))
	require.Equal(t, int64(10), issuedShares.RoundInt64())
	keeper.SetPool(ctx, pool)
	validator = testingUpdateValidator(keeper, ctx, validator)
	pool = keeper.GetPool(ctx)
	delegation := types.Delegation{
		DelegatorAddr: addrDels[0],
		ValidatorAddr: addrVals[0],
		Shares:        issuedShares,
	}
	keeper.SetDelegation(ctx, delegation)

	header := ctx.BlockHeader()
	blockHeight := int64(10)
	header.Height = blockHeight
	blockTime := time.Unix(333, 0)
	header.Time = blockTime
	ctx = ctx.WithBlockHeader(header)

	// unbond the all self-delegation to put validator in unbonding state
	_, err := keeper.BeginUnbonding(ctx, val0AccAddr, addrVals[0], sdk.NewDec(10))
	require.NoError(t, err)

	// end block
	updates := keeper.ApplyAndReturnValidatorSetUpdates(ctx)
	require.Equal(t, 1, len(updates))

	validator, found := keeper.GetValidator(ctx, addrVals[0])
	require.True(t, found)
	require.Equal(t, blockHeight, validator.UnbondingHeight)
	params := keeper.GetParams(ctx)
	require.True(t, blockTime.Add(params.UnbondingTime).Equal(validator.UnbondingMinTime))

	// change the context to one which makes the validator considered unbonded
	header = ctx.BlockHeader()
	blockHeight2 := int64(20)
	header.Height = blockHeight2
	blockTime2 := time.Unix(444, 0).Add(params.UnbondingTime)
	header.Time = blockTime2
	ctx = ctx.WithBlockHeader(header)

	// unbond some of the other delegation's shares
	_, err = keeper.BeginUnbonding(ctx, addrDels[0], addrVals[0], sdk.NewDec(6))
	require.NoError(t, err)

	// no ubd should have been found, coins should have been returned direcly to account
	ubd, found := keeper.GetUnbondingDelegation(ctx, addrDels[0], addrVals[0])
	require.False(t, found, "%v", ubd)
}

// Make sure that that the retrieving the delegations doesn't affect the state
func TestGetRedelegationsFromValidator(t *testing.T) {
	ctx, _, keeper := CreateTestInput(t, false, 0)

	rd := types.Redelegation{
		DelegatorAddr:    addrDels[0],
		ValidatorSrcAddr: addrVals[0],
		ValidatorDstAddr: addrVals[1],
		CreationHeight:   0,
		MinTime:          time.Unix(0, 0),
		SharesSrc:        sdk.NewDec(5),
		SharesDst:        sdk.NewDec(5),
	}

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

	rd := types.Redelegation{
		DelegatorAddr:    addrDels[0],
		ValidatorSrcAddr: addrVals[0],
		ValidatorDstAddr: addrVals[1],
		CreationHeight:   0,
		MinTime:          time.Unix(0, 0),
		SharesSrc:        sdk.NewDec(5),
		SharesDst:        sdk.NewDec(5),
	}

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

	redelegations = keeper.GetAllRedelegations(ctx, addrDels[0])
	require.Equal(t, 1, len(redelegations))
	require.True(t, redelegations[0].Equal(resRed))

	// check if has the redelegation
	has = keeper.HasReceivingRedelegation(ctx, addrDels[0], addrVals[1])
	require.True(t, has)

	// modify a records, save, and retrieve
	rd.SharesSrc = sdk.NewDec(21)
	rd.SharesDst = sdk.NewDec(21)
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

	redelegations = keeper.GetAllRedelegations(ctx, addrDels[0])
	require.Equal(t, 0, len(redelegations))
}

func TestRedelegateSelfDelegation(t *testing.T) {

	ctx, _, keeper := CreateTestInput(t, false, 0)
	pool := keeper.GetPool(ctx)
	pool.LooseTokens = sdk.NewDec(30)

	//create a validator with a self-delegation
	validator := types.NewValidator(addrVals[0], PKs[0], types.Description{})
	validator, pool, issuedShares := validator.AddTokensFromDel(pool, sdk.NewInt(10))
	require.Equal(t, int64(10), issuedShares.RoundInt64())
	keeper.SetPool(ctx, pool)
	validator = testingUpdateValidator(keeper, ctx, validator)
	pool = keeper.GetPool(ctx)
	val0AccAddr := sdk.AccAddress(addrVals[0].Bytes())
	selfDelegation := types.Delegation{
		DelegatorAddr: val0AccAddr,
		ValidatorAddr: addrVals[0],
		Shares:        issuedShares,
	}
	keeper.SetDelegation(ctx, selfDelegation)

	// create a second validator
	validator2 := types.NewValidator(addrVals[1], PKs[1], types.Description{})
	validator2, pool, issuedShares = validator2.AddTokensFromDel(pool, sdk.NewInt(10))
	require.Equal(t, int64(10), issuedShares.RoundInt64())
	pool.BondedTokens = pool.BondedTokens.Add(sdk.NewDec(10))
	keeper.SetPool(ctx, pool)
	validator2 = testingUpdateValidator(keeper, ctx, validator2)
	require.Equal(t, sdk.Bonded, validator2.Status)

	// create a second delegation to this validator
	validator, pool, issuedShares = validator.AddTokensFromDel(pool, sdk.NewInt(10))
	require.Equal(t, int64(10), issuedShares.RoundInt64())
	keeper.SetPool(ctx, pool)
	validator = testingUpdateValidator(keeper, ctx, validator)
	pool = keeper.GetPool(ctx)
	delegation := types.Delegation{
		DelegatorAddr: addrDels[0],
		ValidatorAddr: addrVals[0],
		Shares:        issuedShares,
	}
	keeper.SetDelegation(ctx, delegation)

	_, err := keeper.BeginRedelegation(ctx, val0AccAddr, addrVals[0], addrVals[1], sdk.NewDec(10))
	require.NoError(t, err)

	// end block
	updates := keeper.ApplyAndReturnValidatorSetUpdates(ctx)
	require.Equal(t, 2, len(updates))

	validator, found := keeper.GetValidator(ctx, addrVals[0])
	require.True(t, found)
	require.Equal(t, int64(10), validator.Tokens.RoundInt64())
	require.Equal(t, sdk.Unbonding, validator.Status)
}

func TestRedelegateFromUnbondingValidator(t *testing.T) {
	ctx, _, keeper := CreateTestInput(t, false, 0)
	pool := keeper.GetPool(ctx)
	pool.LooseTokens = sdk.NewDec(30)

	//create a validator with a self-delegation
	validator := types.NewValidator(addrVals[0], PKs[0], types.Description{})
	validator.BondIntraTxCounter = 1

	validator, pool, issuedShares := validator.AddTokensFromDel(pool, sdk.NewInt(10))
	require.Equal(t, int64(10), issuedShares.RoundInt64())
	keeper.SetPool(ctx, pool)
	validator = testingUpdateValidator(keeper, ctx, validator)
	pool = keeper.GetPool(ctx)
	val0AccAddr := sdk.AccAddress(addrVals[0].Bytes())
	selfDelegation := types.Delegation{
		DelegatorAddr: val0AccAddr,
		ValidatorAddr: addrVals[0],
		Shares:        issuedShares,
	}
	keeper.SetDelegation(ctx, selfDelegation)

	// create a second delegation to this validator
	keeper.DeleteValidatorByPowerIndex(ctx, validator, pool)
	validator, pool, issuedShares = validator.AddTokensFromDel(pool, sdk.NewInt(10))
	require.Equal(t, int64(10), issuedShares.RoundInt64())
	keeper.SetPool(ctx, pool)
	validator = testingUpdateValidator(keeper, ctx, validator)
	pool = keeper.GetPool(ctx)
	delegation := types.Delegation{
		DelegatorAddr: addrDels[0],
		ValidatorAddr: addrVals[0],
		Shares:        issuedShares,
	}
	keeper.SetDelegation(ctx, delegation)

	// create a second validator
	validator2 := types.NewValidator(addrVals[1], PKs[1], types.Description{})
	validator2.BondIntraTxCounter = 2
	validator2, pool, issuedShares = validator2.AddTokensFromDel(pool, sdk.NewInt(10))
	require.Equal(t, int64(10), issuedShares.RoundInt64())
	keeper.SetPool(ctx, pool)
	validator2 = testingUpdateValidator(keeper, ctx, validator2)

	header := ctx.BlockHeader()
	blockHeight := int64(10)
	header.Height = blockHeight
	blockTime := time.Unix(333, 0)
	header.Time = blockTime
	ctx = ctx.WithBlockHeader(header)

	// unbond the all self-delegation to put validator in unbonding state
	_, err := keeper.BeginUnbonding(ctx, val0AccAddr, addrVals[0], sdk.NewDec(10))
	require.NoError(t, err)

	// end block
	updates := keeper.ApplyAndReturnValidatorSetUpdates(ctx)
	require.Equal(t, 1, len(updates))

	validator, found := keeper.GetValidator(ctx, addrVals[0])
	require.True(t, found)
	require.Equal(t, blockHeight, validator.UnbondingHeight)
	params := keeper.GetParams(ctx)
	require.True(t, blockTime.Add(params.UnbondingTime).Equal(validator.UnbondingMinTime))

	//change the context
	header = ctx.BlockHeader()
	blockHeight2 := int64(20)
	header.Height = blockHeight2
	blockTime2 := time.Unix(444, 0)
	header.Time = blockTime2
	ctx = ctx.WithBlockHeader(header)

	// unbond some of the other delegation's shares
	_, err = keeper.BeginRedelegation(ctx, addrDels[0], addrVals[0], addrVals[1], sdk.NewDec(6))
	require.NoError(t, err)

	// retrieve the unbonding delegation
	ubd, found := keeper.GetRedelegation(ctx, addrDels[0], addrVals[0], addrVals[1])
	require.True(t, found)
	require.True(t, ubd.Balance.IsEqual(sdk.NewInt64Coin(params.BondDenom, 6)))
	assert.Equal(t, blockHeight, ubd.CreationHeight)
	assert.True(t, blockTime.Add(params.UnbondingTime).Equal(ubd.MinTime))
}

func TestRedelegateFromUnbondedValidator(t *testing.T) {
	ctx, _, keeper := CreateTestInput(t, false, 0)
	pool := keeper.GetPool(ctx)
	pool.LooseTokens = sdk.NewDec(30)

	//create a validator with a self-delegation
	validator := types.NewValidator(addrVals[0], PKs[0], types.Description{})

	validator, pool, issuedShares := validator.AddTokensFromDel(pool, sdk.NewInt(10))
	require.Equal(t, int64(10), issuedShares.RoundInt64())
	keeper.SetPool(ctx, pool)
	validator = testingUpdateValidator(keeper, ctx, validator)
	pool = keeper.GetPool(ctx)
	val0AccAddr := sdk.AccAddress(addrVals[0].Bytes())
	selfDelegation := types.Delegation{
		DelegatorAddr: val0AccAddr,
		ValidatorAddr: addrVals[0],
		Shares:        issuedShares,
	}
	keeper.SetDelegation(ctx, selfDelegation)

	// create a second delegation to this validator
	keeper.DeleteValidatorByPowerIndex(ctx, validator, pool)
	validator, pool, issuedShares = validator.AddTokensFromDel(pool, sdk.NewInt(10))
	validator.BondIntraTxCounter = 1
	require.Equal(t, int64(10), issuedShares.RoundInt64())
	keeper.SetPool(ctx, pool)
	validator = testingUpdateValidator(keeper, ctx, validator)
	pool = keeper.GetPool(ctx)
	delegation := types.Delegation{
		DelegatorAddr: addrDels[0],
		ValidatorAddr: addrVals[0],
		Shares:        issuedShares,
	}
	keeper.SetDelegation(ctx, delegation)

	// create a second validator
	validator2 := types.NewValidator(addrVals[1], PKs[1], types.Description{})
	validator2.BondIntraTxCounter = 2
	validator2, pool, issuedShares = validator2.AddTokensFromDel(pool, sdk.NewInt(10))
	require.Equal(t, int64(10), issuedShares.RoundInt64())
	keeper.SetPool(ctx, pool)
	validator2 = testingUpdateValidator(keeper, ctx, validator2)
	require.Equal(t, sdk.Bonded, validator2.Status)

	header := ctx.BlockHeader()
	blockHeight := int64(10)
	header.Height = blockHeight
	blockTime := time.Unix(333, 0)
	header.Time = blockTime
	ctx = ctx.WithBlockHeader(header)

	// unbond the all self-delegation to put validator in unbonding state
	_, err := keeper.BeginUnbonding(ctx, val0AccAddr, addrVals[0], sdk.NewDec(10))
	require.NoError(t, err)

	// end block
	updates := keeper.ApplyAndReturnValidatorSetUpdates(ctx)
	require.Equal(t, 1, len(updates))

	validator, found := keeper.GetValidator(ctx, addrVals[0])
	require.True(t, found)
	require.Equal(t, blockHeight, validator.UnbondingHeight)
	params := keeper.GetParams(ctx)
	require.True(t, blockTime.Add(params.UnbondingTime).Equal(validator.UnbondingMinTime))

	// change the context to one which makes the validator considered unbonded
	header = ctx.BlockHeader()
	blockHeight2 := int64(20)
	header.Height = blockHeight2
	blockTime2 := time.Unix(444, 0).Add(params.UnbondingTime)
	header.Time = blockTime2
	ctx = ctx.WithBlockHeader(header)

	// unbond some of the other delegation's shares
	_, err = keeper.BeginRedelegation(ctx, addrDels[0], addrVals[0], addrVals[1], sdk.NewDec(6))
	require.NoError(t, err)

	// no ubd should have been found, coins should have been returned direcly to account
	ubd, found := keeper.GetRedelegation(ctx, addrDels[0], addrVals[0], addrVals[1])
	require.False(t, found, "%v", ubd)
}
