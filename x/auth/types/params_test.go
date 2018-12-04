package types

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParamsEqual(t *testing.T) {
	p1 := DefaultParams()
	p2 := DefaultParams()
	require.Equal(t, p1, p2)

	p1.TxSigLimit += 10
	require.NotEqual(t, p1, p2)
}
