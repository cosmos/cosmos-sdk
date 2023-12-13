package unorderedtx_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"cosmossdk.io/x/auth/ante/unorderedtx"
)

func TestUnorderedTxManager_Close(t *testing.T) {
	txm := unorderedtx.NewManager()
	txm.Start()

	require.NoError(t, txm.Close())
	require.Panics(t, func() { txm.Close() })
}

func TestUnorderedTxManager_SimpleSize(t *testing.T) {
	txm := unorderedtx.NewManager()
	defer txm.Close()

	txm.Start()

	txm.Add([32]byte{0xFF}, 100)
	txm.Add([32]byte{0xAA}, 100)
	txm.Add([32]byte{0xCC}, 100)

	require.Equal(t, 3, txm.Size())
}

func TestUnorderedTxManager_SimpleContains(t *testing.T) {
	txm := unorderedtx.NewManager()
	defer txm.Close()

	txm.Start()

	for i := 0; i < 10; i++ {
		txHash := [32]byte{byte(i)}
		txm.Add(txHash, 100)
		require.True(t, txm.Contains(txHash))
	}

	for i := 10; i < 20; i++ {
		txHash := [32]byte{byte(i)}
		require.False(t, txm.Contains(txHash))
	}
}

func TestUnorderedTxManager_Flow(t *testing.T) {
	txm := unorderedtx.NewManager()
	defer txm.Close()

	txm.Start()

	// Seed the manager with a txs, some of which should eventually be purged and
	// the others will remain. Txs with TTL less than or equal to 50 should be purged.
	for i := 1; i <= 100; i++ {
		txHash := [32]byte{byte(i)}

		if i <= 50 {
			txm.Add(txHash, uint64(i))
		} else {
			txm.Add(txHash, 100)
		}
	}

	// start a goroutine that mimics new blocks being made every 500ms
	doneBlockCh := make(chan bool)
	go func() {
		ticker := time.NewTicker(time.Millisecond * 500)
		defer ticker.Stop()

		var (
			height uint64 = 1
			i             = 101
		)
		for range ticker.C {
			txm.OnNewBlock(height)
			height++

			if height > 51 {
				doneBlockCh <- true
				return
			} else {
				txm.Add([32]byte{byte(i)}, 50)
			}
		}
	}()

	// Eventually all the txs that should be expired by block 50 should be purged.
	// The remaining txs should remain.
	require.Eventually(
		t,
		func() bool {
			return txm.Size() == 50
		},
		2*time.Minute,
		5*time.Second,
	)

	<-doneBlockCh
}
