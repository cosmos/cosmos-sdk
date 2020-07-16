package exported

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestClientTypeString(t *testing.T) {
	cases := []struct {
		msg        string
		name       string
		clientType ClientType
	}{
		{"tendermint client", ClientTypeTendermint, Tendermint},
		{"empty type", "", 0},
	}

	for _, tt := range cases {
		tt := tt
		require.Equal(t, tt.clientType, ClientTypeFromString(tt.name), tt.msg)
		require.Equal(t, tt.name, tt.clientType.String(), tt.msg)
	}
}

func TestClientTypeMarshalJSON(t *testing.T) {
	cases := []struct {
		msg        string
		name       string
		clientType ClientType
		expectPass bool
	}{
		{"tendermint client should have passed", ClientTypeTendermint, Tendermint, true},
		{"empty type should have failed", "", 0, false},
	}

	for _, tt := range cases {
		tt := tt
		bz, err := tt.clientType.MarshalJSON()
		require.NoError(t, err)
		var ct ClientType
		if tt.expectPass {
			require.NoError(t, ct.UnmarshalJSON(bz), tt.msg)
			require.Equal(t, tt.name, ct.String(), tt.msg)
		} else {
			require.Error(t, ct.UnmarshalJSON(bz), tt.msg)
		}
	}
}

func TestCompareHeights(t *testing.T) {
	testCases := []struct {
		name        string
		height1     Height
		height2     Height
		compareSign int64
	}{
		{"epoch number 1 is lesser", NewHeight(1, 3), NewHeight(3, 4), -1},
		{"epoch number 1 is greater", NewHeight(7, 5), NewHeight(4, 5), 1},
		{"epoch height 1 is lesser", NewHeight(3, 4), NewHeight(3, 9), -1},
		{"epoch height 1 is greater", NewHeight(3, 8), NewHeight(3, 3), 1},
		{"height is equal", NewHeight(4, 4), NewHeight(4, 4), 0},
	}

	for i, tc := range testCases {
		compare := tc.height1.Compare(tc.height2)

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
	validDecrement := NewHeight(3, 3)
	expected := NewHeight(3, 2)

	actual, success := validDecrement.Decrement()
	require.Equal(t, expected, actual, "decrementing %s did not return expected height: %s. got %s",
		validDecrement.String(), expected.String(), actual.String())
	require.True(t, success, "decrement failed unexpectedly")

	invalidDecrement := NewHeight(3, 1)
	actual, success = invalidDecrement.Decrement()

	require.Equal(t, Height{}, actual, "invalid decrement returned non-zero height: %s", actual.String())
	require.False(t, success, "invalid decrement passed")
}

func TestValid(t *testing.T) {
	valid := NewHeight(0, 2)
	require.True(t, valid.Valid(), "valid height did not return true on Valid()")

	invalid := NewHeight(2, 0)
	require.False(t, invalid.Valid(), "invalid height returned true on Valid()")
}
