package services

import (
	"context"

	corecontext "cosmossdk.io/core/context"
	"cosmossdk.io/core/transaction"
)

var _ transaction.Service = &ContextAwareTransactionService{}

type ContextAwareTransactionService struct{}

func (c ContextAwareTransactionService) ExecMode(ctx context.Context) transaction.ExecMode {
	return ctx.Value(corecontext.ExecModeKey).(transaction.ExecMode)
}

func NewContextAwareTransactionService() transaction.Service {
	return &ContextAwareTransactionService{}
}
