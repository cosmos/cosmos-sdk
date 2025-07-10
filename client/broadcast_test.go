package client

import (
	"context"
	"fmt"
	"testing"

	"github.com/cometbft/cometbft/v2/crypto/tmhash"
	"github.com/cometbft/cometbft/v2/mempool"
	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/client/flags"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

func CreateContextWithErrorAndMode(err error, mode string) Context {
	return Context{
		Client:        MockClient{err: err},
		BroadcastMode: mode,
	}
}

// Test the correct code is returned when
func TestBroadcastError(t *testing.T) {
	errors := map[error]uint32{
		mempool.ErrTxInCache:       sdkerrors.ErrTxInMempoolCache.ABCICode(),
		mempool.ErrTxTooLarge{}:    sdkerrors.ErrTxTooLarge.ABCICode(),
		mempool.ErrMempoolIsFull{}: sdkerrors.ErrMempoolIsFull.ABCICode(),
	}

	modes := []string{
		flags.BroadcastAsync,
		flags.BroadcastSync,
	}

	txBytes := []byte{0xA, 0xB}
	txHash := fmt.Sprintf("%X", tmhash.Sum(txBytes))

	for _, mode := range modes {
		for err, code := range errors {
			ctx := CreateContextWithErrorAndMode(err, mode)
			resp, returnedErr := ctx.BroadcastTx(txBytes)
			require.NoError(t, returnedErr)
			require.Equal(t, code, resp.Code)
			require.NotEmpty(t, resp.Codespace)
			require.Equal(t, txHash, resp.TxHash)
		}
	}
}

func TestBroadcastCancellation(t *testing.T) {
	modes := []string{
		flags.BroadcastAsync,
		flags.BroadcastSync,
	}

	txBytes := []byte{0xA, 0xB}
	cmdCtx, cancel := context.WithCancel(context.Background())
	cancel()

	for _, mode := range modes {
		ctx := CreateContextWithErrorAndMode(nil, mode).WithCmdContext(cmdCtx)
		_, err := ctx.BroadcastTx(txBytes)
		require.ErrorIs(t, err, context.Canceled)
	}
}

func BenchmarkCheckCometErrorCompare(b *testing.B) {
	txBytes := []byte{0xA, 0xB}

	errs := []error{
		mempool.ErrTxInCache,
		mempool.ErrTxTooLarge{},
		mempool.ErrMempoolIsFull{},
	}

	for _, err := range errs {
		errName := fmt.Sprintf("%T", err)

		b.Run(errName, func(b *testing.B) {
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				if CheckCometError(err, txBytes) == nil {
					b.FailNow()
				}
			}
		})
	}
}
