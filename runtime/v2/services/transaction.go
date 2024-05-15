package services

import (
	"context"

	corecontext "cosmossdk.io/core/context"
	"cosmossdk.io/core/transaction"
)

var _ transaction.Service = &ContextAwareTransactionService{}

// ContextAwareTransactionService implements the transaction.Service interface.
// It is used to retrieve the execution mode in the context.
type ContextAwareTransactionService struct{}

// ExecMode returns the execution mode stored in the context.
func (c ContextAwareTransactionService) ExecMode(ctx context.Context) transaction.ExecMode {
	return ctx.Value(corecontext.ExecModeKey).(transaction.ExecMode)
}

func NewContextAwareTransactionService() transaction.Service {
	return &ContextAwareTransactionService{}
}
