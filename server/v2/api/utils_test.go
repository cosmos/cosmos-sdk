package api

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestDoUntilCtxExpired(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		ctx := context.Background()

		funcRan := false
		err := DoUntilCtxExpired(ctx, func() {
			funcRan = true
		})
		require.NoError(t, err)
		require.True(t, funcRan)
	})

	t.Run("context expired", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())

		funcRan := false
		err := DoUntilCtxExpired(ctx, func() {
			cancel()
			funcRan = true
			<-time.After(time.Second)
		})
		require.ErrorIs(t, err, context.Canceled)
		require.True(t, funcRan)
	})
}
