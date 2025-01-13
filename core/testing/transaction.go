package coretesting

import (
	"context"

	"cosmossdk.io/core/transaction"
)

var _ transaction.Service = &TestTransactionService{}

type TestTransactionService struct{}

func (m TestTransactionService) ExecMode(ctx context.Context) transaction.ExecMode {
	dummy := unwrap(ctx)

	return dummy.execMode
}
