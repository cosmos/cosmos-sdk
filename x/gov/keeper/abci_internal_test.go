package keeper

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/runtime/protoiface"

	"cosmossdk.io/core/router"
)

type mockRouterService struct {
	router.Service

	panic bool
}

func (m *mockRouterService) InvokeUntyped(ctx context.Context, req protoiface.MessageV1) (res protoiface.MessageV1, err error) {
	if m.panic {
		panic("test-fail")
	}

	return nil, nil
}

func TestSafeExecuteHandler(t *testing.T) {
	t.Parallel()

	require := require.New(t)
	ctx := context.Background()

	r, err := safeExecuteHandler(ctx, nil, &mockRouterService{panic: true})
	require.ErrorContains(err, "test-fail")
	require.Nil(r)

	_, err = safeExecuteHandler(ctx, nil, &mockRouterService{panic: false})
	require.Nil(err)
}
