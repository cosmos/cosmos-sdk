package testutil

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func AssertError(t *testing.T, err error, expectedErr error, expectedErrMsg string) {
	switch {
	case expectedErr != nil:
		require.ErrorIs(t, err, expectedErr)
	case expectedErrMsg != "":
		require.ErrorContainsf(t, err, expectedErrMsg, "expected error %s, got %v", expectedErrMsg, err)
	default:
		require.Error(t, err)
	}
}
