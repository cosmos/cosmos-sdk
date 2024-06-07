package v6

import (
	"context"
)

type migrateAccNumFunc = func(ctx context.Context) error

// Migrate account number from x/auth account number to x/accounts account number
func Migrate(ctx context.Context, f migrateAccNumFunc) error {
	return f(ctx)
}
