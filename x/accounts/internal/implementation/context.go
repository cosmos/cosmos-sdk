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
func MakeAccountContext(ctx context.Context, storeSvc store.KVStoreService, accountAddr []byte, sender []byte) context.Context {
	return context.WithValue(ctx, contextKey{}, contextValue{
		store:           prefixstore.New(storeSvc.OpenKVStore(ctx), accountAddr),
		sender:          sender,
		whoami:          accountAddr,
		originalContext: ctx,
	})
}
