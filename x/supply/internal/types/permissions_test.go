package types

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRemovePermissions(t *testing.T) {
	name := "test"
	permAddr := NewPermAddr(name, []string{})
	require.Empty(t, permAddr.GetPermissions())

	permAddr.AddPermissions(Basic, Minter, Burner)
	require.Equal(t, []string{Basic, Minter, Burner}, permAddr.GetPermissions(), "did not add permissions")

	err := permAddr.RemovePermission("random")
	require.Error(t, err, "did not error on removing nonexistent permission")

	err = permAddr.RemovePermission(Burner)
	require.NoError(t, err, "failed to remove permission")
	require.Equal(t, []string{Basic, Minter}, permAddr.GetPermissions(), "does not have correct permissions")

	err = permAddr.RemovePermission(Basic)
	require.NoError(t, err, "failed to remove permission")
	require.Equal(t, []string{Minter}, permAddr.GetPermissions(), "does not have correct permissions")
}

func TestValidatePermissions(t *testing.T) {
	cases := []struct {
		name        string
		permissions []string
		expectPass  bool
	}{
		{"no permissions", []string{}, true},
		{"one permission", []string{Basic}, true},
		{"multiple permissions", []string{Basic, Minter, Burner}, true},
		{"invalid permission", []string{"other"}, false},
		{"multiple invalid permissions", []string{Burner, "other", "invalid"}, false},
	}

	for i, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := validatePermissions(tc.permissions)
			if tc.expectPass {
				require.NoError(t, err, "test case #%d", i)
			} else {
				require.Error(t, err, "test case #%d", i)
			}
		})
	}
}
