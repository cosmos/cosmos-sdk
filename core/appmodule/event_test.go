package appmodule

import (
	"context"
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestEventListenerRegistrar(t *testing.T) {
	registrar := &EventListenerRegistrar{}
	RegisterEventListener(registrar, func(ctx context.Context, dummy *timestamppb.Timestamp) {})
	RegisterEventInterceptor(registrar, func(ctx context.Context, dummy *structpb.Struct) error {
		return nil
	})
	require.Len(t, registrar.listeners, 2)
	require.Equal(t, reflect.Func, reflect.TypeOf(registrar.listeners[0]).Kind())
	require.Equal(t, reflect.Func, reflect.TypeOf(registrar.listeners[1]).Kind())
}
