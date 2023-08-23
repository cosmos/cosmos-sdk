package implementation

import (
	"context"

	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/wrapperspb"

	"cosmossdk.io/collections"
)

var _ Account = (*TestAccount)(nil)

var itemPrefix = collections.NewPrefix([]byte{0})

// NewTestAccount creates a new TestAccount.
func NewTestAccount(sb *collections.SchemaBuilder) (TestAccount, error) {
	ta := TestAccount{
		Item: collections.NewItem(sb, itemPrefix, "test", collections.Uint64Value),
	}
	return ta, nil
}

type TestAccount struct {
	Item collections.Item[uint64]
}

func (TestAccount) RegisterInitHandler(builder *InitBuilder) {
	RegisterInitHandler(builder, func(_ context.Context, req *wrapperspb.StringValue) (*wrapperspb.StringValue, error) {
		return &wrapperspb.StringValue{Value: req.Value + "init-echo"}, nil
	})
}

func (t TestAccount) RegisterExecuteHandlers(builder *ExecuteBuilder) {
	RegisterExecuteHandler(builder, func(_ context.Context, req *wrapperspb.StringValue) (*wrapperspb.StringValue, error) {
		return &wrapperspb.StringValue{Value: req.Value + "execute-echo"}, nil
	})

	RegisterExecuteHandler(builder, func(_ context.Context, req *wrapperspb.BytesValue) (*wrapperspb.BytesValue, error) {
		return &wrapperspb.BytesValue{Value: append(req.Value, "bytes-execute-echo"...)}, nil
	})

	// State tester
	RegisterExecuteHandler(builder, func(ctx context.Context, req *wrapperspb.UInt64Value) (*emptypb.Empty, error) {
		return &emptypb.Empty{}, t.Item.Set(ctx, req.Value)
	})
}

func (TestAccount) RegisterQueryHandlers(builder *QueryBuilder) {
	RegisterQueryHandler(builder, func(_ context.Context, req *wrapperspb.StringValue) (*wrapperspb.StringValue, error) {
		return &wrapperspb.StringValue{Value: req.Value + "query-echo"}, nil
	})
	RegisterQueryHandler(builder, func(_ context.Context, req *wrapperspb.BytesValue) (*wrapperspb.BytesValue, error) {
		return &wrapperspb.BytesValue{Value: append(req.Value, "bytes-query-echo"...)}, nil
	})
}
