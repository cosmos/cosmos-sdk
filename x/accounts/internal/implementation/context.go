package implementation

import (
	"context"

	"cosmossdk.io/core/store"
	"cosmossdk.io/x/accounts/internal/prefixstore"
)

type contextKey struct{}

type contextValue struct {
	store           store.KVStore   // store is the prefixed store for the account.
	sender          []byte          // sender is the address of the entity invoking the account action.
	whoami          []byte          // whoami is the address of the account being invoked.
	originalContext context.Context // originalContext that was used to build the account context.
}

// MakeAccountContext creates a new account execution context given:
// storeSvc: which fetches the x/accounts module store.
// accountAddr: the address of the account being invoked, which is used to give the
// account a prefixed storage.
// sender: the address of entity invoking the account action.
func MakeAccountContext(ctx context.Context, storeSvc store.KVStoreService, accountAddr, sender []byte) context.Context {
	return context.WithValue(ctx, contextKey{}, contextValue{
		store:           prefixstore.New(storeSvc.OpenKVStore(ctx), accountAddr),
		sender:          sender,
		whoami:          accountAddr,
		originalContext: ctx,
	})
}

// OpenKVStore returns the prefixed store for the account given the context.
func OpenKVStore(ctx context.Context) store.KVStore {
	return ctx.Value(contextKey{}).(contextValue).store
}

// Sender returns the address of the entity invoking the account action.
func Sender(ctx context.Context) []byte {
	return ctx.Value(contextKey{}).(contextValue).sender
}

// Whoami returns the address of the account being invoked.
func Whoami(ctx context.Context) []byte {
	return ctx.Value(contextKey{}).(contextValue).whoami
}
