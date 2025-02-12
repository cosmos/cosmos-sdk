package coretesting

import (
	"context"

	"cosmossdk.io/core/branch"
)

var _ branch.Service = &TestBranchService{}

type TestBranchService struct{}

func (bs TestBranchService) Execute(ctx context.Context, f func(ctx context.Context) error) error {
	unwrap(ctx) // check that this is a testing context
	return f(ctx)
}

func (bs TestBranchService) ExecuteWithGasLimit(
	ctx context.Context,
	gasLimit uint64,
	f func(ctx context.Context) error,
) (gasUsed uint64, err error) {
	unwrap(ctx) // check that this is a testing context
	return gasLimit, f(ctx)
}
