package types_test

import (
	"testing"

	clientexported "github.com/cosmos/cosmos-sdk/x/ibc/02-client/exported"
	"github.com/cosmos/cosmos-sdk/x/ibc/07-tendermint/types"
	"github.com/stretchr/testify/require"
)

func TestCompareHeights(t *testing.T) {
	testCases := []struct {
		name        string
		height1     types.Height
		height2     clientexported.Height
		compareSign int64
		expPass     bool
	}{
		{"epoch number 1 is lesser", types.NewHeight(1, 3), types.NewHeight(3, 4), -1, true},
		{"epoch number 1 is greater", types.NewHeight(7, 5), types.NewHeight(4, 5), 1, true},
		{"epoch height 1 is lesser", types.NewHeight(3, 4), types.NewHeight(3, 9), -1, true},
		{"epoch height 1 is greater", types.NewHeight(3, 8), types.NewHeight(3, 3), 1, true},
		{"height is equal", types.NewHeight(4, 4), types.NewHeight(4, 4), 0, true},
		{"other height is not tm height", types.NewHeight(5, 4), nil, 0, false},
	}

	for i, tc := range testCases {
		compare, err := tc.height1.Compare(tc.height2)

		if tc.expPass {
			require.Nil(t, err, "case %d: %s returned unexpected error: %v", i, tc.name, err)
		} else {
			require.NotNil(t, err, "case %d: %s did not error", i, tc.name)
		}

		switch tc.compareSign {
		case -1:
			require.True(t, compare < 0, "case %d: %s should return negative value on comparison, got: %d",
				i, tc.name, compare)
		case 0:
			require.True(t, compare == 0, "case %d: %s should return zero on comparison, got: %d",
				i, tc.name, compare)
		case 1:
			require.True(t, compare > 0, "case %d: %s should return positive value on comparison, got: %d",
				i, tc.name, compare)
		}
	}
}

func TestDecrement(t *testing.T) {
	validDecrement := types.NewHeight(3, 3)
	expected := types.NewHeight(3, 2)

	actual, err := validDecrement.Decrement()
	require.Equal(t, expected, actual, "decrementing %s did not return expected height: %s. got %s",
		validDecrement.String(), expected.String(), actual.String())
	require.Nil(t, err, "decrement returned unexpected error: %v", err)

	invalidDecrement := types.NewHeight(3, 1)
	actual, err = invalidDecrement.Decrement()

	require.Equal(t, types.Height{}, actual, "invalid decrement returned non-zero height: %s", actual.String())
	require.NotNil(t, err, "expected error on invalid decrement")
}

func TestValid(t *testing.T) {
	valid := types.NewHeight(0, 2)
	require.True(t, valid.Valid(), "valid height did not return true on Valid()")

	invalid := types.NewHeight(2, 0)
	require.False(t, invalid.Valid(), "invalid height returned true on Valid()")
}
