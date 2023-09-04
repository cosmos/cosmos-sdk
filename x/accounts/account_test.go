package accounts

import (
	"context"
	"strconv"

	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/wrapperspb"

	"cosmossdk.io/x/accounts/internal/implementation"
)

var _ implementation.Account = (*TestAccount)(nil)

type TestAccount struct{}

func (t TestAccount) RegisterInitHandler(builder *implementation.InitBuilder) {
	implementation.RegisterInitHandler(builder, func(_ context.Context, _ *emptypb.Empty) (*emptypb.Empty, error) {
		return &emptypb.Empty{}, nil
	})
}

func (t TestAccount) RegisterExecuteHandlers(builder *implementation.ExecuteBuilder) {
	implementation.RegisterExecuteHandler(builder, func(_ context.Context, _ *emptypb.Empty) (*emptypb.Empty, error) {
		return &emptypb.Empty{}, nil
	})

	implementation.RegisterExecuteHandler(builder, func(_ context.Context, req *wrapperspb.StringValue) (*wrapperspb.UInt64Value, error) {
		value, err := strconv.ParseUint(req.Value, 10, 64)
		if err != nil {
			return nil, err
		}

		return wrapperspb.UInt64(value), nil
	})
}

func (t TestAccount) RegisterQueryHandlers(builder *implementation.QueryBuilder) {
	implementation.RegisterQueryHandler(builder, func(_ context.Context, _ *emptypb.Empty) (*emptypb.Empty, error) {
		return &emptypb.Empty{}, nil
	})

	implementation.RegisterQueryHandler(builder, func(_ context.Context, req *wrapperspb.UInt64Value) (*wrapperspb.StringValue, error) {
		return wrapperspb.String(strconv.FormatUint(req.Value, 10)), nil
	})
}
