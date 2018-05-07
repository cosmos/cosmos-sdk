package stake

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// tests GetDelegatorBond, GetDelegatorBonds
func TestViewSlashBond(t *testing.T) {
	ctx, _, keeper := createTestInput(t, false, 0)

	//construct the candidates
	amts := []int64{9, 8, 7}
	var candidates [3]Candidate
	for i, amt := range amts {
		candidates[i] = Candidate{
			Address:     addrVals[i],
			PubKey:      pks[i],
			Assets:      sdk.NewRat(amt),
			Liabilities: sdk.NewRat(amt),
		}
	}

	// first add a candidates[0] to delegate too
	keeper.setCandidate(ctx, candidates[0])

	bond1to1 := DelegatorBond{
		DelegatorAddr: addrDels[0],
		CandidateAddr: addrVals[0],
		Shares:        sdk.NewRat(9),
	}

	viewSlashKeeper := NewViewSlashKeeper(keeper)

	// check the empty keeper first
	_, found := viewSlashKeeper.GetDelegatorBond(ctx, addrDels[0], addrVals[0])
	assert.False(t, found)

	// set and retrieve a record
	keeper.setDelegatorBond(ctx, bond1to1)
	resBond, found := viewSlashKeeper.GetDelegatorBond(ctx, addrDels[0], addrVals[0])
	assert.True(t, found)
	assert.True(t, bondsEqual(bond1to1, resBond))

	// modify a records, save, and retrieve
	bond1to1.Shares = sdk.NewRat(99)
	keeper.setDelegatorBond(ctx, bond1to1)
	resBond, found = viewSlashKeeper.GetDelegatorBond(ctx, addrDels[0], addrVals[0])
	assert.True(t, found)
	assert.True(t, bondsEqual(bond1to1, resBond))

	// add some more records
	keeper.setCandidate(ctx, candidates[1])
	keeper.setCandidate(ctx, candidates[2])
	bond1to2 := DelegatorBond{addrDels[0], addrVals[1], sdk.NewRat(9), 0}
	bond1to3 := DelegatorBond{addrDels[0], addrVals[2], sdk.NewRat(9), 1}
	bond2to1 := DelegatorBond{addrDels[1], addrVals[0], sdk.NewRat(9), 2}
	bond2to2 := DelegatorBond{addrDels[1], addrVals[1], sdk.NewRat(9), 3}
	bond2to3 := DelegatorBond{addrDels[1], addrVals[2], sdk.NewRat(9), 4}
	keeper.setDelegatorBond(ctx, bond1to2)
	keeper.setDelegatorBond(ctx, bond1to3)
	keeper.setDelegatorBond(ctx, bond2to1)
	keeper.setDelegatorBond(ctx, bond2to2)
	keeper.setDelegatorBond(ctx, bond2to3)

	// test all bond retrieve capabilities
	resBonds := viewSlashKeeper.GetDelegatorBonds(ctx, addrDels[0], 5)
	require.Equal(t, 3, len(resBonds))
	assert.True(t, bondsEqual(bond1to1, resBonds[0]))
	assert.True(t, bondsEqual(bond1to2, resBonds[1]))
	assert.True(t, bondsEqual(bond1to3, resBonds[2]))
	resBonds = viewSlashKeeper.GetDelegatorBonds(ctx, addrDels[0], 3)
	require.Equal(t, 3, len(resBonds))
	resBonds = viewSlashKeeper.GetDelegatorBonds(ctx, addrDels[0], 2)
	require.Equal(t, 2, len(resBonds))
	resBonds = viewSlashKeeper.GetDelegatorBonds(ctx, addrDels[1], 5)
	require.Equal(t, 3, len(resBonds))
	assert.True(t, bondsEqual(bond2to1, resBonds[0]))
	assert.True(t, bondsEqual(bond2to2, resBonds[1]))
	assert.True(t, bondsEqual(bond2to3, resBonds[2]))

}
