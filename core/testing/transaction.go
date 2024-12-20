package coretesting

import (
	"context"
	"cosmossdk.io/core/transaction"
)

var _ transaction.Service = &MemTransactionService{}

type MemTransactionService struct{}

func (m MemTransactionService) ExecMode(ctx context.Context) transaction.ExecMode {
	dummy := unwrap(ctx)

	return dummy.execMode
}
