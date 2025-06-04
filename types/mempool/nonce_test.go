package mempool_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/types/mempool"
)

func TestChooseNonce(t *testing.T) {
	testCases := []struct {
		name      string
		seq       uint64
		unordered bool
		timeout   time.Time
		expErr    string
		expNonce  int64
	}{
		{
			name:      "unordered nonce chosen",
			unordered: true,
			timeout:   time.Unix(100, 15),
			expNonce:  time.Unix(100, 15).UnixNano(),
		},
		{
			name:     "sequence chosen",
			seq:      15,
			expNonce: 15,
		},
		{
			name:      "timeout invalid",
			unordered: true,
			timeout:   time.Time{},
			expErr:    "invalid timestamp value",
		},
		{
			name:      "invalid if sequence and unordered set",
			unordered: true,
			seq:       15,
			expErr:    "unordered txs must not have sequence set",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tx := testTx{unordered: tc.unordered, nonce: tc.seq, timeout: &tc.timeout}
			nonce, err := mempool.ChooseNonce(tc.seq, tx)
			if tc.expErr != "" {
				require.ErrorContains(t, err, tc.expErr)
			} else {
				require.NoError(t, err)
				require.Equal(t, nonce, uint64(tc.expNonce))
			}
		})
	}
}
