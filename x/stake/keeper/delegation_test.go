package keeper

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/stake/types"

	"github.com/stretchr/testify/require"
)

// tests GetDelegation, GetDelegations, SetDelegation, RemoveDelegation, GetDelegations
func TestDelegation(t *testing.T) {
	ctx, _, keeper := CreateTestInput(t, false, 10)
	pool := keeper.GetPool(ctx)

	//construct the validators
	amts := []int64{9, 8, 7}
	var validators [3]types.Validator
	for i, amt := range amts {
		validators[i] = types.NewValidator(addrVals[i], PKs[i], types.Description{})
		validators[i], pool, _ = validators[i].AddTokensFromDel(pool, amt)
	}

	keeper.SetPool(ctx, pool)
	validators[0] = keeper.UpdateValidator(ctx, validators[0])
	validators[1] = keeper.UpdateValidator(ctx, validators[1])
	validators[2] = keeper.UpdateValidator(ctx, validators[2])

	// first add a validators[0] to delegate too
	bond1to1 := types.Delegation{
		DelegatorAddr: addrDels[0],
		ValidatorAddr: addrVals[0],
		Shares:        sdk.NewRat(9),
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
	bond1to1.Shares = sdk.NewRat(99)
	keeper.SetDelegation(ctx, bond1to1)
	resBond, found = keeper.GetDelegation(ctx, addrDels[0], addrVals[0])
	require.True(t, found)
	require.True(t, bond1to1.Equal(resBond))

	// add some more records
	bond1to2 := types.Delegation{addrDels[0], addrVals[1], sdk.NewRat(9), 0}
	bond1to3 := types.Delegation{addrDels[0], addrVals[2], sdk.NewRat(9), 1}
	bond2to1 := types.Delegation{addrDels[1], addrVals[0], sdk.NewRat(9), 2}
	bond2to2 := types.Delegation{addrDels[1], addrVals[1], sdk.NewRat(9), 3}
	bond2to3 := types.Delegation{addrDels[1], addrVals[2], sdk.NewRat(9), 4}
	keeper.SetDelegation(ctx, bond1to2)
	keeper.SetDelegation(ctx, bond1to3)
	keeper.SetDelegation(ctx, bond2to1)
	keeper.SetDelegation(ctx, bond2to2)
	keeper.SetDelegation(ctx, bond2to3)

	// test all bond retrieve capabilities
	resBonds := keeper.GetDelegations(ctx, addrDels[0], 5)
	require.Equal(t, 3, len(resBonds))
	require.True(t, bond1to1.Equal(resBonds[0]))
	require.True(t, bond1to2.Equal(resBonds[1]))
	require.True(t, bond1to3.Equal(resBonds[2]))
	resBonds = keeper.GetDelegations(ctx, addrDels[0], 3)
	require.Equal(t, 3, len(resBonds))
	resBonds = keeper.GetDelegations(ctx, addrDels[0], 2)
	require.Equal(t, 2, len(resBonds))
	resBonds = keeper.GetDelegations(ctx, addrDels[1], 5)
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

	// delete a record
	keeper.RemoveDelegation(ctx, bond2to3)
	_, found = keeper.GetDelegation(ctx, addrDels[1], addrVals[2])
	require.False(t, found)
	resBonds = keeper.GetDelegations(ctx, addrDels[1], 5)
	require.Equal(t, 2, len(resBonds))
	require.True(t, bond2to1.Equal(resBonds[0]))
	require.True(t, bond2to2.Equal(resBonds[1]))

	// delete all the records from delegator 2
	keeper.RemoveDelegation(ctx, bond2to1)
	keeper.RemoveDelegation(ctx, bond2to2)
	_, found = keeper.GetDelegation(ctx, addrDels[1], addrVals[0])
	require.False(t, found)
	_, found = keeper.GetDelegation(ctx, addrDels[1], addrVals[1])
	require.False(t, found)
	resBonds = keeper.GetDelegations(ctx, addrDels[1], 5)
	require.Equal(t, 0, len(resBonds))
}

// tests Get/Set/Remove UnbondingDelegation
func TestUnbondingDelegation(t *testing.T) {
	ctx, _, keeper := CreateTestInput(t, false, 0)

	ubd := types.UnbondingDelegation{
		DelegatorAddr:  addrDels[0],
		ValidatorAddr:  addrVals[0],
		CreationHeight: 0,
		MinTime:        0,
		Balance:        sdk.NewCoin("steak", 5),
	}

	// set and retrieve a record
	keeper.SetUnbondingDelegation(ctx, ubd)
	resBond, found := keeper.GetUnbondingDelegation(ctx, addrDels[0], addrVals[0])
	require.True(t, found)
	require.True(t, ubd.Equal(resBond))

	// modify a records, save, and retrieve
	ubd.Balance = sdk.NewCoin("steak", 21)
	keeper.SetUnbondingDelegation(ctx, ubd)
	resBond, found = keeper.GetUnbondingDelegation(ctx, addrDels[0], addrVals[0])
	require.True(t, found)
	require.True(t, ubd.Equal(resBond))

	// delete a record
	keeper.RemoveUnbondingDelegation(ctx, ubd)
	_, found = keeper.GetUnbondingDelegation(ctx, addrDels[0], addrVals[0])
	require.False(t, found)
}

func TestUnbondDelegation(t *testing.T) {
	ctx, _, keeper := CreateTestInput(t, false, 0)
	pool := keeper.GetPool(ctx)
	pool.LooseTokens = sdk.NewRat(10)

	//create a validator and a delegator to that validator
	validator := types.NewValidator(addrVals[0], PKs[0], types.Description{})
	validator, pool, issuedShares := validator.AddTokensFromDel(pool, 10)
	require.Equal(t, int64(10), issuedShares.RoundInt64())
	keeper.SetPool(ctx, pool)
	validator = keeper.UpdateValidator(ctx, validator)

	pool = keeper.GetPool(ctx)
	require.Equal(t, int64(10), pool.BondedTokens.RoundInt64())
	require.Equal(t, int64(10), validator.BondedTokens().RoundInt64())

	delegation := types.Delegation{
		DelegatorAddr: addrDels[0],
		ValidatorAddr: addrVals[0],
		Shares:        issuedShares,
	}
	keeper.SetDelegation(ctx, delegation)

	var err error
	var amount sdk.Rat
	amount, err = keeper.unbond(ctx, addrDels[0], addrVals[0], sdk.NewRat(6))
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

// Make sure that that the retrieving the delegations doesn't affect the state
func TestGetRedelegationsFromValidator(t *testing.T) {
	ctx, _, keeper := CreateTestInput(t, false, 0)

	rd := types.Redelegation{
		DelegatorAddr:    addrDels[0],
		ValidatorSrcAddr: addrVals[0],
		ValidatorDstAddr: addrVals[1],
		CreationHeight:   0,
		MinTime:          0,
		SharesSrc:        sdk.NewRat(5),
		SharesDst:        sdk.NewRat(5),
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
		MinTime:          0,
		SharesSrc:        sdk.NewRat(5),
		SharesDst:        sdk.NewRat(5),
	}

	// test shouldn't have and redelegations
	has := keeper.HasReceivingRedelegation(ctx, addrDels[0], addrVals[1])
	require.False(t, has)

	// set and retrieve a record
	keeper.SetRedelegation(ctx, rd)
	resBond, found := keeper.GetRedelegation(ctx, addrDels[0], addrVals[0], addrVals[1])
	require.True(t, found)

	redelegations := keeper.GetRedelegationsFromValidator(ctx, addrVals[0])
	require.Equal(t, 1, len(redelegations))
	require.True(t, redelegations[0].Equal(resBond))

	// check if has the redelegation
	has = keeper.HasReceivingRedelegation(ctx, addrDels[0], addrVals[1])
	require.True(t, has)

	// modify a records, save, and retrieve
	rd.SharesSrc = sdk.NewRat(21)
	rd.SharesDst = sdk.NewRat(21)
	keeper.SetRedelegation(ctx, rd)

	resBond, found = keeper.GetRedelegation(ctx, addrDels[0], addrVals[0], addrVals[1])
	require.True(t, found)
	require.True(t, rd.Equal(resBond))

	redelegations = keeper.GetRedelegationsFromValidator(ctx, addrVals[0])
	require.Equal(t, 1, len(redelegations))
	require.True(t, redelegations[0].Equal(resBond))

	// delete a record
	keeper.RemoveRedelegation(ctx, rd)
	_, found = keeper.GetRedelegation(ctx, addrDels[0], addrVals[0], addrVals[1])
	require.False(t, found)
}
