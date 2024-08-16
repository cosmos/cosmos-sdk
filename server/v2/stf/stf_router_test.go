package stf

import (
	"context"
	"testing"

	gogoproto "github.com/cosmos/gogoproto/proto"
	gogotypes "github.com/cosmos/gogoproto/types"
	"github.com/stretchr/testify/assert"
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

	t.Run("invoke untyped", func(t *testing.T) {
		gotResp, err := router.InvokeUntyped(context.Background(), expectedMsg)
		require.NoError(t, err)
		require.Equal(t, expectedResp, gotResp)
	})

	t.Run("invoked typed", func(t *testing.T) {
		gotResp := new(gogotypes.StringValue)
		err := router.InvokeTyped(context.Background(), expectedMsg, gotResp)
		require.NoError(t, err)
		require.Equal(t, expectedResp, gotResp)
	})
}

func TestMerge(t *testing.T) {
	tests := []struct {
		name     string
		src      transaction.Msg
		dst      transaction.Msg
		expected transaction.Msg
		wantErr  bool
	}{
		{
			name:     "success",
			src:      &gogotypes.BoolValue{Value: true},
			dst:      &gogotypes.BoolValue{},
			expected: &gogotypes.BoolValue{Value: true},
			wantErr:  false,
		},
		{
			name:     "nil src",
			src:      nil,
			dst:      &gogotypes.StringValue{},
			expected: &gogotypes.StringValue{},
			wantErr:  true,
		},
		{
			name:     "nil dst",
			src:      &gogotypes.StringValue{Value: "hello"},
			dst:      nil,
			expected: nil,
			wantErr:  true,
		},
		{
			name:     "incompatible types",
			src:      &gogotypes.StringValue{Value: "hello"},
			dst:      &gogotypes.BoolValue{},
			expected: &gogotypes.BoolValue{},
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := merge(tt.src, tt.dst)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, tt.dst)
			}
		})
	}
}
