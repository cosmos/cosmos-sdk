package stf

import (
	"context"
	"testing"
)

func TestGetExecutionCtxFromContext(t *testing.T) {
	t.Run("direct type *executionContext", func(t *testing.T) {
		ec := &executionContext{}
		result, err := getExecutionCtxFromContext(ec)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if result != ec {
			t.Fatalf("expected %v, got %v", ec, result)
		}
	})

	t.Run("context value of type *executionContext", func(t *testing.T) {
		ec := &executionContext{}
		ctx := context.WithValue(context.Background(), executionContextKey, ec)
		result, err := getExecutionCtxFromContext(ctx)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if result != ec {
			t.Fatalf("expected %v, got %v", ec, result)
		}
	})

	t.Run("invalid context type or value", func(t *testing.T) {
		ctx := context.Background()
		_, err := getExecutionCtxFromContext(ctx)
		if err == nil {
			t.Fatalf("expected error, got nil")
		}
		expectedErr := "failed to get executionContext from context"
		if err.Error() != expectedErr {
			t.Fatalf("expected error message %v, got %v", expectedErr, err.Error())
		}
	})
}
