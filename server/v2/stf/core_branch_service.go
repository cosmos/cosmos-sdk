package stf

import (
	"context"

	"cosmossdk.io/core/branch"
	"cosmossdk.io/core/store"
)

type branchFn func(state store.ReaderMap) store.WriterMap

var _ branch.Service = (*BranchService)(nil)

type BranchService struct{}

func (bs BranchService) Execute(ctx context.Context, f func(ctx context.Context) error) error {
	// todo
	return nil
}

func (bs BranchService) ExecuteWithGasLimit(ctx context.Context, gasLimit uint64, f func(ctx context.Context) error) (gasUsed uint64, err error) {
	// todo
	return 0, nil
}
