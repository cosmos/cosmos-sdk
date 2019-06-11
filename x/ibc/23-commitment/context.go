package commitment

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type Context struct {
	sdk.Context
}

type contextKeyRemoteKVStore struct{}

func WithStore(ctx sdk.Context, store Store) Context {
	return Context{ctx.WithValue(contextKeyRemoteKVStore{}, store)}
}

func (ctx Context) Store() Store {
	return ctx.Value(contextKeyRemoteKVStore{}).(Store)
}
