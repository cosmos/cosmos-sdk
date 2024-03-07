package keeper

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/runtime/protoiface"

	"cosmossdk.io/core/router"
)

type mockRouter struct {
	router.Router

	panic bool
}

func (m *mockRouter) InvokeUntyped(ctx context.Context, req protoiface.MessageV1) (res protoiface.MessageV1, err error) {
	if m.panic {
		panic("test-fail")
	}

	return nil, nil
}

func TestSafeExecuteHandler(t *testing.T) {
	t.Parallel()

	require := require.New(t)
	ctx := context.Background()

	r, err := safeExecuteHandler(ctx, nil, &mockRouter{panic: true})
	require.ErrorContains(err, "test-fail")
	require.Nil(r)

	_, err = safeExecuteHandler(ctx, nil, &mockRouter{panic: false})
	require.Nil(err)
}
