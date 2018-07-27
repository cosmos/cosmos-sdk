package gas

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	params "github.com/cosmos/cosmos-sdk/x/params/store"
)

// Keys for parameter access
const (
	DefaultParamSpace = "GasConfig"
)

// Key generators for parameter store
func KVStoreKey() params.Key        { return params.NewKey("KVStore") }
func TransientStoreKey() params.Key { return params.NewKey("TransientStore") }

// Cached parameter keys
var (
	kvStoreKey        = KVStoreKey()
	transientStoreKey = TransientStoreKey()
)

// NewAnteHandler returns AnteHandler
// that overrides existing gasconfig in the context
// with the gasconfig stored in the paramstore
func NewAnteHandler(store params.Store) sdk.AnteHandler {
	return func(ctx sdk.Context, tx sdk.Tx) (sdk.Context, sdk.Result, bool) {
		if store.Has(ctx, kvStoreKey) {
			var config sdk.GasConfig
			store.Get(ctx, kvStoreKey, &config)
			ctx = ctx.WithKVGasConfig(config)
		}
		if store.Has(ctx, transientStoreKey) {
			var config sdk.GasConfig
			store.Get(ctx, transientStoreKey, &config)
			ctx = ctx.WithTransientGasConfig(config)
		}
		return ctx, sdk.Result{}, false
	}
}
