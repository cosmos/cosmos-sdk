package types

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestConvertIntArrayToInt64(t *testing.T) {
	cases := map[string]struct {
		input	[]int
		expect []int64
	}{
		"with int array": {
			input: []int{1, 2, 3},
			expect: []int64{1, 2, 3},
		},
		"with empty array": {
			input: []int{},
			expect: []int64{},
		},
	}

	for name, tc := range cases {
		tc := tc // copy to local variable for scopelint
		t.Run(name, func(t *testing.T) {
			s := ConvertIntArrayToInt64(tc.input)
			require.Equal(t, tc.expect, s)
		})
	}

	t.Run("not equal", func(t *testing.T) {
		s := ConvertIntArrayToInt64([]int{1, 2, 3})
		require.NotEqual(t, []int64{1, 2, 4}, s)
	})
}

func TestContains(t *testing.T) {
	cases := map[string]struct {
		input	[]int64
		testInput int64
		expect bool
	}{
		"with element in array": {
			input: []int64{1, 2, 3},
			testInput: 1,
			expect: true,
		},
		"with element not in array": {
			input: []int64{1, 2, 3},
			testInput: 4,
			expect: false,
		},
		"with empty array": {
			input: []int64{},
			testInput: 1,
			expect: false,
		},
	}

	for name, tc := range cases {
		tc := tc // copy to local variable for scopelint
		t.Run(name, func(t *testing.T) {
			s := Contains(tc.input, tc.testInput)
			require.Equal(t, tc.expect, s)
		})
	}
}