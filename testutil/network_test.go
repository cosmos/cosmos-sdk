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
	require.NotEmpty(t, n.Validators)
	client := n.Validators[0].RPCClient
	require.NotNil(t, client)

	ticker := time.NewTicker(5 * time.Second)
	timeout := time.After(time.Minute)

	for {
		select {
		case <-ticker.C:
			s, _ := client.Status()
			if s != nil && s.SyncInfo.LatestBlockHeight >= 10 {
				t.Logf("successfully process %d blocks", s.SyncInfo.LatestBlockHeight)
				return
			}
		case <-timeout:
			t.Fatal("timeout exceeded waiting for enough committed blocks")
		}
	}
}
