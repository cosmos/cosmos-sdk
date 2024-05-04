package mock

import (
	"context"

	"cosmossdk.io/core/container"
	"cosmossdk.io/core/store"
)

// Map represents the basic collections object.
// It is used to map arbitrary keys to arbitrary
// objects.
type MockMap struct {
	// store accessor
	sa     func(context.Context) store.KVStore
	prefix []byte
	name   string
}

// MockItem is a type declaration based on MockMap
// without encode/decode so saving raw bytes in store
type MockItem struct {
	m            MockMap
	getContainer func(ctx context.Context) container.Service
}

// NewMockItem instantiates a new MockItem instance for  types,
// given the store service.
func NewMockItem(
	storeService store.KVStoreService,
	prefix []byte,
	name string,
) MockItem {
	m := MockMap{
		sa:     storeService.OpenKVStore,
		prefix: prefix,
		name:   name,
	}
	item := MockItem{
		m: m,
	}

	kl, ok := storeService.(interface {
		OpenContainer(ctx context.Context) container.Service
	})
	if ok {
		item.getContainer = kl.OpenContainer
	}

	return item
}

// Get gets the item, if it is not set it returns an ErrNotFound error.
// If value decoding fails then an ErrEncoding is returned.
func (i MockItem) Get(ctx context.Context) ([]byte, error) {
	var toCache bool
	if i.getContainer != nil {
		cached, found := i.getContainer(ctx).Get(i.m.prefix)
		if found {
			return cached.([]byte), nil
		} else {
			toCache = true
		}
	}
	kvStore := i.m.sa(ctx)
	v, err := kvStore.Get(i.m.prefix)
	if err == nil && toCache {
		i.getContainer(ctx).Set(i.m.prefix, v)
	}
	return v, err
}

// Set sets the item in the store. If Value encoding fails then an ErrEncoding is returned.
func (i MockItem) Set(ctx context.Context, value []byte) error {
	kvStore := i.m.sa(ctx)
	err := kvStore.Set(i.m.prefix, value)
	if err != nil {
		return err
	}
	if i.getContainer != nil {
		i.getContainer(ctx).Set(i.m.prefix, value)
	}
	return nil
}

// Has reports whether the item exists in the store or not.
// Returns an error in case encoding fails.
func (i MockItem) Has(ctx context.Context) (bool, error) {
	kvStore := i.m.sa(ctx)
	return kvStore.Has(i.m.prefix)
}

// Remove removes the item in the store.
func (i MockItem) Remove(ctx context.Context) error {
	kvStore := i.m.sa(ctx)
	err := kvStore.Delete(i.m.prefix)
	if err != nil {
		return err
	}
	if i.getContainer != nil {
		i.getContainer(ctx).Remove(i.m.prefix)
	}
	return nil
}
