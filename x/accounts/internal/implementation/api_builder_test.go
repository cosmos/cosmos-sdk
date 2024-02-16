package implementation

import (
	"context"
	"testing"

	"github.com/cosmos/gogoproto/types"
	"github.com/stretchr/testify/require"
)

func TestRouterDoubleRegistration(t *testing.T) {
	router := NewExecuteBuilder()
	RegisterExecuteHandler(router, func(_ context.Context, req *types.StringValue) (*types.StringValue, error) { return nil, nil })
	RegisterExecuteHandler(router, func(_ context.Context, req *types.StringValue) (*types.StringValue, error) { return nil, nil })

	_, err := router.makeHandler()
	require.ErrorContains(t, err, "already registered")
}

func TestEmptyQueryExecuteHandler(t *testing.T) {
	qr := NewQueryBuilder()
	er := NewExecuteBuilder()

	qh, err := qr.makeHandler()
	require.NoError(t, err)
	eh, err := er.makeHandler()
	require.NoError(t, err)

	ctx := context.Background()

	_, err = qh(ctx, &types.StringValue{})
	require.ErrorIs(t, err, errNoExecuteHandler)
	_, err = eh(ctx, &types.StringValue{})
	require.ErrorIs(t, err, errNoExecuteHandler)
}
