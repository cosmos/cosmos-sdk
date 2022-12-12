package collections

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestNameRegex(t *testing.T) {
	require.Regexp(t, nameRegex, "a")
	require.Regexp(t, nameRegex, "foo1_xyz")
	require.NotRegexp(t, nameRegex, "1foo")
	require.NotRegexp(t, nameRegex, "_bar")
	require.NotRegexp(t, nameRegex, "abc-xyz")
}
