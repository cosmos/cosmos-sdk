package implementation

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

func TestRouterDoubleRegistration(t *testing.T) {
	router := NewExecuteBuilder()
	RegisterExecuteHandler(router, func(_ context.Context, req *wrapperspb.StringValue) (*wrapperspb.StringValue, error) { return nil, nil })
	RegisterExecuteHandler(router, func(_ context.Context, req *wrapperspb.StringValue) (*wrapperspb.StringValue, error) { return nil, nil })

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

	_, err = qh(ctx, &wrapperspb.StringValue{})
	require.ErrorIs(t, err, errNoExecuteHandler)
	_, err = eh(ctx, &wrapperspb.StringValue{})
	require.ErrorIs(t, err, errNoExecuteHandler)
}
