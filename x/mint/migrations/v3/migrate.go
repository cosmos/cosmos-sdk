package v3

import (
	"context"
	"fmt"

	"cosmossdk.io/collections"
)

func MigrateStore(ctx context.Context, lastReductionEpoch collections.Item[int64]) error {
	lre, err := lastReductionEpoch.Get(ctx)
	if err != nil {
		return fmt.Errorf("failed to get last reduction epoch: %w", err)
	}
	return lastReductionEpoch.Set(ctx, lre)
}
