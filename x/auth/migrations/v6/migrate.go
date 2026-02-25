package v6

import (
	"context"
	"cosmossdk.io/collections"
	storetypes "cosmossdk.io/core/store"
	"errors"
)

func Migrate(ctx context.Context, storeService storetypes.KVStoreService, sequence collections.Sequence) error {
	_, err := sequence.Peek(ctx)
	if err != nil && !errors.Is(err, collections.ErrNotFound) {
		return err
	}

	// remove the global account number.
	err = (collections.Item[uint64])(sequence).Remove(ctx)
	if err != nil {
		return err
	}

	return nil
}
