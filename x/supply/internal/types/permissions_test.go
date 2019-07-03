package types

import (
	"testing"

	"github.com/stretchr/testify/require"
)

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
