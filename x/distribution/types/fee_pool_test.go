package types

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
)

func TestTotalAccumUpdate(t *testing.T) {

	ta := NewTotalAccum(0)

	ta.Update(5, sdk.NewDec(3))
	require.True(DecEq(t, sdk.NewDec(15), ta.Accum))

	ta.Update(8, sdk.NewDec(2))
	require.True(DecEq(t, sdk.NewDec(21), ta.Accum))
}
