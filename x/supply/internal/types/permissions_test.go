package types

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestHasPermission(t *testing.T) {
	emptyPermAddr := NewPermAddr("empty", []string{})
	has := emptyPermAddr.HasPermission(Basic)
	require.False(t, has)

	cases := []struct {
		permission string
		expectHas  bool
	}{
		{Basic, true},
		{Minter, true},
		{Burner, true},
		{Staking, true},
		{"random", false},
		{"", false},
	}
	permAddr := NewPermAddr("test", []string{Basic, Minter, Burner, Staking})
	for i, tc := range cases {
		has = permAddr.HasPermission(tc.permission)
		require.Equal(t, tc.expectHas, has, "test case #%d", i)
	}

}

func TestValidatePermissions(t *testing.T) {
	cases := []struct {
		name        string
		permissions []string
		expectPass  bool
	}{
		{"no permissions", []string{}, true},
		{"valid permission", []string{Basic}, true},
		{"invalid permission", []string{""}, false},
		{"invalid and valid permission", []string{Basic, ""}, false},
	}

	for i, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := validatePermissions(tc.permissions...)
			if tc.expectPass {
				require.NoError(t, err, "test case #%d", i)
			} else {
				require.Error(t, err, "test case #%d", i)
			}
		})
	}
}
