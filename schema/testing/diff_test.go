package schematesting

import (
	"testing"

	"cosmossdk.io/schema"
)

func TestCompareKindValues(t *testing.T) {
	tt := []struct {
		kind        schema.Kind
		expected    any
		actual      any
		equal       bool
		expectError bool
	}{
		{
			kind:        schema.BoolKind,
			expected:    "true",
			actual:      false,
			expectError: true,
		},
		{
			kind:        schema.BoolKind,
			expected:    true,
			actual:      "false",
			expectError: true,
		},
		{
			kind:     schema.BoolKind,
			expected: true,
			actual:   false,
			equal:    false,
		},
		{
			kind:     schema.BoolKind,
			expected: true,
			actual:   true,
			equal:    true,
		},
		{
			kind:     schema.BytesKind,
			expected: []byte("hello"),
			actual:   []byte("world"),
			equal:    false,
		},
		{
			kind:        schema.IntegerStringKind,
			expected:    "a123",
			actual:      "123",
			expectError: true,
		},
		{
			kind:        schema.IntegerStringKind,
			expected:    "123",
			actual:      "123b",
			expectError: true,
		},
		{
			kind:     schema.IntegerStringKind,
			expected: "123",
			actual:   "1234",
			equal:    false,
		},
		{
			kind:     schema.IntegerStringKind,
			expected: "000123",
			actual:   "123",
			equal:    true,
		},
		{
			kind:        schema.DecimalStringKind,
			expected:    "abc",
			actual:      "100.001",
			expectError: true,
		},
		{
			kind:        schema.DecimalStringKind,
			expected:    "1",
			actual:      "b",
			expectError: true,
		},
		{
			kind:     schema.DecimalStringKind,
			expected: "1.00001",
			actual:   "100.001",
			equal:    false,
		},
		{
			kind:     schema.DecimalStringKind,
			expected: "1.00001e2",
			actual:   "100.001",
			equal:    true,
		},
		{
			kind:     schema.DecimalStringKind,
			expected: "00000100.00100000",
			actual:   "100.001",
			equal:    true,
		},
	}
	for _, tc := range tt {
		eq, err := CompareKindValues(tc.kind, tc.expected, tc.actual)
		if eq != tc.equal {
			t.Errorf("expected %v, got %v", tc.equal, eq)
		}
		if (err != nil) != tc.expectError {
			t.Errorf("expected error: %v, got %v", tc.expectError, err)
		}
	}
}
