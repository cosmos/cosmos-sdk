package accounts

import "context"

type AuthKeeper interface {
	NextAccountNumber(ctx context.Context) uint64
	CurrentAccountNumber(ctx context.Context) (uint64, error)
}
