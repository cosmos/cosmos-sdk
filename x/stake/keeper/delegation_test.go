package keeper

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/stake/types"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// tests GetDelegation, GetDelegations, SetDelegation, RemoveDelegation, GetBonds
func TestDelegation(t *testing.T) {
	ctx, _, keeper := CreateTestInput(t, false, 0)

	//construct the validators
	amts := []int64{9, 8, 7}
	var validators [3]types.Validator
	for i, amt := range amts {
		validators[i] = types.NewValidator(addrVals[i], PKs[i], types.Description{})
		validators[i].PoolShares = types.NewUnbondedShares(sdk.NewRat(amt))
		validators[i].DelegatorShares = sdk.NewRat(amt)
	}

	// first add a validators[0] to delegate too
	validators[0] = keeper.UpdateValidator(ctx, validators[0])

	bond1to1 := types.Delegation{
		DelegatorAddr: addrDels[0],
		ValidatorAddr: addrVals[0],
		Shares:        sdk.NewRat(9),
	}

	// check the empty keeper first
	_, found := keeper.GetDelegation(ctx, addrDels[0], addrVals[0])
	assert.False(t, found)

	// set and retrieve a record
	keeper.SetDelegation(ctx, bond1to1)
	resBond, found := keeper.GetDelegation(ctx, addrDels[0], addrVals[0])
	assert.True(t, found)
	assert.True(t, bond1to1.Equal(resBond))

	// modify a records, save, and retrieve
	bond1to1.Shares = sdk.NewRat(99)
	keeper.SetDelegation(ctx, bond1to1)
	resBond, found = keeper.GetDelegation(ctx, addrDels[0], addrVals[0])
	assert.True(t, found)
	assert.True(t, bond1to1.Equal(resBond))

	// add some more records
	validators[1] = keeper.UpdateValidator(ctx, validators[1])
	validators[2] = keeper.UpdateValidator(ctx, validators[2])
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
	assert.True(t, bond1to1.Equal(resBonds[0]))
	assert.True(t, bond1to2.Equal(resBonds[1]))
	assert.True(t, bond1to3.Equal(resBonds[2]))
	resBonds = keeper.GetDelegations(ctx, addrDels[0], 3)
	require.Equal(t, 3, len(resBonds))
	resBonds = keeper.GetDelegations(ctx, addrDels[0], 2)
	require.Equal(t, 2, len(resBonds))
	resBonds = keeper.GetDelegations(ctx, addrDels[1], 5)
	require.Equal(t, 3, len(resBonds))
	assert.True(t, bond2to1.Equal(resBonds[0]))
	assert.True(t, bond2to2.Equal(resBonds[1]))
	assert.True(t, bond2to3.Equal(resBonds[2]))
	allBonds := keeper.GetAllDelegations(ctx)
	require.Equal(t, 6, len(allBonds))
	assert.True(t, bond1to1.Equal(allBonds[0]))
	assert.True(t, bond1to2.Equal(allBonds[1]))
	assert.True(t, bond1to3.Equal(allBonds[2]))
	assert.True(t, bond2to1.Equal(allBonds[3]))
	assert.True(t, bond2to2.Equal(allBonds[4]))
	assert.True(t, bond2to3.Equal(allBonds[5]))

	// delete a record
	keeper.RemoveDelegation(ctx, bond2to3)
	_, found = keeper.GetDelegation(ctx, addrDels[1], addrVals[2])
	assert.False(t, found)
	resBonds = keeper.GetDelegations(ctx, addrDels[1], 5)
	require.Equal(t, 2, len(resBonds))
	assert.True(t, bond2to1.Equal(resBonds[0]))
	assert.True(t, bond2to2.Equal(resBonds[1]))

	// delete all the records from delegator 2
	keeper.RemoveDelegation(ctx, bond2to1)
	keeper.RemoveDelegation(ctx, bond2to2)
	_, found = keeper.GetDelegation(ctx, addrDels[1], addrVals[0])
	assert.False(t, found)
	_, found = keeper.GetDelegation(ctx, addrDels[1], addrVals[1])
	assert.False(t, found)
	resBonds = keeper.GetDelegations(ctx, addrDels[1], 5)
	require.Equal(t, 0, len(resBonds))
}
