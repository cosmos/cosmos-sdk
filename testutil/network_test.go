package testutil

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestNetwork_Liveness(t *testing.T) {
	n := NewTestNetwork(t, DefaultConfig())
	defer n.Cleanup()
	require.NotNil(t, n)

	h, err := n.WaitForHeightWithTimeout(10, time.Minute)
	require.NoError(t, err, "expected to reach 10 blocks; got %d", h)
}
