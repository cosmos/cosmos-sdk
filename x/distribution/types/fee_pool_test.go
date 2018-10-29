package types

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
)

func TestUpdateTotalValAccum(t *testing.T) {

	fp := InitialFeePool()

	fp = fp.UpdateTotalValAccum(5, sdk.NewDec(3))
	require.True(sdk.DecEq(t, sdk.NewDec(15), fp.TotalValAccum.Accum))

	fp = fp.UpdateTotalValAccum(8, sdk.NewDec(2))
	require.True(sdk.DecEq(t, sdk.NewDec(21), fp.TotalValAccum.Accum))
}
