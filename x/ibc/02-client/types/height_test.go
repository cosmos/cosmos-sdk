package types_test

import (
	"testing"

	"github.com/cosmos/cosmos-sdk/x/ibc/02-client/types"
	"github.com/stretchr/testify/require"
)

func TestZeroHeight(t *testing.T) {
	require.Equal(t, types.Height{}, types.ZeroHeight())
}

func TestCompareHeights(t *testing.T) {
	testCases := []struct {
		name        string
		height1     types.Height
		height2     types.Height
		compareSign int64
	}{
		{"epoch number 1 is lesser", types.NewHeight(1, 3), types.NewHeight(3, 4), -1},
		{"epoch number 1 is greater", types.NewHeight(7, 5), types.NewHeight(4, 5), 1},
		{"epoch height 1 is lesser", types.NewHeight(3, 4), types.NewHeight(3, 9), -1},
		{"epoch height 1 is greater", types.NewHeight(3, 8), types.NewHeight(3, 3), 1},
		{"height is equal", types.NewHeight(4, 4), types.NewHeight(4, 4), 0},
	}

	for i, tc := range testCases {
		compare := tc.height1.Compare(tc.height2)

		switch tc.compareSign {
		case -1:
			require.True(t, compare == -1, "case %d: %s should return negative value on comparison, got: %d",
				i, tc.name, compare)
		case 0:
			require.True(t, compare == 0, "case %d: %s should return zero on comparison, got: %d",
				i, tc.name, compare)
		case 1:
			require.True(t, compare == 1, "case %d: %s should return positive value on comparison, got: %d",
				i, tc.name, compare)
		}
	}
}

func TestDecrement(t *testing.T) {
	validDecrement := types.NewHeight(3, 3)
	expected := types.NewHeight(3, 2)

	actual, success := validDecrement.Decrement()
	require.Equal(t, expected, actual, "decrementing %s did not return expected height: %s. got %s",
		validDecrement, expected, actual)
	require.True(t, success, "decrement failed unexpectedly")

	invalidDecrement := types.NewHeight(3, 0)
	actual, success = invalidDecrement.Decrement()

	require.Equal(t, types.ZeroHeight(), actual, "invalid decrement returned non-zero height: %s", actual)
	require.False(t, success, "invalid decrement passed")
}

func TestString(t *testing.T) {
	_, err := types.ParseHeight("height")
	require.Error(t, err, "invalid height string passed")

	_, err = types.ParseHeight("epoch-10")
	require.Error(t, err, "invalid epoch string passed")

	_, err = types.ParseHeight("3-height")
	require.Error(t, err, "invalid epoch-height string passed")

	height := types.NewHeight(3, 4)
	recovered, err := types.ParseHeight(height.String())

	require.NoError(t, err, "valid height string could not be parsed")
	require.Equal(t, height, recovered, "recovered height not equal to original height")

	parse, err := types.ParseHeight("3-10")
	require.NoError(t, err, "parse err")
	require.Equal(t, types.NewHeight(3, 10), parse, "parse height returns wrong height")
}
