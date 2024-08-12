package stf

import (
	"context"
	"testing"

	gogoproto "github.com/cosmos/gogoproto/proto"
	gogotypes "github.com/cosmos/gogoproto/types"
	"github.com/stretchr/testify/require"

	"cosmossdk.io/core/appmodule/v2"
	transaction "cosmossdk.io/core/transaction"
)

func TestRouter(t *testing.T) {
	expectedMsg := &gogotypes.BoolValue{Value: true}
	expectedMsgName := gogoproto.MessageName(expectedMsg)

	expectedResp := &gogotypes.StringValue{Value: "test"}

	router := coreRouterImpl{handlers: map[string]appmodule.Handler{
		gogoproto.MessageName(expectedMsg): func(ctx context.Context, gotMsg transaction.Msg) (msgResp transaction.Msg, err error) {
			require.Equal(t, expectedMsg, gotMsg)
			return expectedResp, nil
		},
	}}

	t.Run("can invoke message by name", func(t *testing.T) {
		err := router.CanInvoke(context.Background(), expectedMsgName)
		require.NoError(t, err, "must be invocable")
	})

	t.Run("can invoke message by type URL", func(t *testing.T) {
		err := router.CanInvoke(context.Background(), "/"+expectedMsgName)
		require.NoError(t, err)
	})

	t.Run("cannot invoke unknown message", func(t *testing.T) {
		err := router.CanInvoke(context.Background(), "not exist")
		require.Error(t, err)
	})

	t.Run("invoke", func(t *testing.T) {
		gotResp, err := router.Invoke(context.Background(), expectedMsg)
		require.NoError(t, err)
		require.Equal(t, expectedResp, gotResp)
	})
}
