package types

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestResult(t *testing.T) {
	var res Result
	require.True(t, res.IsOK())

	res.Data = []byte("data")
	require.True(t, res.IsOK())

	res.Code = CodeType(1)
	require.False(t, res.IsOK())
}
