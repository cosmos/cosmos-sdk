package context

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCLIContext_WithOffline(t *testing.T) {
	ctx := NewCLIContext()
	require.False(t, ctx.Offline)

	ctx = ctx.WithOffline(true)
	require.True(t, ctx.Offline)
}
