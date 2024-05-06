package v4

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

	// Set LastReductionEpoch to default value `0`
	if lre != 0 {
		lre = 0
	}

	// Initialize the LastReductionEpoch with the default value
	if err := lastReductionEpoch.Set(ctx, lre); err != nil {
		return err
	}

	return nil
}
