package stf

import (
	"context"
	"reflect"
	"testing"

	gogoproto "github.com/cosmos/gogoproto/proto"
	gogotypes "github.com/cosmos/gogoproto/types"

	appmodulev2 "cosmossdk.io/core/appmodule/v2"
	"cosmossdk.io/core/transaction"
)

func TestRouter(t *testing.T) {
	expectedMsg := &gogotypes.BoolValue{Value: true}
	expectedMsgName := gogoproto.MessageName(expectedMsg)

	expectedResp := &gogotypes.StringValue{Value: "test"}

	router := coreRouterImpl{handlers: map[string]appmodulev2.HandlerFunc{
		gogoproto.MessageName(expectedMsg): func(ctx context.Context, gotMsg transaction.Msg) (msgResp transaction.Msg, err error) {
			if !reflect.DeepEqual(expectedMsg, gotMsg) {
				t.Errorf("expected message: %v, got: %v", expectedMsg, gotMsg)
			}
			return expectedResp, nil
		},
	}}

	t.Run("can invoke message by name", func(t *testing.T) {
		err := router.CanInvoke(context.Background(), expectedMsgName)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("can invoke message by type URL", func(t *testing.T) {
		err := router.CanInvoke(context.Background(), "/"+expectedMsgName)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("cannot invoke unknown message", func(t *testing.T) {
		err := router.CanInvoke(context.Background(), "not exist")
		if err == nil {
			t.Error("expected error, got nil")
		}
	})

	t.Run("invoke", func(t *testing.T) {
		gotResp, err := router.Invoke(context.Background(), expectedMsg)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if !reflect.DeepEqual(expectedResp, gotResp) {
			t.Errorf("expected response: %v, got: %v", expectedResp, gotResp)
		}
	})
}
