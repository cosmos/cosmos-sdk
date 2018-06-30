package types

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParamsEqual(t *testing.T) {
	p1 := DefaultParams()
	p2 := DefaultParams()

	ok := p1.Equal(p2)
	require.True(t, ok)

	p2.UnbondingTime = 60 * 60 * 24 * 2
	p2.BondDenom = "soup"

	ok = p1.Equal(p2)
	require.False(t, ok)
}
