package commitment

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// TODO: define Context type which embeds sdk.Context and ensures the existence of RemoteKVStore

type ContextKeyRemoteKVStore struct{}

func WithStore(ctx sdk.Context, store Store) sdk.Context {
	return ctx.WithValue(ContextKeyRemoteKVStore{}, store)
}

func GetStore(ctx sdk.Context) Store {
	return ctx.Value(ContextKeyRemoteKVStore{}).(Store)
}
