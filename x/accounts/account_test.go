package accounts

import (
	"context"

	"google.golang.org/protobuf/types/known/emptypb"

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
}

func (t TestAccount) RegisterQueryHandlers(builder *implementation.QueryBuilder) {
	implementation.RegisterQueryHandler(builder, func(_ context.Context, _ *emptypb.Empty) (*emptypb.Empty, error) {
		return &emptypb.Empty{}, nil
	})
}
