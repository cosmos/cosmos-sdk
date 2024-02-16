package implementation

import (
	"context"

	"github.com/cosmos/gogoproto/types"

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
	RegisterInitHandler(builder, func(_ context.Context, req *types.StringValue) (*types.StringValue, error) {
		return &types.StringValue{Value: req.Value + "init-echo"}, nil
	})
}

func (t TestAccount) RegisterExecuteHandlers(builder *ExecuteBuilder) {
	RegisterExecuteHandler(builder, func(_ context.Context, req *types.StringValue) (*types.StringValue, error) {
		return &types.StringValue{Value: req.Value + "execute-echo"}, nil
	})

	RegisterExecuteHandler(builder, func(_ context.Context, req *types.BytesValue) (*types.BytesValue, error) {
		return &types.BytesValue{Value: append(req.Value, "bytes-execute-echo"...)}, nil
	})

	// State tester
	RegisterExecuteHandler(builder, func(ctx context.Context, req *types.UInt64Value) (*types.Empty, error) {
		return &types.Empty{}, t.Item.Set(ctx, req.Value)
	})
}

func (t TestAccount) RegisterQueryHandlers(builder *QueryBuilder) {
	RegisterQueryHandler(builder, func(_ context.Context, req *types.StringValue) (*types.StringValue, error) {
		return &types.StringValue{Value: req.Value + "query-echo"}, nil
	})
	RegisterQueryHandler(builder, func(_ context.Context, req *types.BytesValue) (*types.BytesValue, error) {
		return &types.BytesValue{Value: append(req.Value, "bytes-query-echo"...)}, nil
	})
}
