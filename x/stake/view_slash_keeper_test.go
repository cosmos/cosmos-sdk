package stake

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// tests GetDelegation, GetDelegations
func TestViewSlashBond(t *testing.T) {
	ctx, _, keeper := createTestInput(t, false, sdk.NewInt(0))

	//construct the validators
	amts := []int64{9, 8, 7}
	var validators [3]Validator
	for i, amt := range amts {
		validators[i] = Validator{
			Owner:           addrVals[i],
			PubKey:          pks[i],
			PoolShares:      NewUnbondedShares(sdk.NewRat(amt)),
			DelegatorShares: sdk.NewRat(amt),
		}
	}

	// first add a validators[0] to delegate too
	keeper.updateValidator(ctx, validators[0])

	bond1to1 := Delegation{
		DelegatorAddr: addrDels[0],
		ValidatorAddr: addrVals[0],
		Shares:        sdk.NewRat(9),
	}

	viewSlashKeeper := NewViewSlashKeeper(keeper)

	// check the empty keeper first
	_, found := viewSlashKeeper.GetDelegation(ctx, addrDels[0], addrVals[0])
	assert.False(t, found)

	// set and retrieve a record
	keeper.setDelegation(ctx, bond1to1)
	resBond, found := viewSlashKeeper.GetDelegation(ctx, addrDels[0], addrVals[0])
	assert.True(t, found)
	assert.True(t, bond1to1.equal(resBond))

	// modify a records, save, and retrieve
	bond1to1.Shares = sdk.NewRat(99)
	keeper.setDelegation(ctx, bond1to1)
	resBond, found = viewSlashKeeper.GetDelegation(ctx, addrDels[0], addrVals[0])
	assert.True(t, found)
	assert.True(t, bond1to1.equal(resBond))

	// add some more records
	keeper.updateValidator(ctx, validators[1])
	keeper.updateValidator(ctx, validators[2])
	bond1to2 := Delegation{addrDels[0], addrVals[1], sdk.NewRat(9), 0}
	bond1to3 := Delegation{addrDels[0], addrVals[2], sdk.NewRat(9), 1}
	bond2to1 := Delegation{addrDels[1], addrVals[0], sdk.NewRat(9), 2}
	bond2to2 := Delegation{addrDels[1], addrVals[1], sdk.NewRat(9), 3}
	bond2to3 := Delegation{addrDels[1], addrVals[2], sdk.NewRat(9), 4}
	keeper.setDelegation(ctx, bond1to2)
	keeper.setDelegation(ctx, bond1to3)
	keeper.setDelegation(ctx, bond2to1)
	keeper.setDelegation(ctx, bond2to2)
	keeper.setDelegation(ctx, bond2to3)

	// test all bond retrieve capabilities
	resBonds := viewSlashKeeper.GetDelegations(ctx, addrDels[0], 5)
	require.Equal(t, 3, len(resBonds))
	assert.True(t, bond1to1.equal(resBonds[0]))
	assert.True(t, bond1to2.equal(resBonds[1]))
	assert.True(t, bond1to3.equal(resBonds[2]))
	resBonds = viewSlashKeeper.GetDelegations(ctx, addrDels[0], 3)
	require.Equal(t, 3, len(resBonds))
	resBonds = viewSlashKeeper.GetDelegations(ctx, addrDels[0], 2)
	require.Equal(t, 2, len(resBonds))
	resBonds = viewSlashKeeper.GetDelegations(ctx, addrDels[1], 5)
	require.Equal(t, 3, len(resBonds))
	assert.True(t, bond2to1.equal(resBonds[0]))
	assert.True(t, bond2to2.equal(resBonds[1]))
	assert.True(t, bond2to3.equal(resBonds[2]))

}
