package commitment

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type ContextKeyRemoteKVStore struct{}

func WithStore(ctx sdk.Context, store Store) sdk.Context {
	return ctx.WithValue(ContextKeyRemoteKVStore{}, store)
}

func WithKeyLoggerStore(ctx sdk.Context, store KeyLoggerStore) sdk.Context {
	return ctx.WithValue(ContextKeyRemoteKVStore{}, store)
}

func GetStore(ctx sdk.Context) sdk.KVStore {
	return ctx.Value(ContextKeyRemoteKVStore{}).(sdk.KVStore)
}

func IsKeyLog(ctx sdk.Context) bool {
	_, ok := ctx.Value(ContextKeyRemoteKVStore{}).(KeyLoggerStore)
	return ok
}
