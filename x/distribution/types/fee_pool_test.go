package types

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
)

func TestTotalAccumUpdateForNewHeight(t *testing.T) {

	ta := NewTotalAccum(0)

	ta = ta.UpdateForNewHeight(5, sdk.NewDec(3))
	require.True(sdk.DecEq(t, sdk.NewDec(15), ta.Accum))

	ta = ta.UpdateForNewHeight(8, sdk.NewDec(2))
	require.True(sdk.DecEq(t, sdk.NewDec(21), ta.Accum))
}

func TestUpdateTotalValAccum(t *testing.T) {

	fp := InitialFeePool()

	fp = fp.UpdateTotalValAccum(5, sdk.NewDec(3))
	require.True(sdk.DecEq(t, sdk.NewDec(15), fp.TotalValAccum.Accum))

	fp = fp.UpdateTotalValAccum(8, sdk.NewDec(2))
	require.True(sdk.DecEq(t, sdk.NewDec(21), fp.TotalValAccum.Accum))
}
