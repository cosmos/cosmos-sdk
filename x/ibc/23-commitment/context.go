package ics23

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// ContextKeyCommitmentKVStore is a singleton type used as the key for the commitment store
type ContextKeyCommitmentKVStore struct{}

// WithStore returns the context updated with the store
func WithStore(ctx sdk.Context, store Store) sdk.Context {
	return ctx.WithValue(ContextKeyCommitmentKVStore{}, store)
}

// GetStore returns the store from the context
func GetStore(ctx sdk.Context) Store {
	return ctx.Value(ContextKeyCommitmentKVStore{}).(Store)
}
